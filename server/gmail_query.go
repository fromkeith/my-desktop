package main

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ListInbox godoc
// @Summary      List the inbox
// @Description  List the user's email inbox
// @Tags         email
// @Produce      json
// @Success      200  {object}  []data.GmailEntry
// @Router       /gmail/inbox [get]
func ListInbox(r *gin.Context) {

	filter := bson.D{
		{"accountId", r.GetString("accountId")},
	}
	sort := bson.D{
		{"updatedAt", -1},
	}
	opts := options.Find().SetSort(sort).SetLimit(100)

	cursor, err := globals.DocDb().Collection("Messages").
		Find(
			r,
			filter,
			opts,
		)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to list Messages from mongo")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	res := make([]data.GmailEntry, 0, 100)
	if err = cursor.All(r, &res); err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to list all cursor for Messages")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get docs"})
		return
	}
	if len(res) == 0 {
		client, err := client.GmailClientFor(r, true)
		if err != nil {
			log.Error().
				Ctx(r).
				Err(err).
				Msg("failed to get GmailClient when wanting ondemand bootstrap")
			r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
			return
		}
		go client.Bootstrap(context.Background())
		r.JSON(http.StatusOK, []data.GmailEntry{})
		return
	}

	r.JSON(http.StatusOK, res)
}

func bootstrap(r *gin.Context) {
	client, err := client.GmailClientFor(r, true)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to get GmailClient when wanting ondemand bootstrap")
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
		return
	}
	go client.Bootstrap(context.Background())
}

// ListThread godoc
// @Summary      List all messages in a thread
// @Tags         email
// @Produce      json
// @Success      200  {object}  []data.GmailEntry
// @Router       /gmail/thread/{threadId} [get]
func ListThread(r *gin.Context) {
	threadId := r.Param("threadId")

	filter := bson.D{
		{"accountId", r.GetString("accountId")},
		{"threadId", threadId},
	}
	sort := bson.D{
		{"updatedAt", -1},
	}
	opts := options.Find().SetSort(sort).SetLimit(100)

	cursor, err := globals.DocDb().Collection("Messages").
		Find(
			r,
			filter,
			opts,
		)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Str("threadId", threadId).
			Msg("failed to get thread for Messages from mongo")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	res := make([]data.GmailEntry, 0, 100)
	if err = cursor.All(r, &res); err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Str("threadId", threadId).
			Msg("failed to list all cursor for Messages Thread")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get docs"})
		return
	}
	if len(res) == 0 {
		r.JSON(http.StatusOK, []data.GmailEntry{})
		return
	}

	r.JSON(http.StatusOK, res)

}

// GetMessage godoc
// @Summary      Get the basic information about a message
// @Tags         email
// @Produce      json
// @Success      200  {object}  data.GmailEntry
// @Router       /gmail/message/:messageId [get]
func GetMessage(r *gin.Context) {
	messageId := r.Param("messageId")

	filter := bson.D{
		{"_id", data.ToDocumentId(r.GetString("accountId"), messageId)},
	}
	opts := options.FindOne()

	result := globals.DocDb().Collection("Messages").
		FindOne(
			r,
			filter,
			opts,
		)
	var entry data.GmailEntry
	if err := result.Decode(&entry); err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Str("messageId", messageId).
			Msg("failed to get Message from mongo")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	r.JSON(http.StatusOK, entry)

}

// GetMessageContents godoc
// @Summary      Get the contents of a message
// @Tags         email
// @Produce      json
// @Success      200  {object}  data.GmailEntryBody
// @Router       /gmail/message/:messageId/contents [get]
func GetMessageContents(r *gin.Context) {
	messageId := r.Param("messageId")
	forceRefresh := r.Query("force")
	if forceRefresh == "1" {
		client, err := client.GmailClientFor(r, true)
		if err != nil {
			log.Error().
				Ctx(r).
				Err(err).
				Str("messageId", messageId).
				Msg("failed to get client for forced body load")
			r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
			return
		}
		// wait, then pull from db
		entry, body, err := client.FetchGmailEntry(r, messageId)
		if err != nil {
			log.Error().
				Ctx(r).
				Err(err).
				Str("messageId", messageId).
				Msg("failed to fetch gmail entry")
			r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail entry"})
			return
		}
		data.WriteGmailEntry(*entry)
		data.WriteGmailEntryBody(*body)
		r.JSON(http.StatusOK, entry)
		return
	}

	filter := bson.D{
		{"_id", data.ToDocumentId(r.GetString("accountId"), messageId)},
	}
	opts := options.FindOne()

	result := globals.DocDb().Collection("MessageBodies").
		FindOne(
			r,
			filter,
			opts,
		)
	var entry data.GmailEntryBody
	if err := result.Decode(&entry); err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Str("messageId", messageId).
			Msg("failed to decode gmail body")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	r.JSON(http.StatusOK, entry)
}
