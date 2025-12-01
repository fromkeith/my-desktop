package messages

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// PullStream godoc
// @Summary      Stream Messages
// @Description  Sync endpoint to allow for for push from server to client of changes to messages.
// @Tags         email
// @Produce      event-stream
// @Router       /messages/pullStream [get]
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
		if stream.Next(r.Request.Context()) {
			var email data.GmailEntry
			var ev bson.M
			if err := stream.Decode(&ev); err != nil {
				log.Error().
					Ctx(r).
					Err(err).
					Msg("failed to unmarshal email doc in stream")
				return false
			}
			full, ok := ev["fullDocument"]
			if !ok {
				return true
			}
			raw, _ := bson.Marshal(full)
			if err := bson.Unmarshal(raw, &email); err != nil {
				log.Error().
					Ctx(r).
					Err(err).
					Any("full", full).
					Msg("failed to unmarshal email in stream")
				return true
			}

			payload, _ := json.Marshal(PullMessagesResponse{
				Messages:   []data.GmailEntry{email},
				Checkpoint: SyncCheckpoint{MessageId: email.MessageId, UpdatedAt: email.UpdatedAt.Format(time.RFC3339Nano)},
			})
			r.SSEvent("message", payload)
			// we return so gin can flush, we will get called again
			return true
		}
		if err := stream.Err(); err != nil {
			// same as context, so probably disconnect
			if err == r.Request.Context().Err() {
				return false
			}
			log.Error().
				Ctx(r).
				Err(err).
				Msg("stream failed")
		}
		return false
	})

}
