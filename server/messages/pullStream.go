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
			{"fullDocument.accountId", accountId},
		},
	}}
	opts := options.ChangeStream().
		SetFullDocument(options.UpdateLookup).
		SetMaxAwaitTime(10 * time.Second)
	stream, err := globals.DocDb().Collection("Messages").Watch(r, mongo.Pipeline{matchStage}, opts)
	if err != nil {
		r.Error(err)
		return
	}
	defer stream.Close(r)

	r.Stream(func(w io.Writer) bool {
		log.Println("stream start")
		if stream.Next(r.Request.Context()) {
			var email data.GmailEntry
			var ev bson.M
			if err := stream.Decode(ev); err != nil {
				log.Println("failed to unmarshal email", err)
				return false
			}
			full, ok := ev["fullDocument"]
			if !ok {
				return true
			}
			raw, _ := bson.Marshal(full)
			if err := bson.Unmarshal(raw, &email); err != nil {
				log.Println("failed to unmarshal email", err)
				return true
			}

			log.Println("Sent sync message")
			r.SSEvent("message", PullMessagesResponse{
				Messages:   []data.GmailEntry{email},
				Checkpoint: SyncCheckpoint{MessageId: email.MessageId, UpdatedAt: email.UpdatedAt.Format(time.RFC3339Nano)},
			})
			// we return so gin can flush, we will get called again
			return true
		}
		if err := stream.Err(); err != nil {
			// same as context, so probably disconnect
			if err == r.Request.Context().Err() {
				return false
			}
			log.Println("stream failed!", err)
		}
		return false
	})

}
