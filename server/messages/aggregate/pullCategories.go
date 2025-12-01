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

type CategoryInfo struct {
	Category     string    `json:"category" bson:"category"`
	MessageCount int64     `json:"messageCount" bson:"messageCount"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
} // @name CategoryInfo
type SyncCheckpointCategory struct {
	Category  string `json:"category" bson:"category"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
} // @name CheckpointCategory

type PullCategoriesResponse struct {
	Categories []CategoryInfo         `json:"categories"`
	Checkpoint SyncCheckpointCategory `json:"checkpoint"`
} // @name PullCategoriesResponse

// PullCategories godoc
// @Summary      Get summary of the categories in this account
// @Description  Sync endpoint to pull all changes to categories for this account.
// @Tags         email
// @Produce      json
// @Param        category query string true "category"
// @Param        updatedAt query string true "Last updated time"
// @Param        limit query int true "Batch size"
// @Success      200  {object}  PullCategoriesResponse
// @Router       /messages/aggregate/pullCategories [get]
func PullCategories(r *gin.Context) {
	accountId := r.GetString("accountId")
	category := r.Query("category")
	lastId := r.GetString("accountId") + ";" + category
	updatedAtStr := r.Query("updatedAt")
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	batchSizeStr := r.Query("limit")
	batchSize, _ := strconv.ParseInt(batchSizeStr, 10, 64)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	opts := options.Find().SetSort(bson.D{{"updatedAt", 1}, {"_id", 1}}).SetLimit(batchSize)
	cursor, err := globals.DocDb().Collection("AccountCategories").Find(
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

	var cats []CategoryInfo
	if err := cursor.All(r, &cats); err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var nextId string
	var nextUpdatedAt string
	if len(cats) > 0 {
		last := cats[len(cats)-1]
		nextId = last.Category
		nextUpdatedAt = last.UpdatedAt.Format(time.RFC3339Nano)
	} else {
		nextId = category
		nextUpdatedAt = updatedAtStr
	}
	if len(cats) < int(batchSize) {
		go client.CheckForGmailsUpdates(accountId)
	}

	r.JSON(200, PullCategoriesResponse{
		Categories: cats,
		Checkpoint: SyncCheckpointCategory{Category: nextId, UpdatedAt: nextUpdatedAt},
	})
}
