package messages

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SyncCheckpoint struct {
	Id        string `json:"id"`
	UpdatedAt uint64 `json:"updatedAt"`
}

type PullMessagesResponse struct {
	Messages   []data.GmailEntry `json:"messages"`
	Checkpoint SyncCheckpoint    `json:"checkpoint"`
}

func PullMessage(r *gin.Context) {
	messageId := r.Query("messageId")
	if messageId == "" {
		r.JSON(400, gin.H{"error": "messageId is required"})
		return
	}
	lastId := toDocumentIdRequest(r, messageId)
	updatedAtStr := r.Query("updatedAt")
	updatedAt, _ := strconv.ParseUint(updatedAtStr, 10, 64)

	batchSizeStr := r.Query("batchSize")
	batchSize, _ := strconv.ParseInt(batchSizeStr, 10, 64)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	opts := options.Find().SetSort(bson.D{{"updatedAt", 1}, {"_id", 1}}).SetLimit(batchSize)
	cursor, err := globals.DocDb().Collection("Messages").Find(
		r,
		bson.D{
			{
				"$or", []bson.D{
					bson.D{{"updatedAt", bson.D{{"$gt", updatedAt}}}},
					bson.D{{"updatedAt", bson.D{{"$eq", updatedAt}, {"_id", bson.D{{"$gt", lastId}}}}}},
				},
			},
		},
		opts,
	)
	if err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(r)

	var messages []data.GmailEntry
	if err := cursor.All(r, &messages); err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var nextId string
	var nextUpdatedAt uint64
	if len(messages) > 0 {
		last := messages[len(messages)-1]
		nextId = toDocumentIdRequest(r, last.MessageId)
		nextUpdatedAt = last.HistoryId
	} else {
		nextId = lastId
		nextUpdatedAt = updatedAt
	}

	r.JSON(200, PullMessagesResponse{
		Messages:   messages,
		Checkpoint: SyncCheckpoint{Id: nextId, UpdatedAt: nextUpdatedAt},
	})
}
