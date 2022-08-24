package aof

import (
	"encoding/binary"
	"errors"
	"github.com/moontrade/wormhole/pkg/util"
	"unsafe"
)

const (
	// MagicTail Little-Endian = [170 36 117 84 99 156 155 65]
	// After each write the MagicTail is appended to the end.
	MagicTail = uint64(4727544184288126122)
	// MagicCheckpoint Little-Endian = [44 219 31 242 165 172 120 248]
	MagicCheckpoint = uint64(17904250147343162156)
)

var (
	RecoveryDefault = Recovery{
		Magic: Magic{
			Tail:       MagicTail,
			Checkpoint: MagicCheckpoint,
		},
		Func: RecoverWithMagic,
	}
	RecoveryReadOnly = Recovery{}
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
}

type RecoveryResult struct {
	Magic      Magic
	Outcome    RecoveryKind
	FileSize   int64
	Checkpoint int64
	Tail       int64
	Err        error
}

type RecoveryKind int

const (
	Empty      RecoveryKind = 0 // The Magic value was found at the tail
	Corrupted  RecoveryKind = 1 // The Magic value was found at the tail
	Tail       RecoveryKind = 2 // The Magic value was found at the tail
	Checkpoint RecoveryKind = 3 // The Magic Checkpoint value was found at the tail
	Panic      RecoveryKind = 4 // The Magic Checkpoint value was found at the tail
)

type RecoveryFunc func(
	fileSize int64,
	data []byte,
	magic Magic,
) (result RecoveryResult)

// Magic provides magic numbers for Tail and Checkpoint.
type Magic struct {
	// Tail is the magic number that marks the tail of the file
	Tail uint64
	// Checkpoint is the magic number that marks the end of a chunk.
	// During recovery if the Magic Tail is not found, it will search
	// for the last Checkpoint
	Checkpoint uint64
}

func (m *Magic) IsDisabled() bool {
	return m.Tail == 0 || m.Checkpoint == 0
}

func (m *Magic) IsEnabled() bool {
	return m.Tail != 0 && m.Checkpoint != 0
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

// RecoverWithMagic finds the last magic tail or last checkpoint
func RecoverWithMagic(fileSize int64, data []byte, magic Magic) (result RecoveryResult) {
	defer func() {
		if e := recover(); e != nil {
			result.Err = util.PanicToError(e)
			result.Outcome = Panic
		}
	}()
	result.Magic = magic
	result.FileSize = fileSize
	if fileSize > int64(len(data)) {
		result.Outcome = Corrupted
		result.Err = errors.New("fileSize is greater than mapping")
		return
	}
	if fileSize < 8 {
		if len(data) == 0 {
			result.Outcome = Tail
			result.Tail = 0
			return
		}
		for i := len(data) - 1; i > -1; i-- {
			if data[i] != 0 {
				result.Outcome = Tail
				result.Tail = int64(i + 1)
				return
			}
		}
		result.Outcome = Empty
		result.Tail = 0
		return
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
						result.Outcome = Tail
						result.Tail = start

						if magic.Checkpoint != 0 {
							magicCheckpointLast := lastByteUint64LE(magic.Checkpoint)
							// Search for last checkpoint.
							last--
							for last > 7 {
								if data[last] != magicCheckpointLast {
									last--
									continue
								}
								d = binary.LittleEndian.Uint64(data[last-7:])
								if d == magic.Checkpoint {
									result.Checkpoint = last - 7
									return
								}
								last--
							}
						}

						return
					}
					if d == magic.Checkpoint {
						result.Outcome = Checkpoint
						result.Tail = start + 8
						result.Checkpoint = start
						return
					}
					tail := last

					if magic.Checkpoint != 0 {
						magicCheckpointLast := lastByteUint64LE(magic.Checkpoint)
						// Search for last checkpoint.
						last--
						for last > 7 {
							if data[last] != magicCheckpointLast {
								last--
								continue
							}
							d = binary.LittleEndian.Uint64(data[last-7:])
							if d == magic.Checkpoint {
								result.Outcome = Checkpoint
								result.Checkpoint = last - 7
								result.Tail = last + 1
								return
							}
							last--
						}
					}

					result.Outcome = Corrupted
					result.Tail = tail + 1
					return
				} else {
					result.Outcome = Corrupted
					result.Tail = last + 1
					return
				}
			}
		}
		result.Outcome = Corrupted
		result.Tail = start + 8
		return
	}
	// File is empty
	result.Outcome = Empty
	result.Tail = 0
	return
}
