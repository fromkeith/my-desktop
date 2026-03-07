package helpers

import (
	"context"
	"fmt"
	"fromkeith/my-desktop-server/shared/globals"
	"time"

	"github.com/hibiken/asynq"
)

const (
	// MaxRetries specifies the maximum number of times a task will be retried.
	MaxRetries = 10
)

// RequeueTaskWithBackoff checks the current retry count of a task and if it's below MaxRetries,
// it enqueues a new task with the same payload to be processed after a delay.
// The delay increases with the number of retries.
// It is intended to be called from within a task handler that has failed.
// The handler should return nil after calling this function to prevent Asynq's default retry mechanism.
func RequeueTaskWithBackoff(ctx context.Context, task *asynq.Task) error {
	retryCount, _ := asynq.GetRetryCount(ctx)

	if retryCount >= MaxRetries {
		return fmt.Errorf("task %s exceeded max retries of %d", task.Type(), MaxRetries)
	}

	// Exponential backoff: 5s, 20s, 45s, ...
	delay := time.Duration((retryCount+1)*(retryCount+1)) * 5 * time.Second

	// Create a new task with the same type and payload.
	newTask := asynq.NewTask(task.Type(), task.Payload())

	// Enqueue the new task to be processed after the delay.
	// The new task will be processed by the same queue as the original task.
	_, err := globals.Asynq().EnqueueContext(ctx, newTask, asynq.ProcessIn(delay))
	if err != nil {
		return fmt.Errorf("failed to requeue task %s: %w", task.Type(), err)
	}

	return nil
}
