package aof

import (
	"errors"
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/mmap"
	"github.com/moontrade/kirana/pkg/pool"
	"github.com/moontrade/kirana/pkg/spinlock"
	"github.com/moontrade/kirana/pkg/timex"
	"github.com/moontrade/kirana/reactor"
	"golang.org/x/sys/cpu"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	ErrShrink         = errors.New("shrink prohibited")
	ErrIsDirectory    = errors.New("path is a directory")
	ErrCorrupted      = errors.New("corrupted")
	ErrEmptyFile      = errors.New("eof mode on empty file")
	ErrFileIsReadOnly = errors.New("file is read-only")
	ErrReadPermission = errors.New("file has no read permission")
)

type FileState int32

const (
	FileStateOpening FileState = 0
	FileStateOpened  FileState = 1
	FileStateEOF     FileState = 2
	FileStateClosing FileState = 3
	FileStateClosed  FileState = 4
)

func (s *FileState) load() FileState {
	return FileState(atomic.LoadInt32((*int32)(s)))
}

func (s *FileState) store(value FileState) {
	atomic.StoreInt32((*int32)(s), int32(value))
}

func (s *FileState) xchg(value FileState) FileState {
	return FileState(atomic.SwapInt32((*int32)(s), int32(value)))
	//return FileState(atomicx.Xchgint32((*int32)(s), int32(value)))
}

func (s *FileState) cas(old, new FileState) bool {
	return atomic.CompareAndSwapInt32((*int32)(s), int32(old), int32(new))
}

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

func hasReadPermission(perm os.FileMode) bool {
	// check for specific permission: user read
	return perm&0b100000000 == 0b100000000
}

func hasWritePermission(perm os.FileMode) bool {

	// check for specific permission: user write
	return perm&0b010000000 == 0b010000000
}

func hasReadWritePermission(perm os.FileMode) bool {
	return perm&0b110000000 == 0b110000000
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
	readOnly   bool
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
			if aof != nil {
				err = aof.err
			}
		}
		if err != nil {
			if aof != nil {
				aof.state.store(FileStateClosed)
				if aof.f != nil {
					_ = aof.f.Close()
				}
				m.gcList.Add(aof)
			}
			m.stats.OpenErrors.Incr()
			m.stats.OpenErrorsDur.Add(elapsed)
		} else {
			aof.state.cas(FileStateOpening, FileStateOpened)
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
	var (
		info os.FileInfo
		mode os.FileMode
	)
	info, aof.err = os.Stat(path)
	if aof.err != nil {
		if os.IsNotExist(aof.err) {
			if !geometry.Create {
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
			mode = info.Mode()
		} else {
			return nil, aof.err
		}
	} else {
		if info.IsDir() {
			aof.err = ErrIsDirectory
			return nil, ErrIsDirectory
		}
		aof.fileSize = info.Size()
		mode = info.Mode()
	}

	if aof.created {
		aof.readOnly = false
		aof.f, aof.err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, m.writeMode)
	} else {
		if hasReadWritePermission(mode.Perm()) {
			aof.readOnly = false
			aof.f, aof.err = os.OpenFile(path, os.O_RDWR, m.writeMode)
		} else {
			aof.readOnly = true
			if !hasReadPermission(mode.Perm()) {
				aof.err = ErrReadPermission
				return nil, ErrReadPermission
			}
			aof.state.store(FileStateEOF)
			aof.f, aof.err = os.OpenFile(path, os.O_RDONLY, m.readMode)
		}
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
		if recovery.Func == nil {
			return nil, ErrEmptyFile
		}
	}

	if aof.fileSize == 0 {
		aof.created = true
	}

	if aof.fileSize < geometry.SizeNow && aof.state != FileStateEOF {
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
	// Is the file done appending?
	if aof.state == FileStateEOF {
		// Map the entire size of the file.
		if aof.readOnly {
			data, aof.err = mmap.MapRegion(aof.f, int(aof.fileSize), mmap.RDONLY, 0, 0)
		} else {
			data, aof.err = mmap.MapRegion(aof.f, int(aof.fileSize), mmap.RDWR, 0, 0)
		}
	} else {
		// Map the max size the file can be.
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

	if !aof.created && !aof.readOnly && aof.recovery.Func != nil {
		result := aof.recovery.Func(aof.fileSize, data[0:aof.fileSize], aof.recovery.Magic)
		aof.recovery.result = result
		aof.err = result.Err
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

		switch result.Outcome {
		case Corrupted:
			aof.err = ErrCorrupted
			return aof, ErrCorrupted

		case Tail:
			aof.size = result.Tail - 8

		case Checkpoint:
			aof.size = result.Checkpoint + 8
			if result.Tail > result.Checkpoint+8 {
				// Warn
			}
			// Truncate to size if needed
			if aof.fileSize > aof.size {
				// Final size
				finalSize := aof.size
				if aof.recovery.Magic.Checkpoint > 0 {
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
				aof.state.store(FileStateEOF)
			}
		}
		aof.data = data
	} else {
		aof.data = data
	}

	return aof, nil
}

func (aof *AOF) destruct() {
	aof.state.cas(FileStateClosing, FileStateClosed)
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
	state := aof.state.load()
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
	if !aof.state.cas(state, FileStateClosing) {
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
	aof.state.store(FileStateOpening)
	aof.openWg.Add(1)
	aof.fileSize = 0
	aof.data = nil
	aof.name = name
	aof.gcIndex = -1
	aof.flushIndex = -1
	return aof
}
