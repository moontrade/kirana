package util

import (
	"errors"
	"fmt"
)

func PanicToError(e any) (err error) {
	switch v := e.(type) {
	case error:
		err = v
	case string:
		err = errors.New(v)
	case int:
		err = fmt.Errorf("panic code: %d", v)
	case int8:
		err = fmt.Errorf("panic code: %d", v)
	case int16:
		err = fmt.Errorf("panic code: %d", v)
	case int32:
		err = fmt.Errorf("panic code: %d", v)
	case int64:
		err = fmt.Errorf("panic code: %d", v)
	case uint:
		err = fmt.Errorf("panic code: %d", v)
	case uintptr:
		err = fmt.Errorf("panic uintptr: %d", v)
	case uint8:
		err = fmt.Errorf("panic code: %d", v)
	case uint16:
		err = fmt.Errorf("panic code: %d", v)
	case uint32:
		err = fmt.Errorf("panic code: %d", v)
	case uint64:
		err = fmt.Errorf("panic code: %d", v)
	case float32:
		err = fmt.Errorf("panic code: %f", v)
	case float64:
		err = fmt.Errorf("panic code: %f", v)
	case fmt.Stringer:
		err = errors.New(v.String())
	default:
		err = errors.New("panic")
	}
	return
}
