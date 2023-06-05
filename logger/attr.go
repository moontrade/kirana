// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"time"
)

// An Attr is a key-value pair.
type Attr struct {
	Key   string
	Value Value
}

// String returns an Attr for a string value.
func String(key, value string) Attr {
	return Attr{key, StringValue(value)}
}

// Int64 returns an Attr for an int64.
func Int64(key string, value int64) Attr {
	return Attr{key, Int64Value(value)}
}

// Int converts an int to an int64 and returns
// an Attr with that value.
func Int(key string, value int) Attr {
	return Int64(key, int64(value))
}

// Uint64 returns an Attr for a uint64.
func Uint64(key string, v uint64) Attr {
	return Attr{key, Uint64Value(v)}
}

// Float64 returns an Attr for a floating-point number.
func Float64(key string, v float64) Attr {
	return Attr{key, Float64Value(v)}
}

// Bool returns an Attr for a bool.
func Bool(key string, v bool) Attr {
	return Attr{key, BoolValue(v)}
}

// Time returns an Attr for a time.Time.
// It discards the monotonic portion.
func Time(key string, v time.Time) Attr {
	return Attr{key, TimeValue(v)}
}

// Duration returns an Attr for a time.Duration.
func Duration(key string, v time.Duration) Attr {
	return Attr{key, DurationValue(v)}
}

// Group returns an Attr for a Group Value.
// The first argument is the key; the remaining arguments
// are converted to Attrs as in [Logger.Log].
//
// Use Group to collect several key-value pairs under a single
// key on a log line, or as the result of LogValue
// in order to log a single value as multiple Attrs.
func Group(key string, args ...any) Attr {
	return Attr{key, GroupValue(argsToAttrSlice(args)...)}
}

func argsToAttrSlice(args []any) []Attr {
	var (
		attr  Attr
		attrs []Attr
	)
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}

const badKey = "!BADKEY"

// argsToAttr turns a prefix of the nonempty args slice into an Attr
// and returns the unconsumed portion of the slice.
// If args[0] is an Attr, it returns it.
// If args[0] is a string, it treats the first two elements as
// a key-value pair.
// Otherwise, it treats args[0] as a value with a missing key.
func argsToAttr(args []any) (Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return String(badKey, x), nil
		}
		return Any(x, args[1]), args[2:]

	case Attr:
		return x, args[1:]

	default:
		return Any(badKey, x), args[1:]
	}
}

// Any returns an Attr for the supplied value.
// See [Value.AnyValue] for how values are treated.
func Any(key string, value any) Attr {
	return Attr{key, AnyValue(value)}
}

// Equal reports whether a and b have equal keys and values.
func (a Attr) Equal(b Attr) bool {
	return a.Key == b.Key && a.Value.Equal(b.Value)
}

func (a Attr) String() string {
	return fmt.Sprintf("%s=%s", a.Key, a.Value)
}

// isEmpty reports whether a has an empty key and a nil value.
// That can be written as Attr{} or Any("", nil).
func (a Attr) isEmpty() bool {
	return a.Key == "" && a.Value.num == 0 && a.Value.any == nil
}