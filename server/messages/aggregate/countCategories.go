package aggregate

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CountCategoriesResponse struct {
	Categories []data.AccountCategory
}

func CountCategories(r *gin.Context) {
	cur, err := globals.DocDb().Collection("AccountCategories").Find(
		r,
		bson.M{
			"messageCount": bson.D{{"$gte", 1}},
			"accountId":    r.GetString("accountId"),
		},
		options.Find().SetSort(bson.D{{"messageCount", -1}}).SetLimit(50),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count categories")
		r.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count messages"})
		return
	}
	var all []data.AccountCategory
	if err := cur.All(r, &all); err != nil {
		log.Error().Err(err).Msg("Failed to count categories (decode)")
		r.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count messages"})
		return
	}
	r.JSON(http.StatusOK, CountCategoriesResponse{all})
}
