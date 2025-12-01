package aggregate

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TagInfo struct {
	Tag          string    `validate:"required" json:"tag" bson:"tag"`
	MessageCount int64     `validate:"required" json:"messageCount" bson:"messageCount"`
	UpdatedAt    time.Time `validate:"required" json:"updatedAt" bson:"updatedAt"`
} // @name TagInfo
type SyncCheckpointTag struct {
	Tag       string `validate:"required" json:"tag" bson:"tag"`
	UpdatedAt string `validate:"required" json:"updatedAt" bson:"updatedAt"`
} // @name CheckpointTag

type PullTagsResponse struct {
	Tags       []TagInfo         `validate:"required" json:"tags"`
	Checkpoint SyncCheckpointTag `validate:"required" json:"checkpoint"`
} // @name PullTagsResponse

// PullTags godoc
// @Summary      Get summary of the tags in this account
// @Description  Sync endpoint to pull all changes to tags for this account.
// @Tags         email
// @Produce      json
// @Param        tag query string true "Tag"
// @Param        updatedAt query string true "Last updated time"
// @Param        limit query int true "Batch size"
// @Success      200  {object}  PullTagsResponse
// @Router       /messages/aggregate/pullTags [get]
func PullTags(r *gin.Context) {
	accountId := r.GetString("accountId")
	tag := r.Query("tag")
	lastId := r.GetString("accountId") + ";" + tag
	updatedAtStr := r.Query("updatedAt")
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	batchSizeStr := r.Query("limit")
	batchSize, _ := strconv.ParseInt(batchSizeStr, 10, 64)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	opts := options.Find().SetSort(bson.D{{"updatedAt", 1}, {"_id", 1}}).SetLimit(batchSize)
	cursor, err := globals.DocDb().Collection("AccountTags").Find(
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

	var tags []TagInfo
	if err := cursor.All(r, &tags); err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var nextId string
	var nextUpdatedAt string
	if len(tags) > 0 {
		last := tags[len(tags)-1]
		nextId = last.Tag
		nextUpdatedAt = last.UpdatedAt.Format(time.RFC3339Nano)
	} else {
		nextId = tag
		nextUpdatedAt = updatedAtStr
	}
	if len(tags) < int(batchSize) {
		go client.CheckForGmailsUpdates(accountId)
	}

	r.JSON(200, PullTagsResponse{
		Tags:       tags,
		Checkpoint: SyncCheckpointTag{Tag: nextId, UpdatedAt: nextUpdatedAt},
	})
}
