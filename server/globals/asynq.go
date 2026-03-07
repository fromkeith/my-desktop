package globals

import (
	"os"

	"github.com/hibiken/asynq"
)

var (
	asynqClient *asynq.Client
)

func Asynq() *asynq.Client {
	if asynqClient == nil {
		asynqClient = asynq.NewClient(asynq.RedisClientOpt{
			Addr: AsyncAddr(),
		})
	}
	return asynqClient
}

func AsyncAddr() string {
	return os.Getenv("REDIS_ADDR")
}
