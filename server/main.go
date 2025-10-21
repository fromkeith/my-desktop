package main

import (
	"github.com/gin-gonic/gin"

	_ "github.com/joho/godotenv/autoload"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {

	open()
	setupGoogle()
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(AuthTokenExtract())

	r.GET("/api/gmail/start", handleAuthStart)
	r.GET("/api/gmail/callback", handleCallback)

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	r.Run()
}
