package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"readmeow/internal/config"
)

type Logger struct {
	Log *slog.Logger
}

const (
	local = "local"
	dev   = "dev"
	prod  = "prod"
	test  = "test"
)

func NewLogger(acfg config.AppConfig) *Logger {
	var log *slog.Logger
	env := acfg.Env
	writer := io.Writer(os.Stdout)
	if acfg.LogPath != "" {
		logFile, err := os.OpenFile(acfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %w", err))
		}
		writer = io.MultiWriter(logFile, os.Stdout)
	}
	switch env {
	case local:
		log = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case dev:
		log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case prod:
		log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case test:
		log = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	default:
		log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	logger := &Logger{
		Log: log.With(
			slog.String("type", "app"),
			slog.String("env", env),
			slog.String("app", acfg.Name),
			slog.String("version", acfg.Version),
		),
	}
	return logger
}

func (l *Logger) Info(msg string, args ...any) {
	l.Log.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Log.Error(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Log.Debug(msg, args...)
}

func (l *Logger) AddOp(op string) *Logger {
	return &Logger{
		Log: l.Log.With(slog.String("op", op)),
	}
}
