package main

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	_ "fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/messages"
	"fromkeith/my-desktop-server/middleware"
	"fromkeith/my-desktop-server/people"

	// for swagger
	_ "fromkeith/my-desktop-server/docs"

	"github.com/gin-gonic/gin"
)

// @title           Desktop Eamil
// @version         1.0
// @description     API Calls
// @termsOfService  http://fromkeith.com

// @host      localhost:5173
// @BasePath  /api
func main() {
	globals.SetupJsonEncoding()

	defer globals.CloseAll()

	client.SetupGoogle()
	bkg := context.Background()
	go data.StartWriter(bkg)
	go data.StartBodyWriter(bkg)
	go client.StartBackgroundRefresher(bkg)

	defer bkg.Done()

	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.AuthTokenExtract())

	r.GET("/api/gmail/start", client.HandleAuthStart)
	r.GET("/api/gmail/callback", client.HandleCallback)
	r.GET("/api/gmail/inbox", ListInbox)
	// TODO: maybe this is just another list query? they return basically the same thing
	r.GET("/api/gmail/thread/:threadId", ListThread)
	r.GET("/api/gmail/message/:messageId/contents", GetMessageContents)
	r.GET("/api/gmail/message/:messageId", GetMessage)

	r.GET("/api/messages/pull", messages.PullMessage)
	r.GET("/api/messages/push", messages.PushMessage)
	r.GET("/api/messages/pullStream", middleware.StreamHeaders(), messages.PullStream)

	r.GET("/api/people/sync", people.SyncPeople)
	r.GET("/api/people/pull", people.PullPeople)

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	r.Run()
}
