package messages

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func PullStream(r *gin.Context) {
	accountId := r.GetString("accountId")

	matchStage := bson.D{{
		"$match", bson.D{
			{"accountId", accountId},
		},
	}}
	opts := options.ChangeStream().SetMaxAwaitTime(2 * time.Second)
	stream, err := globals.DocDb().Collection("Messages").Watch(r, mongo.Pipeline{matchStage}, opts)
	if err != nil {
		r.Error(err)
		return
	}
	defer stream.Close(r)

	r.Stream(func(w io.Writer) bool {
		if stream.Next(r) {
			var email data.GmailEntry
			if err := stream.Decode(&email); err != nil {
				log.Println("failed to unmarshal email", err)
				return false
			}
			r.SSEvent("message", email)
			// we return so gin can flush, we will get called again
			return true
		}
		if err := stream.Err(); err != nil {
			// same as context, so probably disconnect
			if err == r.Err() {
				return false
			}
			log.Println("stream failed!", err)
		}
		return false
	})

}
