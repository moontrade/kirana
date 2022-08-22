package reactor

import (
	"github.com/moontrade/wormhole/pkg/pmath"
	"github.com/moontrade/wormhole/pkg/pool"
	"reflect"
	"runtime"
	"unsafe"
)

var (
	taskPool = pool.NewPool[Task](pool.Config[Task]{
		SizeClass:     int(unsafe.Sizeof(Task{})),
		PageSize:      1024,
		PagesPerShard: 4096,
	})
	taskSlotPool = pool.NewPool[TaskSlot](pool.Config[TaskSlot]{
		SizeClass:     int(unsafe.Sizeof(TaskSlot{})),
		PageSize:      1024,
		PagesPerShard: 4096,
	})
	taskSwapSlicePool = pool.NewPool[TaskSwapSlice](pool.Config[TaskSwapSlice]{
		SizeClass:     int(unsafe.Sizeof(TaskSwapSlice{})),
		PageSize:      1024,
		PagesPerShard: 4096,
	})
	funcSlotPool = pool.NewPool[FuncSlot](pool.Config[FuncSlot]{
		SizeClass:     int(unsafe.Sizeof(FuncSlot{})),
		PageSize:      1024,
		PagesPerShard: 4096,
	})
	funcSwapSlicePool = pool.NewPool[FuncSwapSlice](pool.Config[FuncSwapSlice]{
		SizeClass:     int(unsafe.Sizeof(FuncSwapSlice{})),
		PageSize:      1024,
		PagesPerShard: 4096,
	})
	wakeListPool = pool.NewPool[WakeList](pool.Config[WakeList]{
		SizeClass:     int(unsafe.Sizeof(WakeList{})),
		PageSize:      1024,
		PagesPerShard: 4096,
	})
	wakeListSlicePool = pool.NewPool[WakeLists](pool.Config[WakeLists]{
		SizeClass:     int(unsafe.Sizeof(reflect.SliceHeader{}) * uintptr(runtime.GOMAXPROCS(0))),
		PageSize:      1024,
		PagesPerShard: 4096,
		AllocFunc: func() unsafe.Pointer {
			return unsafe.Pointer(&WakeLists{slots: make([]*WakeList, 0, pmath.CeilToPowerOf2(NumReactors()))})
		},
	})
)

type WakeLists struct {
	slots []*WakeList
	count int
}

func (w *WakeLists) Len() int {
	return len(w.slots)
}

func (w *WakeLists) Get(reactorID int) (*WakeList, bool) {
	if !w.Ensure(reactorID) {
		return nil, false
	}
	return w.slots[reactorID], true
}

func (w *WakeLists) GetOrCreate(reactorID int, tl *TaskSet) (*WakeList, bool) {
	if !w.Ensure(reactorID) {
		return nil, false
	}
	slot := w.slots[reactorID]
	if slot == nil {
		slot = wakeListPool.Get()
		slot.init(tl)
		w.slots[reactorID] = slot
	}
	return w.slots[reactorID], true
}

func (w *WakeLists) Put(reactorID int, l *WakeList) {
	w.slots[reactorID] = l
	w.count++
}

func (w *WakeLists) Ensure(reactorID int) bool {
	if reactorID < 0 || reactorID > 65535 {
		return false
	}
	if reactorID >= len(w.slots) {
		numReactors := pmath.CeilToPowerOf2(NumReactors())
		if numReactors <= reactorID {
			numReactors = pmath.CeilToPowerOf2(reactorID)
		}
		next := make([]*WakeList, numReactors)
		if len(w.slots) > 0 {
			copy(next, w.slots)
		}
		w.slots = next
	}
	return true
}
