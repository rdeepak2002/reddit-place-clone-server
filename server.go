package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	// get port form environment or default to 8080
	port := os.Getenv("PORT")

	if port == "" {
		defaultPort := "8080"
		port = defaultPort
		fmt.Printf("Manually setting port to %s\n", defaultPort)
	}

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
	router.Run(":" + port)
}
