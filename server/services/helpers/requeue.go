package helpers

import (
	"context"
	"fromkeith/my-desktop-server/globals"

	"github.com/wagslane/go-rabbitmq"
)

const (
	backoffQueueName = "backoff_queue"
)

var (
	backoffPublisher *rabbitmq.Publisher
)

// Uses rabbitMq's delayed message exchange plugin to backoff on failed messages
func BackoffMessage(ctx context.Context, msg rabbitmq.Delivery) error {
	backoffPublisher, err = rabbitmq.NewPublisher(
		globals.Rabbit(),
		rabbitmq.WithPublisherOptionsExchangeName("delay-retry-exchange"),
		rabbitmq.WithPublisherOptionsExchangeKind("x-delayed-message"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeArgs(rabbitmq.Table{
			"x-delayed-type": "direct",
		}),
	)

	globals.Rabbit().PublishWithContext(
		ctx,
		msg.Body,
		[]string{"delay-retry-queue"},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange("emails"),
		rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table{
			"x-retry-count": 1,
			"x-routing-key": msg.RoutingKey,
		}),
	)
	return nil
}

// Pushes the message into the backoff queue
func RequeueMessage(ctx context.Context, queueName string, msg rabbitmq.Delivery) error {
	if backoffPublisher == nil {
		var err error
		backoffPublisher, err = rabbitmq.NewPublisher(
			globals.Rabbit(),
			rabbitmq.WithPublisherOptionsExchangeName("emails"),
			rabbitmq.WithPublisherOptionsExchangeDeclare,
		)
		if err != nil {
			// make sure we nack, and requeue
			msg.Nack(false, true)
			return err
		}
	}
	// nack it, and don't requeue
	defer msg.Nack(false, false)

	retryCount := 0
	if len(msg.Headers) > 0 {
		if cnt, ok := msg.Headers["x-retry-count"].(int); ok {
			retryCount = cnt
		}
		if retryCount > 10 {

		}
	}
	headers := rabbitmq.Table{
		"x-retry-count": retryCount + 1,
		"x-routing-key": msg.RoutingKey,
		"x-queue-name":  queueName,
	}
	return backoffPublisher.PublishWithContext(
		ctx,
		msg.Body,
		[]string{backoffQueueName},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange("emails"),
		rabbitmq.WithPublishOptionsHeaders(headers),
	)

}
