package aof

import (
	"fmt"
	"github.com/moontrade/wormhole/pkg/hashmap"
	"github.com/moontrade/wormhole/pkg/spinlock"
	"github.com/moontrade/wormhole/pkg/swap"
	"github.com/moontrade/wormhole/pkg/timex"
	"os"
	"time"

	. "github.com/moontrade/wormhole/pkg/counter"
)

type Stats struct {
	Creates               Counter
	CreatesDur            TimeCounter
	Opens                 Counter
	OpensDur              TimeCounter
	OpenErrors            Counter
	OpenErrorsDur         TimeCounter
	OpenFileCount         Counter
	OpenFileDur           TimeCounter
	OpenFileErrors        Counter
	OpenFileErrorsDur     TimeCounter
	Closes                Counter
	CloseDur              TimeCounter
	ActiveMaps            Counter
	ActiveFileSize        Counter
	ActiveMappedMemory    Counter
	ActiveAnonymousMemory Counter
	LifetimeMemory        Counter
	Flushes               Counter
	FlushesDur            TimeCounter
	FlushErrors           Counter
	FlushErrorsDur        TimeCounter
	Finishes              Counter
	FinishesDur           TimeCounter
	FinishErrors          Counter
	FinishErrorsDur       TimeCounter
	Syncs                 Counter
	SyncsDur              TimeCounter
	SyncErrors            Counter
	SyncErrorsDur         TimeCounter
	Maps                  Counter
	MapsDur               TimeCounter
	MapErrors             Counter
	MapErrorsDur          TimeCounter
	Unmaps                Counter
	UnmapsDur             TimeCounter
	UnmapErrors           Counter
	UnmapErrorsDur        TimeCounter
	Finalizes             Counter
	FinalizesDur          TimeCounter
	Truncates             Counter
	TruncatesDur          TimeCounter
	TruncateErrors        Counter
	TruncateErrorsDur     TimeCounter
}

type Manager struct {
	dir       string
	absDir    string
	stats     Stats
	perm      os.FileMode
	closing   int64
	closed    int64
	isClosed  bool
	files     *hashmap.Sync[string, *AOF]
	gcList    *swap.SyncSlice[*AOF]
	flushList *swap.SyncSlice[*AOF]
	mu        spinlock.Mutex
}

func (m *Manager) Stats() Stats {
	return m.stats
}

func NewManager(dir string, perm os.FileMode) (*Manager, error) {
	if perm == 0 {
		perm = 0755
	}
	if len(dir) > 0 {
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, perm)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			if !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", dir)
			}
		}
	} else {
	}
	m := &Manager{
		dir:       dir,
		perm:      perm,
		files:     hashmap.NewSync[string, *AOF](1024, 1024, hashmap.HashString),
		gcList:    swap.NewSync[*AOF](getGCIndex, setGCIndex),
		flushList: swap.NewSync[*AOF](getFlushIndex, setFlushIndex),
	}
	go m.run()
	return m, nil
}

func (m *Manager) run() {
	var (
		gcList []*AOF
	)
	for !m.isClosed {
		time.Sleep(time.Second)

		flush := m.flushList.Unsafe()
		for i := 0; i < len(flush); i++ {
			aof := flush[i]
			if aof == nil {
				continue
			}
			_ = aof.Flush()
		}

		gcList = m.gcList.CopyTo(gcList)
		for i, aof := range gcList {
			gcList[i] = nil
			if aof == nil {
				continue
			}
			if !aof.tailers.IsEmpty() {
				_ = aof.tailers.Wake()
				continue
			}
			m.gcList.Remove(aof)
		}
	}
}

func (m *Manager) Close() error {
	m.mu.Lock()
	if m.closing > 0 || m.closed > 0 {
		m.mu.Unlock()
		return os.ErrClosed
	}
	m.closing = timex.NanoTime()
	m.isClosed = true
	m.mu.Unlock()
	m.files.Scan(func(key string, value *AOF) bool {
		//_ = value.Close()
		return true
	})
	m.files.Scan(func(key string, value *AOF) bool {
		return true
	})
	return nil
}
