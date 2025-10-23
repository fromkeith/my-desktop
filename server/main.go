package main

import (
	_ "fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/gmail_oauth"
	"fromkeith/my-desktop-server/middleware"
	"unicode"
	"unicode/utf8"

	// for swagger
	_ "fromkeith/my-desktop-server/docs"

	"github.com/gin-gonic/gin"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// @title           Desktop Eamil
// @version         1.0
// @description     API Calls
// @termsOfService  http://fromkeith.com

// @host      localhost:5173
// @BasePath  /api
func main() {
	// LowerCamelCase: just lowercase the first rune.
	extra.SetNamingStrategy(func(name string) string {
		if name == "" {
			return name
		}
		r, size := utf8.DecodeRuneInString(name)
		return string(unicode.ToLower(r)) + name[size:]
	})

	gmail_oauth.SetupGoogle()
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.AuthTokenExtract())

	r.GET("/api/gmail/start", gmail_oauth.HandleAuthStart)
	r.GET("/api/gmail/callback", gmail_oauth.HandleCallback)
	r.GET("/api/gmail/inbox", ListInbox)
	// TODO: maybe this is just another list query? they return basically the same thing
	r.GET("/api/gmail/thread/:threadId", ListThread)
	r.GET("/api/gmail/message/:messageId/contents", GetMessageContents)

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	r.Run()
}
