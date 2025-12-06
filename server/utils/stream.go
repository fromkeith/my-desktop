package utils

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func MongoStreamChannel(ctx context.Context, stream *mongo.ChangeStream) (chan bson.M, chan error) {
	streamRes := make(chan bson.M)
	errChan := make(chan error)
	go func() {
		for stream.Next(ctx) {
			var ev bson.M
			if err := stream.Decode(&ev); err != nil {
				// NOTE: must pass &ev — decoding into a nil value triggers "cannot Decode to nil value"
				log.Error().Ctx(ctx).Stack().Err(err).Msg("decode error")
				continue
			}
			streamRes <- ev
		}
		close(streamRes)
		if err := stream.Err(); err != nil && !errors.Is(err, context.Canceled) {
			errChan <- err
		}
		close(errChan)
	}()
	return streamRes, errChan
}

func BatchMongoStreamChannel(ctx context.Context, inputStream *mongo.ChangeStream, max int, timeout time.Duration) (chan []bson.M, chan error) {

	bulk := make([]bson.M, 0, max)

	stream, errStream := MongoStreamChannel(ctx, inputStream)
	outErrStream := make(chan error)
	bulkStream := make(chan []bson.M)

	go func() {
		for {
			select {
			case item := <-stream:
				bulk = append(bulk, item)
				if len(bulk) == max {
					bulkStream <- bulk
					bulk = make([]bson.M, 0, max)
				}
			case err := <-errStream:
				outErrStream <- err
				close(bulkStream)
				close(outErrStream)
				return
			case <-time.After(timeout):
				if len(bulk) > 0 {
					bulkStream <- bulk
					bulk = make([]bson.M, 0, max)
				}
			}
		}
	}()

	return bulkStream, outErrStream
}
