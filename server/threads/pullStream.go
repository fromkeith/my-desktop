package threads

import (
	"context"
	"fromkeith/my-desktop-server/globals"
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
// @Summary      Stream Threads
// @Description  Sync endpoint to allow for for push from server to client of changes to Threads.
// @Tags         email
// @Produce      event-stream
// @Router       /threads/pullStream [get]
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
	stream, err := globals.DocDb().Collection("MessageThreads").Watch(r, mongo.Pipeline{matchStage}, opts)
	if err != nil {
		r.Error(err)
		return
	}
	defer stream.Close(r)

	streamCtx, cancel := context.WithCancel(r.Request.Context())
	defer cancel()

	batchChan, batchErr := utils.BatchMongoStreamChannel(streamCtx, stream, 10, time.Second)

	r.Set("Content-Type", "text/event-stream")
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
			log.Info().Msg("pullStream for threads had batch")
			payloads := make([]ThreadEntry, 0, len(batch))
			chkPoint := SyncCheckpoint{}
			for _, ev := range batch {
				full, ok := ev["fullDocument"]
				if !ok {
					return true
				}
				raw, _ := bson.Marshal(full)
				var thread ThreadEntry
				if err := bson.Unmarshal(raw, &thread); err != nil {
					log.Error().
						Ctx(r).
						Err(err).
						Any("full", full).
						Msg("failed to unmarshal thread in stream")
					return true
				}
				payloads = append(payloads, thread)
				at := thread.UpdatedAt.Format(time.RFC3339Nano)
				if at > chkPoint.UpdatedAt {
					chkPoint = SyncCheckpoint{ThreadId: thread.ThreadId, UpdatedAt: at}
				} else if at == chkPoint.UpdatedAt && thread.ThreadId > chkPoint.ThreadId {
					chkPoint = SyncCheckpoint{ThreadId: thread.ThreadId, UpdatedAt: at}
				}
			}
			payload, _ := json.Marshal(PullThreadResponse{
				Threads:    payloads,
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
