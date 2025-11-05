package main

import (
	"context"
	"database/sql"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
		log.Println("doc failed", err)
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	res := make([]data.GmailEntry, 0, 100)
	if err = cursor.All(r, &res); err != nil {
		log.Println("doc failed", err)
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get docs"})
		return
	}
	if len(res) == 0 {
		client, _ := client.GmailClientFor(r, true)
		go client.Bootstrap(context.Background())
		r.JSON(http.StatusOK, []data.GmailEntry{})
		return
	}

	r.JSON(http.StatusOK, res)
}

func unmarshalEmailEntry(rows *sql.Rows) *data.GmailEntry {
	var entry data.GmailEntry
	var labelsJson, headersJson, senderJson, receiverJson, replyToJson, additionalReceiversJson []byte
	err := rows.Scan(
		&entry.UserId,
		&entry.MessageId,
		&entry.ThreadId,
		&labelsJson,
		&entry.Subject,
		&entry.Snippet,
		&entry.HistoryId,
		&entry.InternalDate,
		&headersJson,
		&senderJson,
		&receiverJson,
		&entry.ReceivedAt,
		&replyToJson,
		&additionalReceiversJson,
	)
	if err != nil {
		log.Println("failed to unmarshal gmail entry", err)
		return nil
	}
	json.Unmarshal((labelsJson), &entry.Labels)
	json.Unmarshal((headersJson), &entry.Headers)
	json.Unmarshal((senderJson), &entry.Sender)
	json.Unmarshal((receiverJson), &entry.Receiver)
	json.Unmarshal((replyToJson), &entry.ReplyTo)
	json.Unmarshal((additionalReceiversJson), &entry.AdditionalReceivers)
	return &entry
}

func bootstrap(r *gin.Context) {
	log.Println("Need to bookstrap")
	client, err := client.GmailClientFor(r, true)
	if err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
		return
	}
	go client.Bootstrap(context.Background())
	// r.JSON(http.StatusOK, make([]string, 0))
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
		log.Println("doc failed", err)
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	res := make([]data.GmailEntry, 0, 100)
	if err = cursor.All(r, &res); err != nil {
		log.Println("doc failed", err)
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
		log.Println("doc read failed", err)
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
			log.Println(err)
			r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
			return
		}
		// wait, then pull from db
		client.FetchOneMessage(r, messageId)
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
		log.Println("doc read failed", err, "docId", data.ToDocumentId(r.GetString("accountId"), messageId))
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to query collection"})
		return
	}
	r.JSON(http.StatusOK, entry)
}
