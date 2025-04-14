package logger

import (
	"context"
	"encoding/json"
	"io"
	stdlog "log"
	"log/slog"

	"github.com/fatih/color"
)

type prettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type prettyHandler struct {
	opts prettyHandlerOptions
	slog.Handler
	l     *stdlog.Logger
	attrs []slog.Attr
}

func (opts prettyHandlerOptions) newPrettyHandler(
	out io.Writer,
) *prettyHandler {
	h := &prettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdlog.New(out, "", 0),
	}

	return h
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
	}
}
