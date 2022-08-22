package aof

import (
	"errors"
	"github.com/moontrade/wormhole/pkg/atomicx"
	"github.com/moontrade/wormhole/pkg/counter"
	"github.com/moontrade/wormhole/pkg/mmap"
	"github.com/moontrade/wormhole/pkg/pool"
	"github.com/moontrade/wormhole/pkg/spinlock"
	"github.com/moontrade/wormhole/pkg/timex"
	"github.com/moontrade/wormhole/reactor"
	"golang.org/x/sys/cpu"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	ErrShrink      = errors.New("shrink prohibited")
	ErrIsDirectory = errors.New("path is a directory")
	ErrCorrupted   = errors.New("corrupted")
	ErrEmptyFile   = errors.New("eof mode on empty file")
)

type FileState int32

const (
	FileStateOpening FileState = 0
	FileStateOpened  FileState = 1
	FileStateEOF     FileState = 2
	FileStateClosing FileState = 3
	FileStateClosed  FileState = 4
)

func (s *FileState) Load() FileState {
	return FileState(atomic.LoadInt32((*int32)(s)))
}

func (s *FileState) Store(value FileState) {
	atomic.StoreInt32((*int32)(s), int32(value))
}

func (s *FileState) Xchg(value FileState) FileState {
	return FileState(atomicx.Xchgint32((*int32)(s), int32(value)))
}

func (s *FileState) CAS(old, new FileState) bool {
	return atomicx.Casint32((*int32)(s), int32(old), int32(new))
}

const (
	// MagicTail Little-Endian = [170 36 117 84 99 156 155 65]
	// After each write the MagicTail is appended to the end.
	MagicTail = uint64(4727544184288126122)
	// MagicEOF Little-Endian = [44 219 31 242 165 172 120 248]
	MagicEOF = uint64(17904250147343162156)
)

var (
	pageSize = int64(os.Getpagesize())
)

var (
	aofPool = pool.NewPool[AOF](pool.Config[AOF]{
		SizeClass:     int(unsafe.Sizeof(AOF{})),
		PageSize:      1024,
		PagesPerShard: 1024,
	})
)

type FileStats struct {
	blockingCount counter.Counter
	blockingDur   counter.TimeCounter
	truncCount    counter.Counter
	truncDur      counter.TimeCounter
	truncErrCount counter.Counter
	truncErrDur   counter.Counter
}

// AOF is a single-producer multiple-consumer memory-mapped
// append only file. The same mapping is shared among any number
// of consumers. Each instance has a single mmap for its lifetime.
// The underlying file is then truncated "extended" in increments.
// Writes are blocked when writing past the current file size
// and a truncation is in-progress. Reads never block.
//
// In order to minimize truncation blocking, the Manager can
// schedule AOF truncation to keep up with the writing pace.
type AOF struct {
	m          *Manager
	f          *os.File
	closed     int64
	name       string
	data       mmap.MMap
	geometry   Geometry
	recovery   Recovery
	truncMu    sync.Mutex
	tailers    reactor.TaskSet
	openWg     sync.WaitGroup
	writeMu    spinlock.Mutex
	truncStart int64
	err        error
	stats      FileStats
	_          cpu.CacheLinePad
	size       int64
	fileSize   int64
	flushSize  int64
	syncSize   int64
	flushIndex int
	gcIndex    int
	gc         bool
	state      FileState
	created    bool
}

func alignToPageSize(size int64) int64 {
	if size < pageSize {
		return pageSize
	}
	s := (size / pageSize) * pageSize
	if size%pageSize != 0 {
		s += pageSize
	}
	return s
}

func (aof *AOF) IsAnonymous() bool { return aof.f == nil }

func getGCIndex(aof *AOF) int { return aof.gcIndex }
func setGCIndex(aof *AOF, index int) {
	aof.gc = true
	aof.gcIndex = index
}

func getFlushIndex(aof *AOF) int { return aof.flushIndex }
func setFlushIndex(aof *AOF, index int) {
	aof.flushIndex = index
}

func (m *Manager) OpenAnonymous(name string, size int64) (*AOF, error) {
	return nil, nil
}

func (m *Manager) Open(name string, geometry Geometry, recovery Recovery) (aof *AOF, err error) {
	var init bool
	aof, init = m.files.GetOrCreate(name, m.createFile)
	if !init {
		if aof.err != nil {
			return nil, aof.err
		}
		if aof.data == nil {
			aof.openWg.Wait()
		}
		return aof, nil
	}

	if recovery.Func == nil {
		recovery = RecoveryDefault
	}
	aof.recovery = recovery

	begin := timex.NanoTime()
	defer func() {
		end := timex.NanoTime()
		elapsed := end - begin
		if err == nil {
			err = aof.err
		}
		if err != nil {
			aof.state.Store(FileStateClosed)
			if aof.f != nil {
				_ = aof.f.Close()
			}
			m.gcList.Add(aof)
			m.stats.OpenErrors.Incr()
			m.stats.OpenErrorsDur.Add(elapsed)
		} else {
			aof.state.CAS(FileStateOpening, FileStateOpened)
			m.stats.Opens.Incr()
			m.stats.OpensDur.Add(elapsed)
			m.stats.ActiveMaps.Incr()
			m.stats.ActiveFileSize.Add(aof.fileSize)
			m.stats.ActiveMappedMemory.Add(int64(len(aof.data)))
			m.stats.LifetimeMemory.Add(int64(len(aof.data)))
		}
	}()
	defer aof.openWg.Done()

	path := name
	if len(m.dir) > 0 {
		path = filepath.Join(m.dir, name)
	}
	geometry.Validate()
	aof.geometry = geometry
	var info os.FileInfo
	info, aof.err = os.Stat(path)
	if aof.err != nil {
		if os.IsNotExist(aof.err) {
			if !geometry.Create || recovery.EOF {
				return nil, os.ErrNotExist
			}
			aof.err = nil
			aof.created = true
		} else if os.IsExist(aof.err) {
			aof.err = nil
			aof.created = false
			if info.IsDir() {
				aof.err = ErrIsDirectory
				return nil, ErrIsDirectory
			}
			aof.fileSize = info.Size()
		} else {
			return nil, aof.err
		}
	} else {
		if info.IsDir() {
			aof.err = ErrIsDirectory
			return nil, ErrIsDirectory
		}
		aof.fileSize = info.Size()
	}

	if aof.created {
		aof.f, aof.err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, m.perm)
	} else {
		aof.f, aof.err = os.OpenFile(path, os.O_RDWR, m.perm)
	}

	elapsed := timex.NanoTime() - begin
	if aof.err != nil {
		m.stats.OpenFileErrors.Incr()
		m.stats.OpenFileErrorsDur.Add(elapsed)
		if aof.f != nil {
			_ = aof.f.Close()
		}
		return nil, aof.err
	}
	m.stats.OpenFileCount.Incr()
	m.stats.OpenFileDur.Add(elapsed)

	if aof.created || aof.fileSize == 0 {
		if recovery.EOF {
			return nil, ErrEmptyFile
		}
	}

	if aof.fileSize == 0 {
		aof.created = true
	}

	if aof.fileSize < geometry.SizeNow {
		before := timex.NanoTime()
		aof.err = os.Truncate(path, geometry.SizeNow)
		elapsed := timex.NanoTime() - before
		m.stats.Truncates.Incr()
		m.stats.TruncatesDur.Add(elapsed)
		if aof.err != nil {
			m.stats.TruncateErrors.Incr()
			m.stats.TruncateErrorsDur.Add(elapsed)
			_ = aof.f.Close()
			return nil, aof.err
		}
		aof.fileSize = geometry.SizeNow
	}

	var data mmap.MMap
	before := timex.NanoTime()
	if recovery.EOF {
		data, aof.err = mmap.MapRegion(aof.f, int(aof.fileSize), mmap.RDWR, 0, 0)
	} else {
		data, aof.err = mmap.MapRegion(aof.f, int(geometry.SizeUpper), mmap.RDWR, 0, 0)
	}
	after := timex.NanoTime()
	elapsed = after - before

	if aof.err != nil {
		m.stats.MapErrors.Incr()
		m.stats.MapErrorsDur.Add(elapsed)
		_ = aof.f.Close()
		return nil, aof.err
	}
	m.stats.Maps.Incr()
	m.stats.MapsDur.Add(elapsed)

	if !aof.created {
		aof.err = aof.recovery.Do(aof.fileSize, data[0:aof.fileSize])
		if aof.err != nil {
			mapped := int64(len(data))
			begin = timex.NanoTime()
			err = data.Unmap()
			elapsed = timex.NanoTime() - begin
			m.stats.Unmaps.Incr()
			m.stats.UnmapsDur.Add(elapsed)
			m.stats.ActiveMaps.Decr()
			m.stats.ActiveMappedMemory.Add(-mapped)
			if err != nil {
				m.stats.UnmapErrors.Incr()
				m.stats.UnmapErrorsDur.Add(elapsed)
			}
			return aof, err
		}
		aof.size = aof.recovery.tail

		switch aof.recovery.result {
		case RecoveryResultCorrupted:
			aof.err = ErrCorrupted
			return aof, ErrCorrupted

		case RecoveryResultTail:

		case RecoveryResultEOF:
			// Truncate to size if needed
			if aof.fileSize > aof.size {
				// Final size
				finalSize := aof.size
				if aof.recovery.Magic.EOF > 0 {
					finalSize += 8
				}
				if finalSize != aof.fileSize {
					before := timex.NanoTime()
					aof.err = os.Truncate(path, finalSize)
					elapsed := timex.NanoTime() - before
					aof.m.stats.Truncates.Incr()
					aof.m.stats.TruncatesDur.Add(elapsed)
					if aof.err != nil {
						aof.m.stats.TruncateErrors.Incr()
						aof.m.stats.TruncateErrorsDur.Add(elapsed)
					}

					// Remap to size?
					mapSize := alignToPageSize(finalSize)
					if mapSize > int64(len(data)) {
						_ = data.Unmap()
						data = nil
						data, aof.err = mmap.MapRegion(aof.f, int(mapSize), mmap.RDONLY, 0, 0)
						if aof.err != nil {
							return aof, aof.err
						}
					}
				}
			}

			if aof.err == nil {
				aof.state.Store(FileStateEOF)
			}
		}
	}
	aof.data = data
	return aof, nil
}

func (aof *AOF) destruct() {
	aof.state.CAS(FileStateClosing, FileStateClosed)
	var err error
	data := aof.data
	// Unmap
	if len(data) > 0 {
		aof.data = nil
		mapped := int64(len(data))
		begin := timex.NanoTime()
		err = data.Unmap()
		elapsed := timex.NanoTime() - begin
		aof.m.stats.Unmaps.Incr()
		aof.m.stats.UnmapsDur.Add(elapsed)
		aof.m.stats.ActiveMaps.Decr()
		aof.m.stats.ActiveMappedMemory.Add(-mapped)
		if err != nil {
			aof.m.stats.UnmapErrors.Incr()
			aof.m.stats.UnmapErrorsDur.Add(elapsed)
		}
	}
	// Close file handle
	f := aof.f
	if f != nil {
		aof.f = nil
		_ = f.Close()
	}
}

func (aof *AOF) tryGC() {
	if !aof.tailers.IsEmpty() {
		return
	}
	_ = aof.tailers.Wake()
	aof.writeMu.Lock()
	defer aof.writeMu.Unlock()
	aof.destruct()
}

func (aof *AOF) Close() error {
	state := aof.state.Load()
	switch state {
	case FileStateClosed:
		return os.ErrClosed
	case FileStateClosing:
		return os.ErrClosed
	case FileStateOpening:
		return nil
	case FileStateOpened, FileStateEOF:
	}

	aof.writeMu.Lock()
	defer aof.writeMu.Unlock()

	// Move into closing state
	if !aof.state.CAS(state, FileStateClosing) {
		return nil
	}

	// Remove from files map
	aof.m.files.Delete(aof.name)

	// Prevent new tailers
	_ = aof.tailers.Close()

	// Can safely destruct?
	if aof.tailers.IsEmpty() {
		aof.destruct()
	} else {
		aof.m.gcList.Add(aof)
		_ = aof.tailers.Wake()
	}
	return nil
}

func (m *Manager) deleteFile(f *AOF) {

}

func (m *Manager) createFile(name string) *AOF {
	aof := aofPool.Get()
	*aof = AOF{}
	aof.m = m
	aof.state.Store(FileStateOpening)
	aof.openWg.Add(1)
	aof.fileSize = 0
	aof.data = nil
	aof.name = name
	aof.gcIndex = -1
	aof.flushIndex = -1
	return aof
}
