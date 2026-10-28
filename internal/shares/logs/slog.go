package logs

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/kernel/ports"
)

var _ ports.Logger = (*SlogAdapter)(nil)

// levelPanic sits above slog.LevelError; slog has no panic level.
const levelPanic = slog.LevelError + 4

// SlogAdapter is a drop-in alternative to [ZerologAdapter] on top of log/slog.
type SlogAdapter struct {
	logger *slog.Logger
	caller bool
	stack  bool
	// stamp: a zero Record.Time makes handlers omit the time, which is cheaper
	// than ReplaceAttr (that forces slog's slow reflect path).
	stamp bool
}

// NewSlogAdapter mirrors [NewZerologAdapter]: same options, JSON output.
func NewSlogAdapter(w io.Writer, opts ...Opt) *SlogAdapter {
	o := newOptions(opts)

	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: o.caller,
		// zerolog: debug=0 info=1 .. panic=5; slog: debug=-4 info=0 .. error=8.
		Level: slog.Level((o.lvl - zerolog.InfoLevel) * 4), //nolint:mnd
	})

	return FromSlog(slog.New(h), opts...)
}

// FromSlog wraps an existing logger, keeping its handler. Level, timestamp and
// source output are then up to that handler; only WithCaller (record source
// capture), WithTimestamp and WithStack apply.
func FromSlog(l *slog.Logger, opts ...Opt) *SlogAdapter {
	o := newOptions(opts)

	return &SlogAdapter{logger: l, caller: o.caller, stack: o.stack, stamp: o.timestamp}
}

func (s *SlogAdapter) attr(f *ports.Field) slog.Attr {
	switch f.Kind() {
	case ports.KindStr:
		return slog.String(f.Key(), f.Str())
	case ports.KindInt:
		return slog.Int64(f.Key(), f.Int())
	case ports.KindUint:
		return slog.Uint64(f.Key(), f.Uint())
	case ports.KindFloat:
		return slog.Float64(f.Key(), f.Float())
	case ports.KindBool:
		return slog.Bool(f.Key(), f.Bool())
	case ports.KindDur:
		return slog.Duration(f.Key(), f.Dur())
	case ports.KindErr, ports.KindAny:
	}

	if err, ok := f.Value().(error); ok {
		// eris trace formatting is costly, opt-in via WithStack.
		if s.stack {
			return slog.Any(f.Key(), eris.ToJSON(err, true))
		}

		return slog.String(f.Key(), err.Error())
	}

	return slog.Any(f.Key(), f.Value())
}

// log is always called straight from a level method, so the user frame is 3 up.
func (s *SlogAdapter) log(lvl slog.Level, msg string, fields []ports.Field) {
	ctx := context.Background()
	if !s.logger.Enabled(ctx, lvl) {
		return
	}

	var pc uintptr
	if s.caller {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:]) //nolint:mnd
		pc = pcs[0]
	}

	var now time.Time
	if s.stamp {
		now = time.Now()
	}

	r := slog.NewRecord(now, lvl, msg, pc)
	for i := range fields {
		r.AddAttrs(s.attr(&fields[i]))
	}

	_ = s.logger.Handler().Handle(ctx, r)
}

// Debug implements [ports.Logger].
func (s *SlogAdapter) Debug(msg string, fields ...ports.Field) {
	s.log(slog.LevelDebug, msg, fields)
}

// Info implements [ports.Logger].
func (s *SlogAdapter) Info(msg string, fields ...ports.Field) {
	s.log(slog.LevelInfo, msg, fields)
}

// Warn implements [ports.Logger].
func (s *SlogAdapter) Warn(msg string, fields ...ports.Field) {
	s.log(slog.LevelWarn, msg, fields)
}

// Error implements [ports.Logger].
func (s *SlogAdapter) Error(msg string, fields ...ports.Field) {
	s.log(slog.LevelError, msg, fields)
}

// Panic implements [ports.Logger]: logs, then panics with msg like zerolog.
func (s *SlogAdapter) Panic(msg string, fields ...ports.Field) {
	s.log(levelPanic, msg, fields)
	panic(msg)
}

// With implements [ports.Logger].
func (s *SlogAdapter) With(field ports.Field) ports.Logger {
	return &SlogAdapter{
		logger: s.logger.With(s.attr(&field)),
		caller: s.caller,
		stack:  s.stack,
		stamp:  s.stamp,
	}
}
