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
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	// Encode as PNG.
	f, _ := os.Create("./static/image.png")
	imageWriteErr := png.Encode(f, img)
	if imageWriteErr != nil {
		log.Fatal("Error writing image: ", imageWriteErr)
	}
}

// main will launch the server application
func main() {
	// generate sample image
	port, imageWidth, imageHeight := readEnvironmentVariables()
	generateImage(imageWidth, imageHeight)

	// create instance of gin router
	router := gin.New()

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

	// start the server
	serverStartErr := router.Run(":" + port)

	if serverStartErr != nil {
		log.Fatal("Error launching server: ", serverStartErr)
	}
}
