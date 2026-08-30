package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestMain(m *testing.M) {
	color.NoColor = true // the pretty handler's output is compared as plain text

	os.Exit(m.Run())
}

func jsonCtx(t *testing.T, opts ...LoggerOpt) (context.Context, *bytes.Buffer) {
	t.Helper()

	buf := new(bytes.Buffer)
	opts = append([]LoggerOpt{WithWriter(buf), WithLevel(LevelDebug)}, opts...)

	return NewLoggerContext(context.Background(), opts...), buf
}

func prettyCtx(t *testing.T, opts ...LoggerOpt) (context.Context, *bytes.Buffer) {
	t.Helper()

	return jsonCtx(t, append([]LoggerOpt{Local()}, opts...)...)
}

// Fatal must terminate the process: a panic would be recoverable and would
// change the exit code.
func TestFatalExitsWithCode1(t *testing.T) {
	if os.Getenv("LOGGER_TEST_FATAL") == "1" {
		Fatal(context.Background(), "boom")

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalExitsWithCode1")
	cmd.Env = append(os.Environ(), "LOGGER_TEST_FATAL=1")

	var exitErr *exec.ExitError
	if err := cmd.Run(); !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("want exit status 1, got %v", err)
	}
}

// A message carrying a literal % must not be run through printf when no
// formatting arguments are supplied.
func TestFormattedLiteralPercent(t *testing.T) {
	ctx, buf := jsonCtx(t)

	// Called through a func value so that vet's printf check, which now covers
	// Infof directly, does not reject the deliberately arg-less format string.
	logf := Infof
	logf(ctx, "disk usage 90% of quota")

	if !strings.Contains(buf.String(), "disk usage 90% of quota") {
		t.Fatalf("message was mangled: %s", buf.String())
	}
}

func TestFormattedWithArgs(t *testing.T) {
	ctx, buf := jsonCtx(t)

	Errorf(ctx, "user %d not found", 42)

	if !strings.Contains(buf.String(), "user 42 not found") {
		t.Fatalf("message was not formatted: %s", buf.String())
	}
}

// The attribute family must emit structured fields, not printf noise.
func TestAttrsReachOutput(t *testing.T) {
	ctx, buf := jsonCtx(t)

	Info(ctx, "user login", "user_id", 42)

	out := buf.String()
	if !strings.Contains(out, `"user_id":42`) || strings.Contains(out, "EXTRA") {
		t.Fatalf("attrs did not reach structured output: %s", out)
	}
}

func TestContextAttrsAccumulate(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts []LoggerOpt
	}{
		{"json", nil},
		{"pretty", []LoggerOpt{Local()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, buf := jsonCtx(t, tc.opts...)
			ctx = WithAttrs(ctx, slog.String("a", "1"), slog.String("b", "2"))

			Info(ctx, "two attrs")

			out := buf.String()
			if !strings.Contains(out, `"a"`) || !strings.Contains(out, `"b"`) {
				t.Fatalf("an earlier context attr was dropped: %s", out)
			}
		})
	}
}

func TestPrettyGroupKeepsEarlierAttrs(t *testing.T) {
	ctx, buf := prettyCtx(t)
	ctx = WithAttrs(ctx, slog.String("outer", "1"))
	ctx = WithGroup(ctx, "req")
	ctx = WithAttrs(ctx, slog.String("id", "abc"))

	Info(ctx, "grouped")

	out := buf.String()
	if !strings.Contains(out, `"outer"`) {
		t.Fatalf("attr set before the group was dropped: %s", out)
	}

	if !strings.Contains(out, `"req.id"`) {
		t.Fatalf("attr set inside the group was not prefixed: %s", out)
	}
}

// A record's own attrs must be nested under the group the context is in.
func TestPrettyRecordAttrsTakeGroupPrefix(t *testing.T) {
	ctx, buf := prettyCtx(t)
	ctx = WithGroup(ctx, "req")

	Info(ctx, "grouped", "id", "abc")

	if out := buf.String(); !strings.Contains(out, `"req.id"`) {
		t.Fatalf("record attr was not prefixed by the group: %s", out)
	}
}

// The levels this package adds must be named, not printed as slog offsets such as
// ERROR+92 or INFO+2.
func TestCustomLevelNames(t *testing.T) {
	levels := map[slog.Level]string{
		LevelDebug:     "DEBUG",
		LevelInfo:      "INFO",
		LevelNotice:    "NOTICE",
		LevelWarn:      "WARNING",
		LevelError:     "ERROR",
		LevelCritical:  "CRITICAL",
		LevelAlert:     "ALERT",
		LevelEmergency: "EMERGENCY",
	}

	for lvl, want := range levels {
		if got := LevelName(lvl); got != want {
			t.Errorf("LevelName(%v) = %q, want %q", lvl, got, want)
		}

		if got := MapLevel(want); got != lvl {
			t.Errorf("MapLevel(%q) = %v, want %v (LevelName and MapLevel must round-trip)", want, got, lvl)
		}

		ctx, buf := jsonCtx(t)
		log(ctx, lvl, "msg")

		if out := buf.String(); !strings.Contains(out, `"level":"`+want+`"`) {
			t.Errorf("json handler labelled %v as %s, want %s", lvl, out, want)
		}

		pctx, pbuf := prettyCtx(t)
		log(pctx, lvl, "msg")

		if out := pbuf.String(); !strings.Contains(out, want+":") {
			t.Errorf("pretty handler labelled %v as %s, want %s", lvl, out, want)
		}
	}
}

// The pretty timestamp used 15:05:05, which prints the second twice and never the
// minute.
func TestPrettyTimestampLayout(t *testing.T) {
	buf := new(bytes.Buffer)
	h := prettyHandlerOptions{SlogOpts: &slog.HandlerOptions{Level: LevelDebug}}.newPrettyHandler(buf)
	at := time.Date(2026, time.August, 30, 15, 4, 5, 0, time.UTC)

	if err := h.Handle(context.Background(), slog.NewRecord(at, LevelInfo, "msg", 0)); err != nil {
		t.Fatal(err)
	}

	if out := buf.String(); !strings.Contains(out, "[15:04:05.000]") {
		t.Fatalf("want [15:04:05.000] in %q", out)
	}
}

// The source must name the caller of the exported function, not this package.
func TestSourceNamesTheCaller(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts []LoggerOpt
	}{
		{"json", nil},
		{"pretty", []LoggerOpt{Local()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, buf := jsonCtx(t, append(tc.opts, WithSource())...)

			Info(ctx, "who is the caller?")

			out := buf.String()
			if !strings.Contains(out, "logger_test.go") {
				t.Fatalf("source does not name the caller: %s", out)
			}

			if strings.Contains(out, "logger.go") {
				t.Fatalf("source names this package instead of the caller: %s", out)
			}
		})
	}
}

// WithSource was accepted and then ignored by the JSON handler.
func TestWithSourceIsHonoured(t *testing.T) {
	ctx, buf := jsonCtx(t, WithSource())
	Info(ctx, "sourced")

	if out := buf.String(); !strings.Contains(out, `"source"`) {
		t.Fatalf("WithSource emitted no source field: %s", out)
	}

	off, offBuf := jsonCtx(t)
	Info(off, "not sourced")

	if out := offBuf.String(); strings.Contains(out, `"source"`) {
		t.Fatalf("source was emitted without WithSource: %s", out)
	}
}

func TestEnabled(t *testing.T) {
	ctx, _ := jsonCtx(t, WithLevel(LevelWarn))

	if Enabled(ctx, LevelInfo) {
		t.Error("info must not be enabled at warn level")
	}

	if !Enabled(ctx, LevelError) {
		t.Error("error must be enabled at warn level")
	}
}

func TestBuilderDefaultsToInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	ctx := NewLoggerContext(context.Background(), WithWriter(buf))

	Debug(ctx, "dropped")

	if buf.Len() != 0 {
		t.Fatalf("debug must be filtered at the default level, got %s", buf.String())
	}

	Info(ctx, "kept")

	if buf.Len() == 0 {
		t.Fatal("info must pass at the default level")
	}
}

// A *slog.LevelVar passed to WithLevel keeps the level adjustable after the logger
// is built.
func TestWithLevelVarIsAdjustable(t *testing.T) {
	buf := new(bytes.Buffer)
	lvl := new(slog.LevelVar)
	lvl.Set(LevelError)
	ctx := NewLoggerContext(context.Background(), WithWriter(buf), WithLevel(lvl))

	Info(ctx, "dropped")

	if buf.Len() != 0 {
		t.Fatalf("info must be filtered at error level, got %s", buf.String())
	}

	lvl.Set(LevelDebug)
	Info(ctx, "kept")

	if buf.Len() == 0 {
		t.Fatal("info must pass once the level var is lowered")
	}
}

func TestWithErrorWriterRoutesByLevel(t *testing.T) {
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	ctx := NewLoggerContext(context.Background(),
		WithWriter(out), WithErrorWriter(errOut), WithLevel(LevelDebug))

	Info(ctx, "ordinary")
	Error(ctx, "broken")

	if !strings.Contains(out.String(), "ordinary") || strings.Contains(out.String(), "broken") {
		t.Fatalf("stdout writer got the wrong records: %s", out.String())
	}

	if !strings.Contains(errOut.String(), "broken") || strings.Contains(errOut.String(), "ordinary") {
		t.Fatalf("error writer got the wrong records: %s", errOut.String())
	}
}

// SetLevel rebuilt the global logger over os.Stdout, silently discarding whatever
// SetGlobalLogger had installed.
func TestSetLevelKeepsTheInstalledLogger(t *testing.T) {
	restoreGlobal(t)

	buf := new(bytes.Buffer)
	lvl := new(slog.LevelVar)
	lvl.Set(LevelDebug)
	SetGlobalLogger(slog.New(newLoggerHandler(lvl, false, buf)))

	SetLevel(LevelWarn)
	Error(context.Background(), "does this reach my buffer?")

	if !strings.Contains(buf.String(), "does this reach my buffer?") {
		t.Fatalf("SetLevel redirected the installed logger, buffer holds %q", buf.String())
	}
}

func TestSetLevelFiltersTheDefaultLogger(t *testing.T) {
	restoreGlobal(t)

	buf := new(bytes.Buffer)
	globalLogger.Store(slog.New(newLoggerHandler(globalLevel, false, buf)))

	SetLevel(LevelError)
	Info(context.Background(), "dropped")

	if buf.Len() != 0 {
		t.Fatalf("SetLevel did not raise the level, got %s", buf.String())
	}

	SetLevel(LevelDebug)
	Info(context.Background(), "kept")

	if buf.Len() == 0 {
		t.Fatal("SetLevel did not lower the level")
	}
}

// Dependencies logging through slog.Default must reach the installed logger.
func TestSetGlobalLoggerBridgesSlogDefault(t *testing.T) {
	restoreGlobal(t)

	buf := new(bytes.Buffer)
	SetGlobalLogger(slog.New(newLoggerHandler(LevelDebug, false, buf)))

	slog.Info("from a dependency")

	if !strings.Contains(buf.String(), "from a dependency") {
		t.Fatalf("slog.Default was not bridged: %q", buf.String())
	}
}

// A nil logger in the context must fall back to the global logger.
func TestWithLoggerNilFallsBack(t *testing.T) {
	Info(WithLogger(context.Background(), nil), "hello")
}

func TestSetGlobalLoggerConcurrent(t *testing.T) {
	restoreGlobal(t)

	SetGlobalLogger(NewDiscardLogger())

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for range 100 {
			SetGlobalLogger(NewDiscardLogger())
		}
	}()

	go func() {
		defer wg.Done()

		for range 100 {
			Info(context.Background(), "concurrent")
		}
	}()

	wg.Wait()

	SetGlobalLogger(nil)

	if globalLogger.Load() == nil {
		t.Fatal("SetGlobalLogger(nil) must not clear the global logger")
	}
}

func restoreGlobal(t *testing.T) {
	t.Helper()

	logger, level, def := globalLogger.Load(), globalLevel.Level(), slog.Default()

	t.Cleanup(func() {
		globalLogger.Store(logger)
		globalLevel.Set(level)
		slog.SetDefault(def)
	})
}
