package aof

import (
	"encoding/binary"
	"errors"
	"unsafe"
)

var (
	RecoveryDefault = Recovery{
		Magic: Magic{
			Tail: MagicTail,
			EOF:  MagicEOF,
		},
		Func: RecoverWithMagic,
	}
	AlwaysEOF = Recovery{
		Func: func(fileSize int64, data []byte, magic Magic) (size int64, result RecoveryResult, err error) {
			return fileSize, RecoveryResultEOF, nil
		},
		EOF: true,
	}
)

// Recovery recovers an existing file on open by finding the tail. Since files are mapped
// and truncated according to the Geometry, it will be zero filled between logical tail
// and the file-system's tail. Furthermore, during crashes there is a potential that a
// partial write occurred. Using Magic numbers is the first line of defense to identify
// possible corruption. Since, this package does not define file formats, it is encouraged
// that the file format have the ability to recover to the last known commit point when
// the Magic number is not found.
type Recovery struct {
	Magic  Magic
	Func   RecoveryFunc
	tail   int64
	result RecoveryResult
	err    error
	EOF    bool
}

type RecoveryResult int

const (
	RecoveryResultEmpty     RecoveryResult = 0 // The Magic value was found at the tail
	RecoveryResultCorrupted RecoveryResult = 1 // The Magic value was found at the tail
	RecoveryResultTail      RecoveryResult = 2 // The Magic value was found at the tail
	RecoveryResultEOF       RecoveryResult = 3 // The Magic EOF value was found at the tail
)

type RecoveryFunc func(
	fileSize int64,
	data []byte,
	magic Magic,
) (size int64, result RecoveryResult, err error)

// Magic provides magic numbers for Tail and EOF.
type Magic struct {
	Tail uint64 // Tail is the magic number that marks the tail of the file
	EOF  uint64 // EOF is the magic number that marks the end of the file
}

func (m *Magic) IsDisabled() bool {
	return m.Tail == 0 || m.EOF == 0
}

func (m *Magic) IsEnabled() bool {
	return m.Tail != 0 && m.EOF != 0
}

func (r *Recovery) Result() RecoveryResult { return r.result }
func (r *Recovery) Tail() int64            { return r.tail }
func (r *Recovery) Err() error             { return r.err }
func (r *Recovery) Clone() Recovery {
	c := Recovery{}
	c.Magic = r.Magic
	c.Func = r.Func
	return c
}
func (r *Recovery) Do(fileSize int64, data []byte) error {
	r.tail, r.result, r.err = r.Func(fileSize, data, r.Magic)
	return r.err
}

// RecoverWithMagic finds the last magic tail or eof. Any other first value results in corruption result.
func RecoverWithMagic(fileSize int64, data []byte, magic Magic) (end int64, result RecoveryResult, err error) {
	if fileSize > int64(len(data)) {
		return 0, RecoveryResultCorrupted, errors.New("fileSize is greater than mapping")
	}
	if fileSize < 8 {
		if len(data) == 0 {
			return 0, RecoveryResultTail, nil
		}
		for i := len(data) - 1; i > -1; i-- {
			if data[i] != 0 {
				return int64(i + 1), RecoveryResultTail, nil
			}
		}
		return 0, RecoveryResultTail, nil
	}
	var d uint64
	for start := fileSize - 8; start >= 0; start -= 8 {
		d = *(*uint64)(unsafe.Pointer(&data[start]))
		if d == 0 {
			continue
		}
		// Find last non-zero byte
		for last := start + 7; last >= start; last-- {
			if data[last] != 0 {
				start = last - 7
				if start >= 0 {
					d = binary.LittleEndian.Uint64(data[start:])
					if d == magic.Tail {
						return start, RecoveryResultTail, nil
					}
					if d == magic.EOF {
						return start, RecoveryResultEOF, nil
					}
					return last + 1, RecoveryResultCorrupted, nil
				} else {
					return last + 1, RecoveryResultCorrupted, nil
				}
			}
		}
		return start + 8, RecoveryResultCorrupted, nil
	}
	// File is empty
	return 0, RecoveryResultEmpty, nil
}

func RecoverFirstNonZero(fileSize int64, data []byte, magic Magic) (size int64, result RecoveryResult, err error) {
	if fileSize > int64(len(data)) {
		return 0, RecoveryResultCorrupted, errors.New("fileSize is greater than mapping")
	}
	if fileSize < 8 {
		if len(data) == 0 {
			return 0, RecoveryResultTail, nil
		}
		for i := len(data) - 1; i > -1; i-- {
			if data[i] != 0 {
				return int64(i + 1), RecoveryResultTail, nil
			}
		}
		return 0, RecoveryResultTail, nil
	}
	for start := fileSize - 8; start >= 0; start -= 8 {
		// Speed up scanning by using a uint64
		if *(*uint64)(unsafe.Pointer(&data[start])) == 0 {
			continue
		}
		// Find last non-zero byte
		for i := int64(7); i > -1; i-- {
			if data[start+i] != 0 {
				return int64(start + i + 1), RecoveryResultTail, nil
			}
		}
		// No zero bytes in uint64
		return start + 8, RecoveryResultTail, nil
	}
	// File is empty
	return 0, RecoveryResultTail, nil
}
