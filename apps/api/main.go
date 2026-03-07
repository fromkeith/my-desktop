package main

import (
	"context"
	api_client "fromkeith/my-desktop-server/apps/api/gmail/client"
	"fromkeith/my-desktop-server/apps/api/messages"
	"fromkeith/my-desktop-server/apps/api/messages/aggregate"
	"fromkeith/my-desktop-server/apps/api/middleware"
	"fromkeith/my-desktop-server/apps/api/people"
	"fromkeith/my-desktop-server/apps/api/threads"
	"fromkeith/my-desktop-server/shared/globals"
	_ "fromkeith/my-desktop-server/shared/globals"
	"fromkeith/my-desktop-server/shared/gmail/client"
	"fromkeith/my-desktop-server/shared/gmail/data"

	"github.com/rs/zerolog/log"

	// for swagger
	_ "fromkeith/my-desktop-server/apps/api/docs"

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

	bkg := context.Background()

	go data.StartWriter(bkg)
	go data.StartBodyWriter(bkg)
	go client.StartBackgroundRefresher(bkg)

	defer bkg.Done()

	// Create a Gin router with default middleware (logger and recovery)
	globals.HookGin()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Debug().
			Str("method", httpMethod).
			Str("path", absolutePath).
			Str("handler", handlerName).
			Int("handlers", nuHandlers).
			Msg("endpoint")
	}
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(middleware.RequestId())
	r.Use(middleware.AuthTokenExtract())

	r.GET("/api/gmail/start", api_client.HandleAuthStart)
	r.GET("/api/gmail/callback", api_client.HandleCallback)
	r.GET("/api/gmail/inbox", ListInbox)
	// TODO: maybe this is just another list query? they return basically the same thing
	r.GET("/api/gmail/thread/:threadId", ListThread)
	r.GET("/api/gmail/message/:messageId/contents", GetMessageContents)
	r.GET("/api/gmail/message/:messageId", GetMessage)

	r.GET("/api/messages/pull", messages.PullMessage)
	r.POST("/api/messages/push", messages.PushMessage)
	r.GET("/api/messages/pullStream", middleware.StreamHeaders(), messages.PullStream)
	r.GET("/api/messages/categories", aggregate.CountCategories)
	r.GET("/api/messages/aggregate/pullCategories", aggregate.PullCategories)
	r.GET("/api/messages/aggregate/pullTags", aggregate.PullTags)
	// THIS IS A DEBUG ENDPOINT
	r.POST("/api/messages/:messageId/redo/:userId", messages.ReInjest)
	r.POST("/api/messages/sync", messages.ForceSyncMessages)
	// END DEBUG ENDPOINTS

	r.GET("/api/threads/pull", threads.PullThread)
	r.GET("/api/threads/pullStream", middleware.StreamHeaders(), threads.PullStream)

	r.GET("/api/people/sync", people.ForceSyncPeople)
	r.GET("/api/people/pull", people.PullPeople)

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	r.Run()
}
