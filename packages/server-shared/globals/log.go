package globals

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	closeLogger func() error
)

// overwrite the logger to write to console and file
func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	l, f, err := newLogger()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Failed to create logger")
	}

	log.Logger = *l

	var once sync.Once
	closeLogger = func() error {
		var err error
		once.Do(func() {
			// try to flush file to disk
			f.Sync()
			f.Close()
		})
		return err
	}
}

func newLogger() (*zerolog.Logger, *os.File, error) {
	// Open or create the log file (append mode)
	exename, err := os.Executable()
	if err != nil {
		exename = "unknown"
	}
	exename = filepath.Base(filepath.Clean(exename))
	log.Print("exename:", exename)
	f, err := os.OpenFile("log_"+exename+".log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, nil, err
	}

	// Pretty console writer
	cw := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Send events to both: pretty console AND raw JSON file
	multi := zerolog.MultiLevelWriter(cw, f)

	// Build the logger
	logger := zerolog.New(multi).
		Hook(ContextExtracHook{}).
		With().
		Timestamp().
		Logger()

	return &logger, f, nil
}

type zlWriter struct {
	L   *zerolog.Logger
	Lvl zerolog.Level
}

func (w zlWriter) Write(p []byte) (int, error) {
	// Trim newline Gin adds
	msg := strings.TrimRight(string(p), "\r\n")
	w.L.WithLevel(w.Lvl).Msg(msg)
	return len(p), nil
}

// Call after Log is constructed (e.g., at end of your init())
func HookGin() {
	gin.DefaultWriter = zlWriter{L: &log.Logger, Lvl: zerolog.InfoLevel}
	gin.DefaultErrorWriter = zlWriter{L: &log.Logger, Lvl: zerolog.ErrorLevel}
}

type ContextExtracHook struct{}

func (h ContextExtracHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if accountId, ok := ctx.Value("accountId").(string); ok && accountId != "" {
		e.Str("accountId", accountId)
	}
	if requestId, ok := ctx.Value("requestId").(string); ok && requestId != "" {
		e.Str("requestId", requestId)
	}
	if service, ok := ctx.Value("service").(string); ok && service != "" {
		e.Str("service", service)
	}
}
