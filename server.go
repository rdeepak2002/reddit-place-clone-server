package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/heroku/x/hmetrics/onload"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

import _ "github.com/joho/godotenv/autoload"

// SetPixelRequestBody JSON format for set pixel request
type SetPixelRequestBody struct {
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Red   uint8 `json:"red"`
	Green uint8 `json:"green"`
	Blue  uint8 `json:"blue"`
}

var ctx = context.Background()
var rdb *redis.Client

func base64ToPng(data string) {
	base64PngPrefix := "data:image/png;base64,"
	base64JpegPrefix := "data:image/jpeg;base64,"

	if strings.HasPrefix(data, base64PngPrefix) {
		log.Println("Converting base 64 png to file")
		data = data[len(base64PngPrefix):]
	} else if strings.HasPrefix(data, base64JpegPrefix) {
		log.Println("Converting base 64 jpeg to file")
		data = data[len(base64JpegPrefix):]
	}

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, formatString, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	bounds := m.Bounds()
	fmt.Println(bounds, formatString)

	//Encode from image format to writer
	pngFilename := "./static/image.png"
	f, err := os.OpenFile(pngFilename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal("Error opening png file for base 64 dump", err)
		return
	}

	err = png.Encode(f, m)
	if err != nil {
		log.Fatal("Error saving base 64 to png", err)
		return
	}

	log.Println("Png file", pngFilename, "created")
}

func imageToBase64String(filename string) string {
	// Read the entire file into a byte slice
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	var base64Encoding string

	// Determine the content type of the image file
	mimeType := http.DetectContentType(bytes)

	// Prepend the appropriate URI scheme header depending
	// on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	}

	// Append the base64 encoded output
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	log.Println("Converted image to base 64")

	// Print the full base64 representation of the image
	return base64Encoding
}

func setupRedisClient() {
	// address usually in form redis-xxx.com:#####
	redisAddress := os.Getenv("REDIS_ADDRESS")

	// password usually in form of a really long string of characters
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// create redis client connection
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword, // no password set
		DB:       0,             // use default DB
	})

	// test whether redis connection works
	err := rdb.Set(ctx, "testKey", "testValue", 0).Err()
	if err != nil {
		log.Fatal("Issue with Redis connection when testing setting key value pair")
	}

	_, err = rdb.Get(ctx, "key").Result()
	if err != nil {
		log.Fatal("Issue with Redis connection when getting key1")
	}
}

func saveCurrentImageToRedis() {
	imagePath := "./static/image.png"

	// save base 64 encoded image to redis
	base64Image := imageToBase64String(imagePath)

	redisImageSaveErr := rdb.Set(ctx, "image", base64Image, 0).Err()
	if redisImageSaveErr != nil {
		log.Fatal("Error saving base64 encoded image to Redis")
	}

	log.Println("Saved base64 image to Redis")
}

func verifyGoogleToken(googleAuthToken string) (string, string) {
	url := "https://oauth2.googleapis.com/tokeninfo?id_token=" + googleAuthToken
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "error", "error forming request"
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "error", "error doing request"
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "error", "error reading response body"
	}

	var responseMap map[string]interface{}

	unmarshalError := json.Unmarshal(body, &responseMap)

	if unmarshalError != nil {
		fmt.Println(unmarshalError)
		return "error", "unable to parse response from Google auth"
	}

	_, ok := responseMap["email"]

	if !ok {
		return "error", "unable to find email in Google token response"
	}

	audValue, ok := responseMap["aud"]

	if !ok {
		return "error", "unable to find aud in Google token response"
	}

	if audValue != os.Getenv("GOOGLE_AUTH_CLIENT_ID") {
		return "error", "aud does not match client id"
	}

	log.Println("Google auth check succeeded")

	return "success", string(body)
}

// readEnvironmentVariables reads the port and image options from environment variables.
// If environment variables are not provided, default values will be used.
// It returns the server port, image width, and image height
func readEnvironmentVariables() (string, int, int) {
	// read port
	port := os.Getenv("PORT")

	if port == "" {
		defaultPort := "3000"
		port = defaultPort
	}

	// read width
	defaultWidth := "32"
	width := os.Getenv("IMAGE_WIDTH")
	if width == "" {
		width = defaultWidth
	}
	widthInt, widthParseErr := strconv.Atoi(width)
	if widthParseErr != nil {
		log.Fatal("Error reading width from environment variables: ", widthParseErr)
	}

	// read height
	defaultHeight := "32"
	height := os.Getenv("IMAGE_HEIGHT")
	if height == "" {
		height = defaultHeight
	}
	heightInt, heightParseErr := strconv.Atoi(height)
	if heightParseErr != nil {
		log.Fatal("Error reading height from environment variables: ", heightParseErr)
	}

	return port, widthInt, heightInt
}

// generateBlankImage generates an image with given width and height
// The image is written to the static folder.
func generateBlankImage(width int, height int) {
	imagePath := "./static/image.png"

	currentBase64Image, err := rdb.Get(ctx, "image").Result()

	if err == redis.Nil {
		// image not present in Redis, so generate a blank one
		upLeft := image.Point{}
		lowRight := image.Point{X: width, Y: height}

		img := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})

		// Set color for each pixel.
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				img.Set(x, y, color.White)
			}
		}

		// Encode as PNG.
		f, _ := os.Create(imagePath)
		imageWriteErr := png.Encode(f, img)
		if imageWriteErr != nil {
			log.Fatal("Error writing image: ", imageWriteErr)
		}
	} else if err != nil {
		log.Fatal("Error getting image from Redis")
	} else {
		log.Println("Current base64 image in Redis: ", currentBase64Image)
		// image present in redis so read it from there and generate it
		base64ToPng(currentBase64Image)
	}
}

// drawPixelToImage draws a pixel with a color to given x, y position
func drawPixelToImage(pixelX int, pixelY int, newColor color.RGBA) (string, string) {
	log.Println("________")
	log.Println("Drawing pixel:")
	log.Println(pixelX)
	log.Println(pixelY)
	log.Println(newColor)

	imagePath := "./static/image.png"

	f, imageFileReadErr := os.Open(imagePath)

	if imageFileReadErr != nil {
		log.Fatal("Error reading image file: ", imageFileReadErr)
	}

	defer func(f *os.File) {
		fileCloseErr := f.Close()
		if fileCloseErr != nil {
			log.Fatal("Error closing file: ", fileCloseErr)
		}
	}(f)

	exisitingImage, _, imageDecodeErr := image.Decode(f)

	if imageDecodeErr != nil {
		log.Fatal("Error decoding image: ", imageDecodeErr)
	}

	width := exisitingImage.Bounds().Size().X
	height := exisitingImage.Bounds().Size().Y

	if pixelX < 0 || pixelX >= width {
		return "Invalid x coordinate for pixel.", "error"
	}

	if pixelY < 0 || pixelY >= height {
		return "Invalid y coordinate for pixel.", "error"
	}

	// generate new image file
	upLeft := image.Point{}
	lowRight := image.Point{X: width, Y: height}

	newImg := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x == pixelX && y == pixelY {
				newImg.Set(x, y, newColor)
			} else {
				newImg.Set(x, y, exisitingImage.At(x, y))
			}
		}
	}

	// Encode as PNG.
	newImageFile, _ := os.Create(imagePath)
	imageWriteErr := png.Encode(newImageFile, newImg)
	if imageWriteErr != nil {
		log.Fatal("Error writing image: ", imageWriteErr)
	}

	// save current image as base64 string to Redis
	saveCurrentImageToRedis()

	return "Successfully placed pixel.", "success"
}

// main will launch the server application
func main() {
	// read environment variables
	port, imageWidth, imageHeight := readEnvironmentVariables()

	// setup the Redis client
	setupRedisClient()

	// generate blank image
	generateBlankImage(imageWidth, imageHeight)

	// create instance of gin router
	router := gin.New()

	// TODO: add rate limiting (https://github.com/ulule/limiter-examples/blob/master/gin/main.go)

	// do not cache
	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Writer.Header().Set("Cache-Control", "no-cache, must-revalidate")
			c.Writer.Header().Set("Pragma", "no-cache")
			c.Writer.Header().Set("Expires", "Sat, 26 Jul 1997 05:00:00 GMT")
		}
	}(),
	)

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowHeaders = []string{"Authorization", "Content-Type"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	// use gin logger
	router.Use(gin.Logger())

	// load static content from static folder
	router.Static("/static", "static")

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("./client/public", true)))

	// return index page for default route
	//router.GET("/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "index.html", nil)
	//})

	// route to set a pixel
	router.POST("/set_pixel", func(c *gin.Context) {
		var setPixelRequest SetPixelRequestBody

		log.Println("________")
		log.Println("Received request to draw pixel:")
		log.Println(setPixelRequest)

		// get google auth token string
		token := c.Request.Header["Authorization"]

		if len(token) < 1 || !strings.HasPrefix(token[0], "Bearer ") {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Invalid request. Invalid auth token provided.",
			})
			return
		}

		tokenStr := strings.TrimSpace(token[0][len("Bearer "):])

		// verify the user's google token
		tokenVerificationStatus, tokenVerificationResult := verifyGoogleToken(tokenStr)

		if tokenVerificationStatus == "error" {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": tokenVerificationResult,
			})
			return
		}

		// convert user data to map of strings
		userDataMap := map[string]string{}

		unmarshalError := json.Unmarshal([]byte(tokenVerificationResult), &userDataMap)
		if unmarshalError != nil {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "error unmarshalling Google user account data",
			})
			return
		}

		// get email of user
		emailValue, ok := userDataMap["email"]

		if !ok {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Google user does not have email",
			})
			return
		}

		// rate limit user to only place 1 pixel every 5 minutes
		rateLimitExpiresAtTime, redisFindUserErr := rdb.Get(ctx, emailValue).Result()

		// get timestamp to expire at
		expirationInSeconds := 300
		expirationTimestamp := time.Now().Unix() + int64(expirationInSeconds)

		if redisFindUserErr == redis.Nil {
			// user rate limit not present so create new one
			expirationDuration, durationCreationException := time.ParseDuration(strconv.FormatInt(int64(expirationInSeconds), 10) + "s")

			if durationCreationException != nil {
				log.Fatal("Error creating duration")
			}

			rateLimitObjCreationErr := rdb.Set(ctx, emailValue, strconv.FormatInt(expirationTimestamp, 10), expirationDuration).Err()

			if rateLimitObjCreationErr != nil {
				log.Fatal("Issue with Redis connection when creating rate limit for pixel placing")
			}

			log.Println("Created Redis rate limit object")
		} else if redisFindUserErr != nil {
			// error
			log.Fatal("Error getting user from Redis")
		} else {
			log.Println("Using existing Redis rate limit object: ", rateLimitExpiresAtTime)

			// user already exists in Redis so return their rate limit time
			c.JSON(200, gin.H{
				"status":  "ratelimit",
				"message": rateLimitExpiresAtTime,
			})
			return
		}

		// bind set pixel request to json struct
		err := c.BindJSON(&setPixelRequest)

		if err != nil {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Invalid request. Please verify your inputs are valid and within the canvas.",
			})
			return
		}

		// draw pixel to image
		pixelDrawStatusMessage, pixelDrawStatus := drawPixelToImage(setPixelRequest.X, setPixelRequest.Y,
			color.RGBA{R: setPixelRequest.Red, G: setPixelRequest.Green, B: setPixelRequest.Blue, A: 0xFF})

		c.JSON(200, gin.H{
			"status":  pixelDrawStatus,
			"message": pixelDrawStatusMessage,
		})
	})

	// start the server
	serverStartErr := router.Run(":" + port)

	if serverStartErr != nil {
		log.Fatal("Error launching server: ", serverStartErr)
	}
}
