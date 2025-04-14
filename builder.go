package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type _key string

//nolint:gochecknoglobals // ...
var loggerKey _key = "_core_logger"

type LoggerOpt func(p *loggerParams)

func NewLoggerContext(ctx context.Context, opts ...LoggerOpt) context.Context {
	p := new(loggerParams)

	for _, o := range opts {
		o(p)
	}

	log := p.build()

	return context.WithValue(ctx, loggerKey, log)
}

type loggerParams struct {
	local     bool
	addSource bool
	lvl       slog.Level
	writers   []io.Writer
	handler   slog.Handler
}

func WithWriter(w io.Writer) LoggerOpt {
	return func(p *loggerParams) {
		p.writers = append(p.writers, w)
	}
}

func WithLevel(l slog.Level) LoggerOpt {
	return func(p *loggerParams) {
		p.lvl = l
	}
}

func Local() LoggerOpt {
	return func(p *loggerParams) {
		p.local = true
	}
}

func WithSource() LoggerOpt {
	return func(p *loggerParams) {
		p.addSource = true
	}
}

func WithHandler(h slog.Handler) LoggerOpt {
	return func(p *loggerParams) {
		p.handler = h
	}
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func MapLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "notice":
		return LevelNotice
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "critical":
		return LevelCritial
	case "alert":
		return LevelAlert
	case "emergency":
		return LevelEmergency
	default:
		return LevelInfo
	}
}

func (b *loggerParams) build() *slog.Logger {
	if len(b.writers) == 0 {
		b.writers = append(b.writers, os.Stdout)
	}

	w := io.MultiWriter(b.writers...)

	var handler slog.Handler

	if b.local {
		opts := prettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level:     b.lvl,
				AddSource: b.addSource,
			},
		}

		if b.handler != nil {
			handler = b.handler
		} else {
			handler = opts.newPrettyHandler(w)
		}

		return slog.New(handler)
	}

	if b.handler != nil {
		handler = b.handler
	} else {
		handler = newLoggerHandler(b.lvl, w)
	}

	return slog.New(handler)
}

func newLoggerHandler(lvl slog.Level, w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)

				switch {
				case level < LevelInfo:
					a.Value = slog.StringValue("DEBUG")
				case level < LevelNotice:
					a.Value = slog.StringValue("INFO")
				case level < LevelWarn:
					a.Value = slog.StringValue("NOTICE")
				case level < LevelError:
					a.Value = slog.StringValue("WARNING")
				case level < LevelCritial:
					a.Value = slog.StringValue("ERROR")
				case level < LevelAlert:
					a.Value = slog.StringValue("CRITICAL")
				case level < LevelEmergency:
					a.Value = slog.StringValue("ALERT")
				default:
					a.Value = slog.StringValue("EMERGENCY")
				}
			}

			return a
		},
	})
}

func loggerFromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}

	return globalLogger
}
