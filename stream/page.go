package stream

import "unsafe"

const (
	// MagicTail Little-Endian = [170 36 117 84 99 156 155 65]
	// After each write the MagicTail is appended to the end.
	MagicTail = uint64(4727544184288126122)
	// MagicCheckpoint Little-Endian = [44 219 31 242 165 172 120 248]
	MagicCheckpoint = uint64(17904250147343162156)

	// [186 134 103 188 127 178 15 80]
	MagicA = uint64(5769025909376386746)
	// [183 246 100 51 172 27 206 147]
	MagicB = 10650480595188381367
	// [154 155 31 68 148 78 128 217]
	MagicC = 15672613101954374554
	// [244 155 201 228 184 125 35 250]
	MagicD = 18024388366732729332
	// [16 57 249 137 219 49 35 55]
	MagicE = 3973074115253319952
)

type PageSize int32

const (
	Page1KB   int32 = 1024
	Page2KB   int32 = 1024 * 2
	Page4KB   int32 = 1024 * 4
	Page8KB   int32 = 1024 * 8
	Page16KB  int32 = 1024 * 16
	Page32KB  int32 = 1024 * 32
	Page64KB  int32 = 1024 * 64
	Page128KB int32 = 1024 * 128
	Page256KB int32 = 1024 * 256
	Page512KB int32 = 1024 * 512
	Page1MB   int32 = 1024 * 1024
	Page2MB   int32 = 1024 * 1024 * 2
)

type PageHeader struct {
	Magic    uint64
	StreamID int64
	Time     int64
	Head     int64
	Size     PageSize
}

type PageTail struct {
	End        int64
	LastID     int64
	LastOffset int32
	Count      int32
	Size       int32
	Magic      uint64
}

type Page struct {
	header pageHeaderPtr
	tail   pageTailPtr
}

type deref[T any] uintptr

func (d deref[T]) Deref() *T {
	return (*T)(unsafe.Pointer(uintptr(d)))
}

type pageHeaderPtr uintptr

func (p pageHeaderPtr) Deref() *PageHeader {
	return (*PageHeader)(unsafe.Pointer(p))
}

type pageTailPtr uintptr

func (p pageTailPtr) Deref() *PageTail {
	return (*PageTail)(unsafe.Pointer(p))
}

type RecordHeader struct {
	ID   int64
	Size uint32
	Seq  uint32
}
