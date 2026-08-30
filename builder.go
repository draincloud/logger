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

// LoggerOpt options for logger builder.
type LoggerOpt func(p *loggerParams)

// NewLoggerContext creates a new context woth logger.
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
	lvl       slog.Leveler
	writers   []io.Writer
	errWriter io.Writer
	handler   slog.Handler
}

// WithWriter sets a writer.
func WithWriter(w io.Writer) LoggerOpt {
	return func(p *loggerParams) {
		p.writers = append(p.writers, w)
	}
}

// WithLevel sets logging level. Pass a *slog.LevelVar instead of a slog.Level to
// keep control of the level after the logger is built.
func WithLevel(l slog.Leveler) LoggerOpt {
	return func(p *loggerParams) {
		p.lvl = l
	}
}

// WithErrorWriter sends records at error level and above to w, and everything below
// it to the writers set by WithWriter. Ignored when WithHandler supplies a handler.
func WithErrorWriter(w io.Writer) LoggerOpt {
	return func(p *loggerParams) {
		p.errWriter = w
	}
}

// Local sets a pretty handler for a logger.
func Local() LoggerOpt {
	return func(p *loggerParams) {
		p.local = true
	}
}

// WithSource adds caller to a logging entry.
func WithSource() LoggerOpt {
	return func(p *loggerParams) {
		p.addSource = true
	}
}

// WithHandler sets custom handler.
func WithHandler(h slog.Handler) LoggerOpt {
	return func(p *loggerParams) {
		p.handler = h
	}
}

// Err is an easy to use error logging attribute.
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

// MapLevel maps string level to slog.
func MapLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "notice":
		return LevelNotice
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "critical":
		return LevelCritical
	case "alert":
		return LevelAlert
	case "emergency":
		return LevelEmergency
	default:
		return LevelInfo
	}
}

func (b *loggerParams) build() *slog.Logger {
	if b.handler != nil {
		return slog.New(b.handler)
	}

	if len(b.writers) == 0 {
		b.writers = append(b.writers, os.Stdout)
	}

	if b.lvl == nil {
		b.lvl = LevelInfo
	}

	w := io.MultiWriter(b.writers...)

	if b.local {
		opts := prettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level:     b.lvl,
				AddSource: b.addSource,
			},
		}

		build := func(w io.Writer) slog.Handler { return opts.newPrettyHandler(w) }

		return slog.New(b.route(build, w))
	}

	build := func(w io.Writer) slog.Handler { return newLoggerHandler(b.lvl, b.addSource, w) }

	return slog.New(b.route(build, w))
}

// route pairs the handler for w with one for the error writer, when there is one.
func (b *loggerParams) route(build func(io.Writer) slog.Handler, w io.Writer) slog.Handler {
	if b.errWriter == nil {
		return build(w)
	}

	return &splitHandler{low: build(w), high: build(b.errWriter)}
}

func newLoggerHandler(lvl slog.Leveler, addSource bool, w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: addSource,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if level, ok := a.Value.Any().(slog.Level); ok {
					a.Value = slog.StringValue(LevelName(level))
				}
			}

			return a
		},
	})
}

// FromContext returns logger from context.
func FromContext(ctx context.Context) *slog.Logger {
	return loggerFromCtx(ctx)
}

// WithLogger returns context with logger l.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func loggerFromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}

	return globalLogger.Load()
}
