package logger

import (
	"io"
	"log/slog"
	"os"
)

// Уровни логирования
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
}

type SLogger struct {
	logger *slog.Logger
}

func NewLogger(w io.Writer, level slog.Level) *SLogger {
	return &SLogger{
		logger: slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		})),
	}
}

func (l *SLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *SLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *SLogger) Fatal(msg string, args ...any) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}