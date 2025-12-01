package messages

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SyncCheckpoint struct {
	MessageId string `json:"messageId"`
	UpdatedAt string `json:"updatedAt"`
} // @name CheckpointMessages

type PullMessagesResponse struct {
	Messages   []data.GmailEntry `json:"messages"`
	Checkpoint SyncCheckpoint    `json:"checkpoint"`
} // @name PullMessagesResponse

// PullMessage godoc
// @Summary      Get Messages
// @Description  Sync endpoint to pull all changes to messages for this account.
// @Tags         email
// @Produce      json
// @Param        messageId query string true "messageid"
// @Param        updatedAt query string true "Last updated time"
// @Param        limit query int true "Batch size"
// @Success      200  {object}  PullMessagesResponse
// @Router       /messages/pull [get]
func PullMessage(r *gin.Context) {
	accountId := r.GetString("accountId")
	messageId := r.Query("messageId")
	lastId := toDocumentIdRequest(r, messageId)
	updatedAtStr := r.Query("updatedAt")
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	batchSizeStr := r.Query("limit")
	batchSize, _ := strconv.ParseInt(batchSizeStr, 10, 64)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	opts := options.Find().SetSort(bson.D{{"updatedAt", 1}, {"_id", 1}}).SetLimit(batchSize)
	cursor, err := globals.DocDb().Collection("Messages").Find(
		r,
		bson.M{
			"$or": []bson.M{
				bson.M{"updatedAt": bson.M{"$gt": updatedAt}},
				bson.M{
					"updatedAt": updatedAt,
					"_id":       bson.M{"$gt": lastId},
				},
			},
			"accountId": accountId,
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
	var nextUpdatedAt string
	if len(messages) > 0 {
		last := messages[len(messages)-1]
		nextId = last.MessageId
		nextUpdatedAt = last.UpdatedAt.Format(time.RFC3339Nano)
	} else {
		nextId = messageId
		nextUpdatedAt = updatedAtStr
	}
	if len(messages) < int(batchSize) {
		go client.CheckForGmailsUpdates(accountId)
	}

	r.JSON(200, PullMessagesResponse{
		Messages:   messages,
		Checkpoint: SyncCheckpoint{MessageId: nextId, UpdatedAt: nextUpdatedAt},
	})
}
