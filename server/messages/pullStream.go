package messages

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/utils"
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

	streamCtx, cancel := context.WithCancel(r.Request.Context())
	defer cancel()

	batchChan, batchErr := utils.BatchMongoStreamChannel(streamCtx, stream, 10, time.Second)

	r.Stream(func(w io.Writer) bool {
		select {
		case err := <-batchErr:
			if err != nil {
				log.Error().
					Ctx(r).
					Err(err).
					Msg("failed to batch email docs in stream")
				return false
			}
		case batch := <-batchChan:
			payloads := make([]data.GmailEntry, 0, len(batch))
			chkPoint := SyncCheckpoint{}
			for _, ev := range batch {
				full, ok := ev["fullDocument"]
				if !ok {
					return true
				}
				raw, _ := bson.Marshal(full)
				var email data.GmailEntry
				if err := bson.Unmarshal(raw, &email); err != nil {
					log.Error().
						Ctx(r).
						Err(err).
						Any("full", full).
						Msg("failed to unmarshal email in stream")
					return true
				}
				ensureJsonEntry(&email)
				payloads = append(payloads, email)
				at := email.UpdatedAt.Format(time.RFC3339Nano)
				if at > chkPoint.UpdatedAt {
					chkPoint = SyncCheckpoint{MessageId: email.MessageId, UpdatedAt: at}
				} else if at == chkPoint.UpdatedAt && email.MessageId > chkPoint.MessageId {
					chkPoint = SyncCheckpoint{MessageId: email.MessageId, UpdatedAt: at}
				}
			}
			payload, _ := json.Marshal(PullMessagesResponse{
				Messages:   payloads,
				Checkpoint: chkPoint,
			})
			r.SSEvent("message", payload)
			return true
		case <-time.After(time.Second):
			return true // allow the request to check its status or end
		}
		return true
	})

}
