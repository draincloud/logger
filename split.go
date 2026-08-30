package logger

import (
	"context"
	"log/slog"
)

// splitHandler routes a record to one of two handlers by level, so that errors can
// go to a different destination than ordinary output.
type splitHandler struct {
	low, high slog.Handler
}

func (h *splitHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.pick(lvl).Enabled(ctx, lvl)
}

func (h *splitHandler) Handle(ctx context.Context, r slog.Record) error {
	//nolint:wrapcheck // the chosen handler already names what went wrong
	return h.pick(r.Level).Handle(ctx, r)
}

func (h *splitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitHandler{low: h.low.WithAttrs(attrs), high: h.high.WithAttrs(attrs)}
}

func (h *splitHandler) WithGroup(name string) slog.Handler {
	return &splitHandler{low: h.low.WithGroup(name), high: h.high.WithGroup(name)}
}

func (h *splitHandler) pick(lvl slog.Level) slog.Handler {
	if lvl >= LevelError {
		return h.high
	}

	return h.low
}
