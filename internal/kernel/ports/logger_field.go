package ports

import (
	"math"
	"time"
)

// Kind tells adapters which Field accessor holds the value.
type Kind uint8

const (
	KindAny Kind = iota
	KindStr
	KindInt
	KindUint
	KindFloat
	KindBool
	KindDur
	KindErr
)

// Field is a log key/value passed by value, so a call through [Logger] costs
// one variadic-slice allocation instead of one per field. Scalars live inline
// (num/str) and are never boxed into any.
type Field struct {
	key  string
	kind Kind
	num  uint64
	str  string
	val  any
}

// Scalar are the types Of stores inline.
type Scalar interface {
	string | int | int64 | uint64 | float64 | bool | time.Duration
}

// Of builds a typed Field; the type switch is resolved per instantiation.
// Prefer it over Any for scalars.
func Of[T Scalar](key string, v T) Field {
	switch x := any(v).(type) {
	case string:
		return Str(key, x)
	case int:
		return Int(key, x)
	case int64:
		return Int64(key, x)
	case uint64:
		return Uint64(key, x)
	case float64:
		return Float64(key, x)
	case bool:
		return Bool(key, x)
	case time.Duration:
		return Dur(key, x)
	}

	panic("unreachable: Scalar is closed")
}

func Any(key string, v any) Field { return Field{key: key, kind: KindAny, val: v} }
func Err(err error) Field         { return Field{key: "err", kind: KindErr, val: err} }
func Str(key, v string) Field     { return Field{key: key, kind: KindStr, str: v} }
func Int(key string, v int) Field { return Field{key: key, kind: KindInt, num: uint64(int64(v))} }
func Int64(key string, v int64) Field {
	return Field{key: key, kind: KindInt, num: uint64(v)}
}
func Uint64(key string, v uint64) Field { return Field{key: key, kind: KindUint, num: v} }
func Float64(key string, v float64) Field {
	return Field{key: key, kind: KindFloat, num: math.Float64bits(v)}
}
func Dur(key string, v time.Duration) Field {
	return Field{key: key, kind: KindDur, num: uint64(v)}
}
func Bool(key string, v bool) Field {
	f := Field{key: key, kind: KindBool}
	if v {
		f.num = 1
	}

	return f
}

func (f Field) Key() string        { return f.key }
func (f Field) Kind() Kind         { return f.kind }
func (f Field) Str() string        { return f.str }
func (f Field) Int() int64         { return int64(f.num) }
func (f Field) Uint() uint64       { return f.num }
func (f Field) Float() float64     { return math.Float64frombits(f.num) }
func (f Field) Bool() bool         { return f.num != 0 }
func (f Field) Dur() time.Duration { return time.Duration(f.num) }

// Value boxes the field's value; adapters should use the typed accessors.
func (f Field) Value() any {
	switch f.kind {
	case KindStr:
		return f.str
	case KindInt:
		return f.Int()
	case KindUint:
		return f.num
	case KindFloat:
		return f.Float()
	case KindBool:
		return f.Bool()
	case KindDur:
		return f.Dur()
	case KindAny, KindErr:
		return f.val
	}

	return f.val
}
