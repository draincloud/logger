package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"runtime"

	"github.com/fatih/color"
)

type prettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type prettyHandler struct {
	slog.Handler
	l         *stdlog.Logger
	attrs     []slog.Attr
	prefix    string
	addSource bool
}

func (opts prettyHandlerOptions) newPrettyHandler(
	out io.Writer,
) *prettyHandler {
	h := &prettyHandler{
		Handler:   slog.NewJSONHandler(out, opts.SlogOpts),
		l:         stdlog.New(out, "", 0),
		addSource: opts.SlogOpts != nil && opts.SlogOpts.AddSource,
	}

	return h
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make(map[string]any, len(h.attrs)+r.NumAttrs())

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	r.Attrs(func(a slog.Attr) bool {
		fields[h.prefix+a.Key] = a.Value.Any()

		return true
	})

	var (
		b   []byte
		err error
	)

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return fmt.Errorf("pretty handler: marshal attrs: %w", err)
		}
	}

	h.l.Println(
		r.Time.Format("[15:04:05.000]"),
		colorLevel(r.Level),
		color.CyanString(r.Message),
		h.source(r),
		color.WhiteString(string(b)),
	)

	return nil
}

// source renders the caller of the logging call, when the handler was built with
// AddSource.
func (h *prettyHandler) source(r slog.Record) string {
	if !h.addSource || r.PC == 0 {
		return ""
	}

	f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()

	return color.HiBlackString("(%s:%d)", f.File, f.Line)
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// the attrs already held were captured under their own groups, so only the new
	// ones take the current prefix
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)

	for _, a := range attrs {
		merged = append(merged, slog.Attr{Key: h.prefix + a.Key, Value: a.Value})
	}

	return &prettyHandler{
		Handler:   h.Handler,
		l:         h.l,
		attrs:     merged,
		prefix:    h.prefix,
		addSource: h.addSource,
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{
		Handler:   h.Handler.WithGroup(name),
		l:         h.l,
		attrs:     h.attrs,
		prefix:    h.prefix + name + ".",
		addSource: h.addSource,
	}
}

// colorLevel labels a record with the same name the JSON handler uses, coloured by
// severity, so the levels this package adds are not printed as slog offsets.
func colorLevel(lvl slog.Level) string {
	name := LevelName(lvl) + ":"

	switch {
	case lvl < LevelInfo:
		return color.MagentaString(name)
	case lvl < LevelNotice:
		return color.BlueString(name)
	case lvl < LevelWarn:
		return color.GreenString(name)
	case lvl < LevelError:
		return color.YellowString(name)
	case lvl < LevelCritical:
		return color.RedString(name)
	default:
		return color.HiRedString(name)
	}
}
