//go:build (amd64 || arm64) && go1.20

package logger

import (
	_ "unsafe"
)

//go:noescape
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:noescape
//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()

//go:nosplit
func procPin() *g {
	gp := getg()
	mp := gp.m
	mp.locks++
	return gp
}

func procUnpinGp(gp *g) {
	gp = getg()
	gp.m.locks--
}

//go:nosplit
func procUnpin() {
	gp := getg()
	gp.m.locks--
}
