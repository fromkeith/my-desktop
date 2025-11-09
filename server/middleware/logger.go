package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var (
	logger = zerolog.New(os.Stdout)
)

func init() {
	logger = logger.Hook(TracingHook{})
}

type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if r, ok := ctx.(*gin.Context); ok {
		accountId := r.GetString("accountId")
		if accountId != "" {
			e.Str("accountId", accountId)
		}
		e.Str("requestId", r.GetString("requestId"))
	}
}
