package wasmtimex

// #cgo CXXFLAGS: -std=c++2a
// #cgo LDFLAGS: -lstdc++
// #cgo CFLAGS:-I${SRCDIR}/build/include
// #cgo CXXFLAGS:-I${SRCDIR}/build/include
// #cgo !windows LDFLAGS:-lwasmtime -lm -ldl -pthread
// #cgo windows CFLAGS:-DWASM_API_EXTERN= -DWASI_API_EXTERN=
// #cgo windows LDFLAGS:-lwasmtime -luserenv -lole32 -lntdll -lws2_32 -lkernel32 -lbcrypt
// #cgo linux,amd64 LDFLAGS:-L${SRCDIR}/build/linux-x86_64
// #cgo linux,arm64 LDFLAGS:-L${SRCDIR}/build/linux-aarch64
// #cgo darwin,amd64 LDFLAGS:-L${SRCDIR}/build/macos-x86_64
// #cgo darwin,arm64 LDFLAGS:-L${SRCDIR}/build/macos-aarch64
// #cgo windows,amd64 LDFLAGS:-L${SRCDIR}/build/windows-x86_64
// #include <wasm.h>
// #include <wasmtime.h>
import "C"

//export GoTick
func GoTick() {
	//fmt.Println("tick")
}
