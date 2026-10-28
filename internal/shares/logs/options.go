package logs

import "github.com/rs/zerolog"

// options are shared by every adapter so they stay interchangeable.
type options struct {
	lvl       zerolog.Level
	caller    bool
	stack     bool
	timestamp bool
}

type Opt func(*options)

func newOptions(opts []Opt) options {
	o := options{lvl: zerolog.InfoLevel}
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func WithCaller() Opt {
	return func(o *options) {
		o.caller = true
	}
}

func WithStack() Opt {
	return func(o *options) {
		o.stack = true
	}
}

func WithTimestamp() Opt {
	return func(o *options) {
		o.timestamp = true
	}
}

func WithLevel(lvl zerolog.Level) Opt {
	return func(o *options) {
		o.lvl = lvl
	}
}
