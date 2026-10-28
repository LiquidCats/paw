package logs_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/shares/logs"
)

// build returns the adapter as the port. noinline: otherwise the compiler
// devirtualizes the calls and hides the variadic-slice escape real callers pay.
//
//go:noinline
func build[L ports.Logger](newL constructor[L], opts ...logs.Opt) ports.Logger {
	return newL(io.Discard, opts...)
}

func benchInfo[L ports.Logger](b *testing.B, newL constructor[L], opts ...logs.Opt) {
	l := build(newL, opts...)
	b.ReportAllocs()
	for b.Loop() {
		l.Info("msg",
			ports.Str("s", "value"), ports.Int("i", 42), ports.Float64("f", 3.14),
			ports.Bool("b", true), ports.Dur("d", time.Second))
	}
}

func benchDisabled[L ports.Logger](b *testing.B, newL constructor[L]) {
	l := build(newL)
	b.ReportAllocs()
	for b.Loop() {
		l.Debug("msg", ports.Str("s", "value"), ports.Int("i", 42))
	}
}

func benchError[L ports.Logger](b *testing.B, newL constructor[L]) {
	l := build(newL)
	err := eris.Wrap(eris.New("boom"), "wrapped")
	b.ReportAllocs()
	for b.Loop() {
		l.Error("msg", ports.Err(err))
	}
}

func benchWith[L ports.Logger](b *testing.B, newL constructor[L]) {
	l := build(newL)
	b.ReportAllocs()
	for b.Loop() {
		_ = l.With(ports.Str("mod", "x"), ports.Int("n", 1))
	}
}

func benchParallel[L ports.Logger](b *testing.B, newL constructor[L]) {
	l := build(newL)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("msg", ports.Str("s", "value"), ports.Int("i", 42), ports.Bool("b", true))
		}
	})
}

func BenchmarkInfo(b *testing.B) {
	b.Run("zerolog", func(b *testing.B) { benchInfo(b, logs.NewZerologAdapter, logs.WithTimestamp()) })
	b.Run("slog", func(b *testing.B) { benchInfo(b, logs.NewSlogAdapter, logs.WithTimestamp()) })
	b.Run("zerolog-caller", func(b *testing.B) { benchInfo(b, logs.NewZerologAdapter, logs.WithCaller()) })
	b.Run("slog-caller", func(b *testing.B) { benchInfo(b, logs.NewSlogAdapter, logs.WithCaller()) })
	b.Run("raw-zerolog", func(b *testing.B) {
		l := zerolog.New(io.Discard).With().Timestamp().Logger()
		b.ReportAllocs()
		for b.Loop() {
			l.Info().Str("s", "value").Int("i", 42).Float64("f", 3.14).Bool("b", true).Dur("d", time.Second).Msg("msg")
		}
	})
	b.Run("raw-slog", func(b *testing.B) {
		l := slog.New(slog.NewJSONHandler(io.Discard, nil)) //nolint:sloglint
		b.ReportAllocs()
		for b.Loop() {
			l.LogAttrs(b.Context(), slog.LevelInfo, "msg",
				slog.String("s", "value"), slog.Int("i", 42), slog.Float64("f", 3.14),
				slog.Bool("b", true), slog.Duration("d", time.Second))
		}
	})
}

func BenchmarkDisabled(b *testing.B) {
	b.Run("zerolog", func(b *testing.B) { benchDisabled(b, logs.NewZerologAdapter) })
	b.Run("slog", func(b *testing.B) { benchDisabled(b, logs.NewSlogAdapter) })
}

func BenchmarkError(b *testing.B) {
	b.Run("zerolog", func(b *testing.B) { benchError(b, logs.NewZerologAdapter) })
	b.Run("slog", func(b *testing.B) { benchError(b, logs.NewSlogAdapter) })
}

func BenchmarkWith(b *testing.B) {
	b.Run("zerolog", func(b *testing.B) { benchWith(b, logs.NewZerologAdapter) })
	b.Run("slog", func(b *testing.B) { benchWith(b, logs.NewSlogAdapter) })
}

func BenchmarkParallel(b *testing.B) {
	b.Run("zerolog", func(b *testing.B) { benchParallel(b, logs.NewZerologAdapter) })
	b.Run("slog", func(b *testing.B) { benchParallel(b, logs.NewSlogAdapter) })
}
