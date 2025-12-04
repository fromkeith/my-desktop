package kafkaservice

import (
	"context"
	"errors"
	"fmt"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/services/helpers"
	"io"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type KafkaWorker func(context.Context, []kafka.Message) (dlq []kafka.Message, err error)

type KafkaService struct {
	Name        string
	Topic       string
	Group       string
	NumMessages int
	MaxWait     time.Duration
	NumWorkers  int
	Worker      KafkaWorker
	Dlq         string
}

func Run(ctx context.Context, opt KafkaService) {
	log.Info().
		Msg("Starting up " + opt.Name)
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	// cancel our context when we receive a terminate signal
	cancelContext, cancel := context.WithCancelCause(ctx)
	sigCtx, stop := signal.NotifyContext(
		cancelContext,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	go func() {
		<-sigCtx.Done()
		cancel(fmt.Errorf("received shutdown signal"))
		stop()
	}()

	// spin up each worker
	wg := sync.WaitGroup{}
	wg.Add(opt.NumWorkers)
	for range opt.NumWorkers {
		go func() {
			defer func() {
				rec := recover()
				if rec != nil {
					log.Error().Ctx(cancelContext).Stack().Err(rec.(error)).Msg("panic!")
				}
			}()
			defer wg.Done()
			run(cancelContext, opt)
		}()
	}
	// wait for workers to finish.. eg when our terminate signal is received, or all workers failed
	wg.Wait()
}

func run(ctx context.Context, opt KafkaService) {

	r := globals.KafkaConsumerGroup(opt.Topic, opt.Group)
	defer r.Close()

	dead := globals.KafkaWriter(opt.Dlq)
	defer dead.Close()

	for {
		log.Info().
			Ctx(ctx).
			Msg("Waiting for messages")
		msgs, err := helpers.FetchBatch(ctx, r, opt.NumMessages, opt.MaxWait)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				log.Info().
					Ctx(ctx).
					Msg("context canceled; exiting")
				break
			}
			var kerr *kafka.Error
			if errors.As(err, &kerr) && kerr.Temporary() {
				log.Warn().
					Ctx(ctx).
					Err(err).
					Msg("temporary kafka error")
				continue
			}
			if errors.Is(err, io.ErrClosedPipe) {
				log.Info().
					Ctx(ctx).
					Err(err).
					Msg("reader closed; exiting")
				break
			}
			log.Error().Ctx(ctx).Err(err).Msg("fetch error")
			continue
		}
		log.Info().
			Ctx(ctx).
			Int("count", len(msgs)).
			Msg("Got messages")

		failed, err := opt.Worker(ctx, msgs)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("worker error")
		}
		if err := r.CommitMessages(ctx, msgs...); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("Failed to commit messages")
		}
		if len(failed) > 0 {
			for i, f := range failed {
				// overwrite topic and other metadata
				failed[i] = kafka.Message{
					Key:     f.Key,
					Value:   f.Value,
					Headers: f.Headers,
				}
			}
			if err := dead.WriteMessages(ctx, failed...); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("failed to write messages to dead topic. Messages lost!")
			}
		}
	}
}
