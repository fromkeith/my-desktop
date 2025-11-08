package helpers

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/segmentio/kafka-go"
)

// FetchBatch pulls up to N messages. It waits indefinitely for the first
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
