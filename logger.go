package logger

import (
	"context"
	"log/slog"
	"os"
)

//nolint:gochecknoglobals // ...
var globalLogger *slog.Logger = slog.New(newLoggerHandler(LevelDebug, os.Stdout))

func SetLevel(l slog.Level) {
	globalLogger = slog.New(newLoggerHandler(l, os.Stdout))
}

const (
	LevelEmergency = slog.Level(10000)
	LevelAlert     = slog.Level(1000)
	LevelCritial   = slog.Level(100)
	LevelError     = slog.LevelError
	LevelWarn      = slog.LevelWarn
	LevelNotice    = slog.Level(2)
	LevelInfo      = slog.LevelInfo
	LevelDebug     = slog.LevelDebug
)

func Fatal(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelEmergency, message, attrs...)

	os.Exit(1)
}

func Emergency(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelEmergency, message, attrs...)
}

func Alert(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelAlert, message, attrs...)
}

func Critial(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelCritial, message, attrs...)
}

func Error(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.ErrorContext(ctx, message, attrs...)
}

func Warn(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.WarnContext(ctx, message, attrs...)
}

func Notice(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelNotice, message, attrs...)
}

func Info(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.InfoContext(ctx, message, attrs...)
}

func Debug(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.DebugContext(ctx, message, attrs...)
}
