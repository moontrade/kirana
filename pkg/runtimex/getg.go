//go:build (amd64 || arm64) && go1.20

package runtimex

func GoroutineID() uint64 {
	return getg().goid
}

func MachineID() int64 {
	return getg().m.id
}

func ProcID() uint64 {
	return getg().m.procid
}

func ProcessorID() int32 {
	return getg().m.p.id
}

func GIDPID() (gid uint64, pid int32) {
	gg := getg()
	return gg.goid, gg.m.p.id
}

func getg() *g
