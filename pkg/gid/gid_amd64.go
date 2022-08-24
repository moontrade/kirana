package gid

import "unsafe"

func GID() int64 {
	return (*g)(getg()).goid
}

func PID() int32 {
	return (*g)(getg()).m.p.id
}

func GIDPID() (gid int64, pid int32) {
	gg := (*g)(getg())
	return gg.goid, gg.m.p.id
}

func getg() unsafe.Pointer
