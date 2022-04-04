package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

type SetPixelRequestBody struct {
	X int   `json:"x"`
	Y int   `json:"y"`
	R uint8 `json:"red"`
	G uint8 `json:"green"`
	B uint8 `json:"blue"`
}

// readEnvironmentVariables reads the port and image options from environment variables.
// If environment variables are not provided, default values will be used.
// It returns the server port, image width, and image height
func readEnvironmentVariables() (string, int, int) {
	// read port
	port := os.Getenv("PORT")

	if port == "" {
		defaultPort := "8080"
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

// generateImage generates an image with given width and height
// The image is written to the static folder.
func generateImage(width int, height int) {
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
	f, _ := os.Create("./static/image.png")
	imageWriteErr := png.Encode(f, img)
	if imageWriteErr != nil {
		log.Fatal("Error writing image: ", imageWriteErr)
	}
}

// drawPixelToImage draws a pixel with a color to given x, y position
func drawPixelToImage(pixelX int, pixelY int, newColor color.RGBA) string {
	log.Println("________")
	log.Println("Drawing pixel:")
	log.Println(pixelX)
	log.Println(pixelY)
	log.Println(newColor)

	f, imageFileReadErr := os.Open("./static/image.png")

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
		return "Invalid x coordinate for pixel."
	}

	if pixelY < 0 || pixelY >= height {
		return "Invalid y coordinate for pixel."
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
	newImageFile, _ := os.Create("./static/image.png")
	imageWriteErr := png.Encode(newImageFile, newImg)
	if imageWriteErr != nil {
		log.Fatal("Error writing image: ", imageWriteErr)
	}

	return "Successfully placed pixel."
}

// main will launch the server application
func main() {
	// read environment variables
	port, imageWidth, imageHeight := readEnvironmentVariables()

	// generate blank image
	generateImage(imageWidth, imageHeight)

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

	// use gin logger
	router.Use(gin.Logger())

	// load html from templates folder
	router.LoadHTMLGlob("templates/*.html")

	// load static content from static folder
	router.Static("/static", "static")

	// return index page for default route
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// route to set a pixel
	router.POST("/set_pixel", func(c *gin.Context) {
		var setPixelRequest SetPixelRequestBody

		log.Println("________")
		log.Println("Received request to draw pixel:")
		log.Println(setPixelRequest)

		err := c.BindJSON(&setPixelRequest)

		if err != nil {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Invalid request body.",
			})
			return
		}

		// TODO: check request colors provided are within 0 to 255 range (inclusive)

		// draw pixel to image
		pixelDrawStatus := drawPixelToImage(setPixelRequest.X, setPixelRequest.Y, color.RGBA{R: uint8(setPixelRequest.R), G: uint8(setPixelRequest.G), B: uint8(setPixelRequest.B), A: 0xFF})

		c.JSON(200, gin.H{
			"status":  "success",
			"message": pixelDrawStatus,
		})
	})

	// start the server
	serverStartErr := router.Run(":" + port)

	if serverStartErr != nil {
		log.Fatal("Error launching server: ", serverStartErr)
	}
}
