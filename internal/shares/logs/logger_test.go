package logs_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/shares/logs"
)

// constructor is satisfied by every adapter, which is what makes them swappable.
type constructor[L ports.Logger] func(io.Writer, ...logs.Opt) L

func assertContains(t *testing.T, out string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
}

// conformance runs the same behaviour checks against any adapter. durNanos is
// how that backend renders one second (zerolog: ms, slog: ns).
func conformance[L ports.Logger](t *testing.T, newL constructor[L], durNanos string) {
	t.Helper()

	t.Run("level, with, caller", func(t *testing.T) {
		var buf bytes.Buffer
		var l ports.Logger = newL(&buf, logs.WithCaller(), logs.WithLevel(zerolog.InfoLevel))

		l.Debug("hidden", ports.Any("a", 1))
		if buf.Len() != 0 {
			t.Fatalf("debug leaked: %s", buf.String())
		}

		l.With(ports.Str("mod", "x")).Info("hi", ports.Any("n", 1), ports.Err(eris.New("boom")))
		assertContains(t, buf.String(), `"mod":"x"`, `"n":1`, `"err":"boom"`, `logger_test.go`)
	})

	t.Run("typed fields", func(t *testing.T) {
		var buf bytes.Buffer
		var l ports.Logger = newL(&buf)

		l.Info("typed",
			ports.Str("s", "v"), ports.Int("i", -5), ports.Int64("i64", 1<<40),
			ports.Uint64("u", 7), ports.Float64("f", 1.5), ports.Bool("t", true),
			ports.Bool("f2", false), ports.Dur("d", time.Second))
		assertContains(t, buf.String(),
			`"s":"v"`, `"i":-5`, `"i64":1099511627776`, `"u":7`, `"f":1.5`, `"t":true`, `"f2":false`, `"d":`+durNanos)
	})

	t.Run("of", func(t *testing.T) {
		var buf bytes.Buffer
		var l ports.Logger = newL(&buf)

		l.Info("of",
			ports.Of("s", "v"), ports.Of("i", -5), ports.Of("i64", int64(9)), ports.Of("u", uint64(7)),
			ports.Of("f", 1.5), ports.Of("t", true), ports.Of("d", time.Second))
		assertContains(t, buf.String(), `"s":"v"`, `"i":-5`, `"i64":9`, `"u":7`, `"f":1.5`, `"t":true`, `"d":`+durNanos)
	})

	t.Run("stack keeps eris trace", func(t *testing.T) {
		var buf bytes.Buffer
		var l ports.Logger = newL(&buf, logs.WithStack())

		l.Error("e", ports.Err(eris.New("boom")))
		assertContains(t, buf.String(), `"err":{`, `"stack"`)
	})

	t.Run("timestamp is opt-in", func(t *testing.T) {
		var off, on bytes.Buffer
		var l ports.Logger = newL(&off)
		l.Info("m")
		l = newL(&on, logs.WithTimestamp())
		l.Info("m")

		if strings.Contains(off.String(), `"time"`) || !strings.Contains(on.String(), `"time"`) {
			t.Errorf("off=%s on=%s", off.String(), on.String())
		}
	})

	t.Run("panic logs then panics with msg", func(t *testing.T) {
		var buf bytes.Buffer
		var l ports.Logger = newL(&buf)

		defer func() {
			if r := recover(); r != "die" {
				t.Errorf("recovered %v", r)
			}
			assertContains(t, buf.String(), `"die"`)
		}()
		l.Panic("die")
	})
}

func TestConformance(t *testing.T) {
	t.Run("zerolog", func(t *testing.T) { conformance(t, logs.NewZerologAdapter, "1000") })
	t.Run("slog", func(t *testing.T) { conformance(t, logs.NewSlogAdapter, "1000000000") })
}
