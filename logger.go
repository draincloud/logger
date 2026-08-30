package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

//nolint:gochecknoglobals // ...
var (
	globalLevel  = new(slog.LevelVar)
	globalLogger = newGlobalLogger()
)

func newGlobalLogger() *atomic.Pointer[slog.Logger] {
	p := new(atomic.Pointer[slog.Logger])
	p.Store(slog.New(newLoggerHandler(globalLevel, false, os.Stdout)))

	return p
}

func SetGlobalLogger(l *slog.Logger) {
	if l == nil {
		return
	}

	globalLogger.Store(l)
	slog.SetDefault(l)
}

// SetLevel sets the level of the logger
func SetLevel(l slog.Level) {
	globalLevel.Set(l)
}

const (
	LevelEmergency = slog.Level(10000)
	LevelAlert     = slog.Level(1000)
	LevelCritical  = slog.Level(100)
	LevelError     = slog.LevelError
	LevelWarn      = slog.LevelWarn
	LevelNotice    = slog.Level(2)
	LevelInfo      = slog.LevelInfo
	LevelDebug     = slog.LevelDebug
)

// LevelName renders lvl under this package's level names.
func LevelName(lvl slog.Level) string {
	switch {
	case lvl < LevelInfo:
		return "DEBUG"
	case lvl < LevelNotice:
		return "INFO"
	case lvl < LevelWarn:
		return "NOTICE"
	case lvl < LevelError:
		return "WARNING"
	case lvl < LevelCritical:
		return "ERROR"
	case lvl < LevelAlert:
		return "CRITICAL"
	case lvl < LevelEmergency:
		return "ALERT"
	default:
		return "EMERGENCY"
	}
}

// Enabled reports whether the context's logger emits records at level.
func Enabled(ctx context.Context, level slog.Level) bool {
	return loggerFromCtx(ctx).Enabled(ctx, level)
}

func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	l := loggerFromCtx(ctx)
	for _, a := range attrs {
		l = l.With(a)
	}

	return context.WithValue(ctx, loggerKey, l)
}

func WithGroup(ctx context.Context, name string) context.Context {
	l := loggerFromCtx(ctx)

	return context.WithValue(ctx, loggerKey, l.WithGroup(name))
}

func Fatal(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelEmergency, message, attrs...)
	os.Exit(1)
}

func Emergency(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelEmergency, message, attrs...)
}

func Alert(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelAlert, message, attrs...)
}

func Critical(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelCritical, message, attrs...)
}

func Error(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelError, message, attrs...)
}

func Warn(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelWarn, message, attrs...)
}

func Notice(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelNotice, message, attrs...)
}

func Info(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelInfo, message, attrs...)
}

func Debug(ctx context.Context, message string, attrs ...any) {
	log(ctx, LevelDebug, message, attrs...)
}

func Fatalf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelEmergency, sprintf(format, args...))
	os.Exit(1)
}

func Emergencyf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelEmergency, sprintf(format, args...))
}

func Alertf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelAlert, sprintf(format, args...))
}

func Criticalf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelCritical, sprintf(format, args...))
}

func Errorf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelError, sprintf(format, args...))
}

func Warnf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelWarn, sprintf(format, args...))
}

func Noticef(ctx context.Context, format string, args ...any) {
	log(ctx, LevelNotice, sprintf(format, args...))
}

func Infof(ctx context.Context, format string, args ...any) {
	log(ctx, LevelInfo, sprintf(format, args...))
}

func Debugf(ctx context.Context, format string, args ...any) {
	log(ctx, LevelDebug, sprintf(format, args...))
}

func log(ctx context.Context, level slog.Level, message string, attrs ...any) {
	l := loggerFromCtx(ctx)
	if !l.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr

	runtime.Callers(3, pcs[:])

	r := slog.NewRecord(time.Now(), level, message, pcs[0])
	r.Add(attrs...)

	_ = l.Handler().Handle(ctx, r)
}

func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}

	return fmt.Sprintf(format, args...)
}
