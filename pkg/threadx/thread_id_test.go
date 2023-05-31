package threadx

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/runtimex"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGetThreadID(t *testing.T) {
	fmt.Println(CurrentThreadID())
}

func BenchmarkGetThreadID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CurrentThreadID()
	}
}

func TestThreading(t *testing.T) {
	wg := new(sync.WaitGroup)
	_ = wg
	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			var (
				tid    = CurrentThreadID()
				pid    = runtimex.ProcessorID()
				goid   = runtimex.GoroutineID()
				mid    = runtimex.MachineID()
				procid = runtimex.ProcID()
			)
			fmt.Println("G2:", "tid", tid, "pid", pid, "gid", goid, "mid", mid, "procid", procid)
			for {
				var (
					ntid    = CurrentThreadID()
					npid    = runtimex.ProcessorID()
					ngoid   = runtimex.GoroutineID()
					nmid    = runtimex.MachineID()
					nprocid = runtimex.ProcID()
				)
				if tid != ntid {
					fmt.Println("G2 thread id changed from:", tid, "to", ntid)
					tid = ntid
				}
				if pid != npid {
					fmt.Println("G2 processor id changed from:", pid, "to", npid)
					pid = npid
				}
				if goid != ngoid {
					fmt.Println("G2 goroutine id changed from:", goid, "to", ngoid)
					goid = ngoid
				}
				if mid != nmid {
					fmt.Println("G2 machine id changed from:", mid, "to", nmid)
					mid = nmid
					t.Fatal("G1 machine id changed")
				}
				if procid != nprocid {
					fmt.Println("G2 proc id changed from:", procid, "to", nprocid)
					procid = nprocid
				}
				//cgo.NonBlocking((*byte)(cgo2.Sleep), uintptr(time.Second), 0)
				time.Sleep(time.Second)
			}
		}()

		var (
			tid    = CurrentThreadID()
			pid    = runtimex.ProcessorID()
			goid   = runtimex.GoroutineID()
			mid    = runtimex.MachineID()
			procid = runtimex.ProcID()
		)
		fmt.Println("G1:", "tid", tid, "pid", pid, "gid", goid, "mid", mid, "procid", procid)

		for {
			var (
				ntid    = CurrentThreadID()
				npid    = runtimex.ProcessorID()
				ngoid   = runtimex.GoroutineID()
				nmid    = runtimex.MachineID()
				nprocid = runtimex.ProcID()
			)
			if tid != ntid {
				fmt.Println("G1 thread id changed from:", tid, "to", ntid)
				tid = ntid
			}
			if pid != npid {
				fmt.Println("G1 processor id changed from:", pid, "to", npid)
				pid = npid
			}
			if goid != ngoid {
				fmt.Println("G1 goroutine id changed from:", goid, "to", ngoid)
				goid = ngoid
			}
			if mid != nmid {
				fmt.Println("G1 machine id changed from:", mid, "to", nmid)
				mid = nmid
				t.Fatal("G1 machine id changed")
			}
			if procid != nprocid {
				fmt.Println("G1 proc id changed from:", procid, "to", nprocid)
				procid = nprocid
			}
			//cgo.NonBlocking((*byte)(cgo2.Sleep), uintptr(time.Second), 0)
			time.Sleep(time.Second)
		}
	}()

	var (
		tid    = CurrentThreadID()
		pid    = runtimex.ProcessorID()
		goid   = runtimex.GoroutineID()
		mid    = runtimex.MachineID()
		procid = runtimex.ProcID()
	)

	if false {
		fmt.Println("G0:", "tid", tid, "pid", pid, "gid", goid, "mid", mid, "procid", procid)

		for {
			var (
				ntid    = CurrentThreadID()
				npid    = runtimex.ProcessorID()
				ngoid   = runtimex.GoroutineID()
				nmid    = runtimex.MachineID()
				nprocid = runtimex.ProcID()
			)
			if tid != ntid {
				fmt.Println("G0 thread id changed from:", tid, "to", ntid)
				tid = ntid
			}
			if pid != npid {
				fmt.Println("G0 processor id changed from:", pid, "to", npid)
				pid = npid
			}
			if goid != ngoid {
				fmt.Println("G0 goroutine id changed from:", goid, "to", ngoid)
				goid = ngoid
			}
			if mid != nmid {
				fmt.Println("G0 machine id changed from:", mid, "to", nmid)
				mid = nmid
			}
			if procid != nprocid {
				fmt.Println("G0 proc id changed from:", procid, "to", nprocid)
				procid = nprocid
			}
			time.Sleep(time.Second)
		}
	}

	wg.Wait()
}
