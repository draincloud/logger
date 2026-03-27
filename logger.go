package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

//nolint:gochecknoglobals // ...
var globalLogger *slog.Logger = slog.New(newLoggerHandler(LevelDebug, os.Stdout))

func SetGlobalLogger(l *slog.Logger) {
	globalLogger = l
}

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

type LogFunc func(context.Context, string, ...any)

var (
	Falalf     LogFunc = FatalKV
	Emergencyf LogFunc = EmergencyKV
	Alertf     LogFunc = AlertKV
	Critialf   LogFunc = CritialKV
	Errorf     LogFunc = ErrorKV
	Warnf      LogFunc = WarnKV
	Noticef    LogFunc = NoticeKV
	Infof      LogFunc = InfoKV
	Debugf     LogFunc = DebugKV
)

func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	l := loggerFromCtx(ctx)
	if l == globalLogger {
		lcopy := *l
		l = &lcopy
	}
	for _, a := range attrs {
		l = l.With(a)
	}
	return context.WithValue(ctx, loggerKey, l)
}

func WithGroup(ctx context.Context, name string) context.Context {
	l := loggerFromCtx(ctx)
	if l == globalLogger {
		lcopy := *l
		l = &lcopy
	}

	return context.WithValue(ctx, loggerKey, l.WithGroup(name))
}

func Fatal(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelEmergency, message, attrs...)
	panic(fmt.Sprintf(message, attrs...))
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

func FatalKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelEmergency, fmt.Sprintf(message, attrs...))
	panic(fmt.Sprintf(message, attrs...))
}

func EmergencyKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelEmergency, fmt.Sprintf(message, attrs...))
}

func AlertKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelAlert, fmt.Sprintf(message, attrs...))
}

func CritialKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelCritial, fmt.Sprintf(message, attrs...))
}

func ErrorKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.ErrorContext(ctx, fmt.Sprintf(message, attrs...))
}

func WarnKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.WarnContext(ctx, fmt.Sprintf(message, attrs...))
}

func NoticeKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.Log(ctx, LevelNotice, fmt.Sprintf(message, attrs...))
}

func InfoKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.InfoContext(ctx, fmt.Sprintf(message, attrs...))
}

func DebugKV(ctx context.Context, message string, attrs ...any) {
	l := loggerFromCtx(ctx)

	l.DebugContext(ctx, fmt.Sprintf(message, attrs...))
}
