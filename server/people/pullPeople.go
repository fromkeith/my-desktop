package people

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
	PersonId  string `json:"personId"`
	UpdatedAt string `json:"updatedAt"`
}

type PullPeopleResponse struct {
	People     []data.GooglePerson `json:"people"`
	Checkpoint SyncCheckpoint      `json:"checkpoint"`
}

// PullPeople godoc
// @Summary      Pull the people database
// @Description  Pulls the people database to be local
// @Tags         people
// @Produce      json
// @Success      200  {object}  PullPeopleResponse
// @Router       /people/pull [get]
func PullPeople(r *gin.Context) {
	accountId := r.GetString("accountId")
	personId := r.Query("personId")
	lastId := toDocumentIdRequest(r, personId)
	updatedAtStr := r.Query("updatedAt")
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	batchSizeStr := r.Query("limit")
	batchSize, _ := strconv.ParseInt(batchSizeStr, 10, 64)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	opts := options.Find().SetSort(bson.D{{"updatedAt", 1}, {"_id", 1}}).SetLimit(batchSize)
	cursor, err := globals.DocDb().Collection("People").Find(
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

	var people []data.GooglePerson
	if err := cursor.All(r, &people); err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var nextId string
	var nextUpdatedAt string
	if len(people) > 0 {
		last := people[len(people)-1]
		nextId = last.PersonId
		nextUpdatedAt = last.UpdatedAt.Format(time.RFC3339Nano)
	} else {
		nextId = personId
		nextUpdatedAt = updatedAtStr
	}
	if len(people) < int(batchSize) {
		go client.CheckForGmailsUpdates(accountId)
	}

	r.JSON(200, PullPeopleResponse{
		People:     people,
		Checkpoint: SyncCheckpoint{PersonId: nextId, UpdatedAt: nextUpdatedAt},
	})
}
