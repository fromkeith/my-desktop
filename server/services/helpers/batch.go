package helpers

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/wagslane/go-rabbitmq"
)

// returns a channel that receives batches of messages from a RabbitMQ queue
func CreateConsumeBatch(ctx context.Context, c *rabbitmq.Consumer, maxMessages int, fillWindow time.Duration) (chan []rabbitmq.Delivery, chan error) {

	var result = make(chan []rabbitmq.Delivery)
	var single = make(chan rabbitmq.Delivery)
	var errChan = make(chan error)

	go func() {
		defer close(errChan)
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Ctx(ctx).
					Stack().
					Err(err.(error)).
					Msg("panic reading from queue")
			}
		}()
		err := c.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			single <- d
			return rabbitmq.Manual
		})
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer close(result)
		defer close(single)

		writeWait := make([]rabbitmq.Delivery, 0, maxMessages)

		defer func() {
			// ensure all messages are nacked if the context is done before the batch is full
			if len(writeWait) > 0 {
				for _, w := range writeWait {
					w.Nack(false, true)
				}
			}
		}()

		for {
			select {
			case d := <-single:
				writeWait = append(writeWait, d)
				if len(writeWait) >= maxMessages {
					result <- writeWait
					writeWait = writeWait[:0]
				}
			case <-time.After(time.Second):
				if len(writeWait) >= maxMessages && len(writeWait) > 0 {
					result <- writeWait
					writeWait = writeWait[:0]
				}
			case <-ctx.Done():
				return
			}
		}

	}()

	return result, errChan

}

// FetchBatch pulls up to N messages from Kafka. It waits indefinitely for the first
// message (respecting ctx). Once the first arrives, it allows up to
// fillWindow for additional messages before returning.
func FetchBatch(ctx context.Context, r *kafka.Reader, maxMessages int, fillWindow time.Duration) ([]kafka.Message, error) {
	batch := make([]kafka.Message, 0, maxMessages)
	var deadline time.Time // zero until first message

	for len(batch) < maxMessages {
		// Build a per-fetch context: no deadline until first message arrives,
		// then time out when the shared fill window expires.
		var fetchCtx context.Context
		var cancel context.CancelFunc = func() {}
		if deadline.IsZero() {
			// Wait indefinitely for the first message (only canceled by ctx)
			fetchCtx = ctx
		} else {
			remain := time.Until(deadline)
			if remain <= 0 {
				break // fill window expired
			}
			fetchCtx, cancel = context.WithTimeout(ctx, remain)
		}

		m, err := r.FetchMessage(fetchCtx)
		cancel()

		if err != nil {
			// If we were in the fill window and just hit its timeout, return what we have
			if !deadline.IsZero() && errors.Is(err, context.DeadlineExceeded) && len(batch) > 0 {
				break
			}
			// Respect shutdown / reader close
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				return batch, err
			}
			// Temporary/other fetch errors: try again unless we’re out of time
			continue
		}

		batch = append(batch, m)
		if deadline.IsZero() {
			deadline = time.Now().Add(fillWindow) // start the 5s fill window
		}
	}

	return batch, nil
}
