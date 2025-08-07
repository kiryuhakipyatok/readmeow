package logger

import (
	"log/slog"
	"os"
)

const (
	localEnv = "local"
	devEnv   = "dev"
	prodEnv  = "prod"
)

type Logger struct {
	Log *slog.Logger
}

func NewLogger(env string) *Logger {
	var log *slog.Logger
	switch env {
	case localEnv:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case devEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case prodEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	logger := &Logger{
		Log: log.With(slog.String("env", env)),
	}
	return logger
}

func (l *Logger) AddOp(op string) {
	l.Log = l.Log.With(slog.String("op", op))
}
