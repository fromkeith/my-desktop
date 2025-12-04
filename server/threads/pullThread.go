package threads

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

type MessageBasic struct {
	MessageId    string          `validate:"required" json:"messageId" bson:"messageId"`
	InternalDate int64           `validate:"required" json:"internalDate" bson:"internalDate"`
	Sender       data.PersonInfo `json:"sender" bson:"sender"`
	Subject      string          `json:"subject" bson:"subject"`
	Snippet      string          `json:"snippet" bson:"snippet"`
} // @name MessageBasic

type ThreadEntry struct {
	Messages               []MessageBasic `validate:"required" json:"messages" bson:"messages"`
	UpdatedAt              time.Time      `validate:"required" json:"updatedAt" bson:"updatedAt"`
	ThreadId               string         `validate:"required" json:"threadId" bson:"threadId"`
	MostRecentInternalDate int64          `validate:"required" json:"mostRecentInternalDate" bson:"mostRecentInternalDate"`
	Labels                 []string       `validate:"required" json:"labels" bson:"labels"`
	Categories             []string       `validate:"required" json:"categories" bson:"categories"`
	Tags                   []string       `validate:"required" json:"tags" bson:"tags"`
} // @name Thread

type SyncCheckpoint struct {
	ThreadId  string `validate:"required" json:"threadId"`
	UpdatedAt string `validate:"required" json:"updatedAt"`
} // @name CheckpointThreads

type PullThreadResponse struct {
	Threads    []ThreadEntry  `validate:"required" json:"threads"`
	Checkpoint SyncCheckpoint `validate:"required" json:"checkpoint"`
} // @name PullThreadResponse

// PullThread godoc
// @Summary      Get Threads
// @Description  Sync endpoint to pull all changes to threads for this account.
// @Tags         email
// @Produce      json
// @Param        threadId query string true "threadId"
// @Param        updatedAt query string true "Last updated time"
// @Param        limit query int true "Batch size"
// @Success      200  {object}  PullThreadResponse
// @Router       /threads/pull [get]
func PullThread(r *gin.Context) {
	accountId := r.GetString("accountId")
	threadId := r.Query("threadId")
	lastId := accountId + ";" + threadId
	updatedAtStr := r.Query("updatedAt")
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	batchSizeStr := r.Query("limit")
	batchSize, _ := strconv.ParseInt(batchSizeStr, 10, 64)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	opts := options.Find().SetSort(bson.D{{"updatedAt", 1}, {"_id", 1}}).SetLimit(batchSize)
	cursor, err := globals.DocDb().Collection("MessageThreads").Find(
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

	var messages []ThreadEntry
	if err := cursor.All(r, &messages); err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var nextId string
	var nextUpdatedAt string
	if len(messages) > 0 {
		last := messages[len(messages)-1]
		nextId = last.ThreadId
		nextUpdatedAt = last.UpdatedAt.Format(time.RFC3339Nano)
	} else {
		nextId = threadId
		nextUpdatedAt = updatedAtStr
	}
	if len(messages) < int(batchSize) {
		go client.CheckForGmailsUpdates(accountId)
	}

	r.JSON(200, PullThreadResponse{
		Threads:    messages,
		Checkpoint: SyncCheckpoint{ThreadId: nextId, UpdatedAt: nextUpdatedAt},
	})
}
