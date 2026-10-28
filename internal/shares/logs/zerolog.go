package logs

import (
	"context"
	"io"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"

	"github.com/LiquidCats/paw/internal/kernel/ports"
)

var _ ports.Logger = (*ZerologAdapter)(nil)

type ZerologAdapter struct {
	logger zerolog.Logger
	stack  bool
}

func NewZerologAdapter(w io.Writer, opts ...Opt) *ZerologAdapter {
	o := newOptions(opts)

	ctx := zerolog.New(w).Level(o.lvl).With()
	if o.caller {
		// +1 frame: the adapter method sits between the caller and zerolog.
		ctx = ctx.CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + 1)
	}
	if o.stack {
		ctx = ctx.Stack()
	}
	if o.timestamp {
		ctx = ctx.Timestamp()
	}

	logger := ctx.Logger()

	zerolog.DefaultContextLogger = &logger //nolint:reassign

	return &ZerologAdapter{
		logger: logger,
		stack:  o.stack,
	}
}

func (z *ZerologAdapter) WithContext(ctx context.Context) context.Context {
	return z.logger.WithContext(ctx)
}

// fields appends args to ev using typed zerolog setters, avoiding reflection
// for common types. Callers must have checked ev.Enabled().
func (z *ZerologAdapter) fields(ev *zerolog.Event, fields []ports.Field) *zerolog.Event {
	for i := range fields {
		f := &fields[i]
		switch f.Kind() {
		case ports.KindStr:
			ev = ev.Str(f.Key(), f.Str())
		case ports.KindInt:
			ev = ev.Int64(f.Key(), f.Int())
		case ports.KindUint:
			ev = ev.Uint64(f.Key(), f.Uint())
		case ports.KindFloat:
			ev = ev.Float64(f.Key(), f.Float())
		case ports.KindBool:
			ev = ev.Bool(f.Key(), f.Bool())
		case ports.KindDur:
			ev = ev.Dur(f.Key(), f.Dur())
		case ports.KindErr, ports.KindAny:
			ev = z.any(ev, f.Key(), f.Value())
		}
	}

	return ev
}

// any handles boxed values; common types still avoid reflection.
func (z *ZerologAdapter) any(ev *zerolog.Event, key string, val any) *zerolog.Event {
	switch v := val.(type) {
	case string:
		return ev.Str(key, v)
	case int:
		return ev.Int(key, v)
	case int64:
		return ev.Int64(key, v)
	case uint64:
		return ev.Uint64(key, v)
	case float64:
		return ev.Float64(key, v)
	case bool:
		return ev.Bool(key, v)
	case time.Duration:
		return ev.Dur(key, v)
	case time.Time:
		return ev.Time(key, v)
	case error:
		// eris trace formatting is ~3x the cost of the log call, opt-in via WithStack.
		if z.stack {
			return ev.Any(key, eris.ToJSON(v, true))
		}

		return ev.Str(key, v.Error())
	default:
		return ev.Any(key, v)
	}
}

// Debug implements [ports.Logger].
func (z *ZerologAdapter) Debug(msg string, fields ...ports.Field) {
	if ev := z.logger.Debug(); ev.Enabled() {
		z.fields(ev, fields).Msg(msg)
	}
}

// Error implements [ports.Logger].
func (z *ZerologAdapter) Error(msg string, fields ...ports.Field) {
	if ev := z.logger.Error(); ev.Enabled() {
		z.fields(ev, fields).Msg(msg)
	}
}

// Info implements [ports.Logger].
func (z *ZerologAdapter) Info(msg string, fields ...ports.Field) {
	if ev := z.logger.Info(); ev.Enabled() {
		z.fields(ev, fields).Msg(msg)
	}
}

// Panic implements [ports.Logger].
func (z *ZerologAdapter) Panic(msg string, fields ...ports.Field) {
	if ev := z.logger.Panic(); ev.Enabled() {
		z.fields(ev, fields).Msg(msg)
	}
}

// Warn implements [ports.Logger].
func (z *ZerologAdapter) Warn(msg string, fields ...ports.Field) {
	if ev := z.logger.Warn(); ev.Enabled() {
		z.fields(ev, fields).Msg(msg)
	}
}

// fieldObject lets With reuse the Event switch via Context.EmbedObject.
type fieldObject struct {
	z *ZerologAdapter
	f ports.Field
}

func (f fieldObject) MarshalZerologObject(e *zerolog.Event) {
	f.z.fields(e, []ports.Field{f.f})
}

// With implements [ports.Logger].
func (z *ZerologAdapter) With(field ports.Field) ports.Logger {
	return &ZerologAdapter{
		logger: z.logger.With().EmbedObject(fieldObject{z: z, f: field}).Logger(),
		stack:  z.stack,
	}
}
