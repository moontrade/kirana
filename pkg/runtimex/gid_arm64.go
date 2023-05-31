package runtimex

import "unsafe"

func GoroutineID() int64 {
	return (*g)(getg()).goid
}

func MachineID() int64 {
	gg := (*g)(getg())
	return gg.m.id
}

func ProcID() uint64 {
	gg := (*g)(getg())
	return gg.m.procid
}

func ProcessorID() int32 {
	gg := (*g)(getg())
	return gg.m.p.id
	//return (*g)(getg()).m.p.id
}

func GIDPID() (gid int64, pid int32) {
	gg := (*g)(getg())
	return gg.goid, gg.m.p.id
}

func getg() unsafe.Pointer
