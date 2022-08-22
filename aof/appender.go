package aof

import (
	"errors"
	logger "github.com/moontrade/log"
	"github.com/moontrade/wormhole/pkg/timex"
	"github.com/moontrade/wormhole/pkg/util"
	"io"
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"
)

var (
	ErrAppendFuncNil = errors.New("append func nil")
)

type AppendEvent struct {
	Begin int64
	End   int64
	Tail  []byte
	file  []byte
}

func (a *AppendEvent) File() []byte {
	return a.file
}

type AppendFunc func(event AppendEvent) (int64, error)

type ErrorFunc func(err error)

var (
	ErrWouldBlock = errors.New("would block")
)

func (aof *AOF) Finish() error {
	aof.writeMu.Lock()
	defer aof.writeMu.Unlock()
	f := aof.f
	if !aof.state.CAS(FileStateOpened, FileStateEOF) {
		return os.ErrClosed
	}
	if f != nil {
		aof.m.flushList.Remove(aof)

		begin := timex.NanoTime()
		aof.m.stats.Finishes.Incr()
		finalSize := aof.size
		// Replace MagicTail with MagicEOF
		if aof.recovery.Magic.EOF != 0 {
			write64LE(unsafe.Pointer(&aof.data[finalSize]), aof.recovery.Magic.EOF)
			finalSize += 8
		}
		// Truncate to size
		aof.err = syscall.Truncate(f.Name(), finalSize)
		if aof.err != nil {
			elapsed := timex.NanoTime() - begin
			aof.m.stats.FinishErrors.Incr()
			aof.m.stats.FinishErrorsDur.Add(elapsed)
			return aof.err
		}
		aof.err = aof.sync()
		elapsed := timex.NanoTime() - begin
		aof.m.stats.FinishesDur.Add(elapsed)
		if aof.err != nil {
			aof.m.stats.FinishErrors.Incr()
			aof.m.stats.FinishErrorsDur.Add(elapsed)
			return aof.err
		}
	}
	_ = aof.tailers.Wake()
	return nil
}

func (aof *AOF) Wake() error {
	if err := aof.tailers.Wake(); err != nil {
		return err
	}
	return nil
}

func (aof *AOF) Flush() error {
	data := aof.data
	if len(data) == 0 {
		return nil
	}
	if aof.flushSize == aof.size {
		return nil
	}
	aof.flushSize = aof.size
	begin := timex.NanoTime()
	err := data.Flush()
	elapsed := timex.NanoTime() - begin
	aof.m.stats.Flushes.Incr()
	aof.m.stats.FlushesDur.Add(elapsed)
	if err != nil {
		aof.m.stats.FlushErrors.Incr()
		aof.m.stats.FlushErrorsDur.Add(elapsed)
	}
	return err
}

func (aof *AOF) flush() error {
	data := aof.data
	if len(data) == 0 {
		return nil
	}
	if aof.flushSize == aof.size {
		return nil
	}
	return data.Flush()
}

func (aof *AOF) Sync() error {
	if aof.syncSize == aof.size {
		return nil
	}
	aof.syncSize = aof.size
	begin := timex.NanoTime()
	err := aof.sync()
	elapsed := timex.NanoTime() - begin
	aof.m.stats.Syncs.Incr()
	aof.m.stats.SyncsDur.Add(elapsed)
	if err != nil {
		aof.m.stats.SyncErrors.Incr()
		aof.m.stats.SyncErrorsDur.Add(elapsed)
	}
	return err
}

func (aof *AOF) sync() error {
	if aof.syncSize == aof.size {
		return nil
	}
	aof.syncSize = aof.size
	err := aof.flush()
	if err != nil {
		_ = aof.f.Sync()
		return err
	}
	return aof.f.Sync()
}

func (aof *AOF) Write(b []byte) (int, error) {
	if aof.err != nil {
		return 0, aof.err
	}
	if len(b) == 0 {
		return 0, io.ErrShortBuffer
	}
	aof.writeMu.Lock()
	defer aof.writeMu.Unlock()
	if aof.state != FileStateOpened {
		return 0, os.ErrClosed
	}
	var (
		size    = atomic.LoadInt64(&aof.size)
		newSize = size + int64(len(b))
	)
	if aof.recovery.Magic.IsEnabled() {
		newSize += 8
	}
	if newSize > int64(len(aof.data)) {
		return 0, io.EOF
	}
	if aof.f != nil {
		fileSize := atomic.LoadInt64(&aof.fileSize)
		if newSize > fileSize {
			aof.stats.blockingCount.Incr()
			fileSize = aof.geometry.Next(newSize)
			begin := timex.NanoTime()
			aof.err = aof.truncate(fileSize)
			elapsed := timex.NanoTime() - begin
			aof.stats.blockingDur.Add(elapsed)
			if aof.err != nil {
				aof.stats.truncErrDur.Add(elapsed)
				aof.stats.truncErrCount.Incr()
				return 0, aof.err
			} else {
				aof.stats.truncDur.Add(elapsed)
				aof.stats.truncCount.Incr()
			}
		}
	}
	copy(aof.data[size:], b)
	if aof.recovery.Magic.IsEnabled() {
		newSize -= 8
		write64LE(unsafe.Pointer(&aof.data[newSize]), aof.recovery.Magic.Tail)
	}
	atomic.StoreInt64(&aof.size, newSize)
	_ = aof.Wake()
	return len(b), nil
}

func (aof *AOF) WriteNonBlocking(b []byte) (int, error) {
	if aof.state.Load() != FileStateOpened {
		return 0, os.ErrClosed
	}
	if aof.err != nil {
		return 0, aof.err
	}
	if len(b) == 0 {
		return 0, io.ErrShortBuffer
	}
	var (
		size    = atomic.LoadInt64(&aof.size)
		newSize = size + int64(len(b))
	)
	if aof.recovery.Magic.IsEnabled() {
		newSize += 8
	}
	if newSize > int64(len(aof.data)) {
		return 0, io.EOF
	}
	if aof.f != nil {
		fileSize := atomic.LoadInt64(&aof.fileSize)
		if newSize > fileSize {
			return 0, ErrWouldBlock
		}
	}
	copy(aof.data[size:], b)
	if aof.recovery.Magic.IsEnabled() {
		newSize -= 8
		write64LE(unsafe.Pointer(&aof.data[newSize]), aof.recovery.Magic.Tail)
	}
	atomic.StoreInt64(&aof.size, newSize)
	_ = aof.Wake()
	return len(b), nil
}

func (aof *AOF) invokeAppendFn(event AppendEvent, fn AppendFunc) (n int64, err error) {
	if fn == nil {
		return 0, nil
	}
	defer func() {
		if e := recover(); e != nil {
			err = util.PanicToError(e)
			logger.Warn(err, "panic")
		}
	}()
	return fn(event)
}

//func (a *AOF) invokeErrorFn(err error) {
//	defer func() {
//		e := recover()
//		logger.Warn(util.PanicToError(e), "panic")
//	}()
//	if a.errFn != nil {
//		a.errFn(err)
//	}
//}

func (aof *AOF) Append(reserve int64, appendFn AppendFunc) error {
	if aof.err != nil {
		return aof.err
	}
	if appendFn == nil {
		return ErrAppendFuncNil
	}
	aof.writeMu.Lock()
	defer aof.writeMu.Unlock()
	if aof.state != FileStateOpened {
		return os.ErrClosed
	}
	var (
		size    = atomic.LoadInt64(&aof.size)
		newSize = size + reserve
	)
	if aof.recovery.Magic.IsEnabled() {
		newSize += 8
	}
	if newSize > int64(len(aof.data)) {
		return io.EOF
	}
	if aof.f != nil {
		fileSize := atomic.LoadInt64(&aof.fileSize)
		if newSize > fileSize {
			aof.stats.blockingCount.Incr()
			fileSize = aof.geometry.Next(newSize)
			begin := timex.NanoTime()
			aof.err = aof.truncate(fileSize)
			elapsed := timex.NanoTime() - begin
			aof.stats.blockingDur.Add(elapsed)
			if aof.err != nil {
				aof.stats.truncErrDur.Add(elapsed)
				aof.stats.truncErrCount.Incr()
				return aof.err
			} else {
				aof.stats.truncDur.Add(elapsed)
				aof.stats.truncCount.Incr()
			}
		}
	}
	if aof.recovery.Magic.IsEnabled() {
		newSize -= 8
	}
	return aof.completeAppend(reserve, AppendEvent{
		Begin: size,
		End:   newSize,
		Tail:  aof.data[size:newSize],
		file:  aof.data[0:newSize],
	}, appendFn)
}

func (aof *AOF) AppendNonBlocking(reserve int64, appendFn AppendFunc) error {
	if aof.state.Load() != FileStateOpened {
		return os.ErrClosed
	}
	if aof.err != nil {
		return aof.err
	}
	if appendFn == nil {
		return nil
	}
	var (
		size    = atomic.LoadInt64(&aof.size)
		newSize = size + reserve
	)
	if aof.recovery.Magic.IsEnabled() {
		newSize += 8
	}
	if newSize > int64(len(aof.data)) {
		return io.EOF
	}
	if aof.f != nil {
		fileSize := atomic.LoadInt64(&aof.fileSize)
		if newSize > fileSize {
			return ErrWouldBlock
		}
	}
	if aof.recovery.Magic.IsEnabled() {
		newSize -= 8
	}
	return aof.completeAppend(reserve, AppendEvent{
		Begin: size,
		End:   newSize,
		Tail:  aof.data[size:newSize],
		file:  aof.data[0:newSize],
	}, appendFn)
}

func (aof *AOF) completeAppend(reserve int64, event AppendEvent, fn AppendFunc) error {
	var n int64
	n, aof.err = aof.invokeAppendFn(event, fn)
	if aof.err != nil {
		return aof.err
	} else {
		if n < 0 {
			n = 0
		}
		if n > reserve {
			n = reserve
		}
		// Write magic tail
		if aof.recovery.Magic.IsEnabled() {
			write64LE(unsafe.Pointer(&aof.data[n]), aof.recovery.Magic.Tail)
		}
		atomic.StoreInt64(&aof.size, event.Begin+n)
		return nil
	}
}

func (aof *AOF) truncate(size int64) error {
	aof.truncMu.Lock()
	defer aof.truncMu.Unlock()
	if size < aof.size {
		return ErrShrink
	}
	err := syscall.Truncate(aof.f.Name(), size)
	if err != nil {
		return err
	}
	atomic.StoreInt64(&aof.fileSize, size)
	return nil
}
