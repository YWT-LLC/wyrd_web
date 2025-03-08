package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	initDB()
	initOAuth2()

	router := gin.Default()

	// OAuth2 token endpoint
	router.POST("/token", TokenHandler)

	// Start server
	log.Println("Starting server on :8080")
	router.Run(":8080")
}
