package logger

import (
	"log/slog"
	"os"
	"readmeow/internal/config"
)

const (
	localEnv = "local"
	devEnv   = "dev"
	prodEnv  = "prod"
)

type Logger struct {
	Log *slog.Logger
}

func NewLogger(env string, acfg config.AppConfig) *Logger {
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
		Log: log.With(
			slog.String("env", env),
			slog.String("app", acfg.Name),
			slog.String("version", acfg.Version),
		),
	}
	return logger
}

func (l *Logger) AddOp(op string) *Logger {
	logger := &Logger{
		Log: l.Log.With(slog.String("op", op)),
	}
	return logger
}
