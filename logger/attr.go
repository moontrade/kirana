// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import "time"

// Kind is the kind of Value.
type Kind uint8

// The following list is sorted alphabetically, but it's also important that
// KindAny is 0 so that a zero Value represents nil.

const (
	KindUnknown Kind = iota
	KindBool
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindFloat32
	KindFloat64
	KindDuration
	KindTime
	KindString
)

type AttrType interface {
	string | int8 | int16 | int32 | int64 | uint8 |
		uint16 | uint32 | uint64 | float32 | float64 |
		time.Duration | time.Time
}

// An Attr is a key-value pair.
type Attr[T AttrType] struct {
	Key   string
	Value T
}

func (a *Attr[T]) Kind() Kind {
	switch ((any)(a.Value)).(type) {
	case string:
		return KindString
	case bool:
		return KindBool
	case int8:
		return KindInt8
	case int16:
		return KindInt16
	case int32:
		return KindInt32
	case int64:
		return KindInt64
	case uint8:
		return KindUint8
	case uint16:
		return KindUint16
	case uint32:
		return KindUint32
	case uint64:
		return KindUint64
	case float32:
		return KindFloat32
	case float64:
		return KindFloat64
	case time.Duration:
		return KindDuration
	case time.Time:
		return KindTime
	default:
		return KindUnknown
	}
}
