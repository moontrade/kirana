// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomicx_test

import (
	"github.com/moontrade/kirana/pkg/atomicx"
	"runtime"
	"testing"
	"unsafe"
)

func runParallel(N, iter int, f func()) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(int(N)))
	done := make(chan bool)
	for i := 0; i < N; i++ {
		go func() {
			for j := 0; j < iter; j++ {
				f()
			}
			done <- true
		}()
	}
	for i := 0; i < N; i++ {
		<-done
	}
}

func TestXadduintptr(t *testing.T) {
	N := 20
	iter := 100000
	if testing.Short() {
		N = 10
		iter = 10000
	}
	inc := uintptr(100)
	total := uintptr(0)
	runParallel(N, iter, func() {
		atomicx.Xadduintptr(&total, inc)
	})
	if want := uintptr(N*iter) * inc; want != total {
		t.Fatalf("xadduintpr error, want %d, got %d", want, total)
	}
	total = 0
	runParallel(N, iter, func() {
		atomicx.Xadduintptr(&total, inc)
		atomicx.Xadduintptr(&total, uintptr(-int64(inc)))
	})
	if total != 0 {
		t.Fatalf("xadduintpr total error, want %d, got %d", 0, total)
	}
}

// Tests that xadduintptr correctly updates 64-bit values. The place where
// we actually do so is mstats.go, functions mSysStat{Inc,Dec}.
func TestXadduintptrOnUint64(t *testing.T) {
	//if goarch.BigEndian {
	//	// On big endian architectures, we never use xadduintptr to update
	//	// 64-bit values and hence we skip the test.  (Note that functions
	//	// mSysStat{Inc,Dec} in mstats.go have explicit checks for
	//	// big-endianness.)
	//	t.Skip("skip xadduintptr on big endian architecture")
	//}
	const inc = 100
	val := uint64(0)
	atomicx.Xadduintptr((*uintptr)(unsafe.Pointer(&val)), inc)
	if inc != val {
		t.Fatalf("xadduintptr should increase lower-order bits, want %d, got %d", inc, val)
	}
}

func shouldPanic(t *testing.T, name string, f func()) {
	defer func() {
		// Check that all GC maps are sane.
		runtime.GC()

		err := recover()
		want := "unaligned 64-bit atomic operation"
		if err == nil {
			t.Errorf("%s did not panic", name)
		} else if s, _ := err.(string); s != want {
			t.Errorf("%s: wanted panic %q, got %q", name, want, err)
		}
	}()
	f()
}

// Variant of sync/atomic's TestUnaligned64:
func TestUnaligned64(t *testing.T) {
	// Unaligned 64-bit atomics on 32-bit systems are
	// a continual source of pain. Test that on 32-bit systems they crash
	// instead of failing silently.

	if unsafe.Sizeof(int(0)) != 4 {
		t.Skip("test only runs on 32-bit systems")
	}

	x := make([]uint32, 4)
	u := unsafe.Pointer(uintptr(unsafe.Pointer(&x[0])) | 4) // force alignment to 4

	up64 := (*uint64)(u) // misaligned
	p64 := (*int64)(u)   // misaligned

	shouldPanic(t, "Load64", func() { atomicx.Load64(up64) })
	shouldPanic(t, "Loadint64", func() { atomicx.Loadint64(p64) })
	shouldPanic(t, "Store64", func() { atomicx.Store64(up64, 0) })
	shouldPanic(t, "Xadd64", func() { atomicx.Xadd64(up64, 1) })
	shouldPanic(t, "Xchg64", func() { atomicx.Xchg64(up64, 1) })
	shouldPanic(t, "Cas64", func() { atomicx.Cas64(up64, 1, 2) })
}

func TestAnd8(t *testing.T) {
	// Basic sanity check.
	x := uint8(0xff)
	for i := uint8(0); i < 8; i++ {
		atomicx.And8(&x, ^(1 << i))
		if r := uint8(0xff) << (i + 1); x != r {
			t.Fatalf("clearing bit %#x: want %#x, got %#x", uint8(1<<i), r, x)
		}
	}

	// Set every bit in array to 1.
	a := make([]uint8, 1<<12)
	for i := range a {
		a[i] = 0xff
	}

	// Clear array bit-by-bit in different goroutines.
	done := make(chan bool)
	for i := 0; i < 8; i++ {
		m := ^uint8(1 << i)
		go func() {
			for i := range a {
				atomicx.And8(&a[i], m)
			}
			done <- true
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}

	// Check that the array has been totally cleared.
	for i, v := range a {
		if v != 0 {
			t.Fatalf("a[%v] not cleared: want %#x, got %#x", i, uint8(0), v)
		}
	}
}

func TestAnd(t *testing.T) {
	// Basic sanity check.
	x := uint32(0xffffffff)
	for i := uint32(0); i < 32; i++ {
		atomicx.And(&x, ^(1 << i))
		if r := uint32(0xffffffff) << (i + 1); x != r {
			t.Fatalf("clearing bit %#x: want %#x, got %#x", uint32(1<<i), r, x)
		}
	}

	// Set every bit in array to 1.
	a := make([]uint32, 1<<12)
	for i := range a {
		a[i] = 0xffffffff
	}

	// Clear array bit-by-bit in different goroutines.
	done := make(chan bool)
	for i := 0; i < 32; i++ {
		m := ^uint32(1 << i)
		go func() {
			for i := range a {
				atomicx.And(&a[i], m)
			}
			done <- true
		}()
	}
	for i := 0; i < 32; i++ {
		<-done
	}

	// Check that the array has been totally cleared.
	for i, v := range a {
		if v != 0 {
			t.Fatalf("a[%v] not cleared: want %#x, got %#x", i, uint32(0), v)
		}
	}
}

func TestOr8(t *testing.T) {
	// Basic sanity check.
	x := uint8(0)
	for i := uint8(0); i < 8; i++ {
		atomicx.Or8(&x, 1<<i)
		if r := (uint8(1) << (i + 1)) - 1; x != r {
			t.Fatalf("setting bit %#x: want %#x, got %#x", uint8(1)<<i, r, x)
		}
	}

	// Start with every bit in array set to 0.
	a := make([]uint8, 1<<12)

	// Set every bit in array bit-by-bit in different goroutines.
	done := make(chan bool)
	for i := 0; i < 8; i++ {
		m := uint8(1 << i)
		go func() {
			for i := range a {
				atomicx.Or8(&a[i], m)
			}
			done <- true
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}

	// Check that the array has been totally set.
	for i, v := range a {
		if v != 0xff {
			t.Fatalf("a[%v] not fully set: want %#x, got %#x", i, uint8(0xff), v)
		}
	}
}

func TestOr(t *testing.T) {
	// Basic sanity check.
	x := uint32(0)
	for i := uint32(0); i < 32; i++ {
		atomicx.Or(&x, 1<<i)
		if r := (uint32(1) << (i + 1)) - 1; x != r {
			t.Fatalf("setting bit %#x: want %#x, got %#x", uint32(1)<<i, r, x)
		}
	}

	// Start with every bit in array set to 0.
	a := make([]uint32, 1<<12)

	// Set every bit in array bit-by-bit in different goroutines.
	done := make(chan bool)
	for i := 0; i < 32; i++ {
		m := uint32(1 << i)
		go func() {
			for i := range a {
				atomicx.Or(&a[i], m)
			}
			done <- true
		}()
	}
	for i := 0; i < 32; i++ {
		<-done
	}

	// Check that the array has been totally set.
	for i, v := range a {
		if v != 0xffffffff {
			t.Fatalf("a[%v] not fully set: want %#x, got %#x", i, uint32(0xffffffff), v)
		}
	}
}

func TestBitwiseContended8(t *testing.T) {
	// Start with every bit in array set to 0.
	a := make([]uint8, 16)

	// Iterations to try.
	N := 1 << 16
	if testing.Short() {
		N = 1 << 10
	}

	// Set and then clear every bit in the array bit-by-bit in different goroutines.
	done := make(chan bool)
	for i := 0; i < 8; i++ {
		m := uint8(1 << i)
		go func() {
			for n := 0; n < N; n++ {
				for i := range a {
					atomicx.Or8(&a[i], m)
					if atomicx.Load8(&a[i])&m != m {
						t.Errorf("a[%v] bit %#x not set", i, m)
					}
					atomicx.And8(&a[i], ^m)
					if atomicx.Load8(&a[i])&m != 0 {
						t.Errorf("a[%v] bit %#x not clear", i, m)
					}
				}
			}
			done <- true
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}

	// Check that the array has been totally cleared.
	for i, v := range a {
		if v != 0 {
			t.Fatalf("a[%v] not cleared: want %#x, got %#x", i, uint8(0), v)
		}
	}
}

func TestBitwiseContended(t *testing.T) {
	// Start with every bit in array set to 0.
	a := make([]uint32, 16)

	// Iterations to try.
	N := 1 << 16
	if testing.Short() {
		N = 1 << 10
	}

	// Set and then clear every bit in the array bit-by-bit in different goroutines.
	done := make(chan bool)
	for i := 0; i < 32; i++ {
		m := uint32(1 << i)
		go func() {
			for n := 0; n < N; n++ {
				for i := range a {
					atomicx.Or(&a[i], m)
					if atomicx.Load(&a[i])&m != m {
						t.Errorf("a[%v] bit %#x not set", i, m)
					}
					atomicx.And(&a[i], ^m)
					if atomicx.Load(&a[i])&m != 0 {
						t.Errorf("a[%v] bit %#x not clear", i, m)
					}
				}
			}
			done <- true
		}()
	}
	for i := 0; i < 32; i++ {
		<-done
	}

	// Check that the array has been totally cleared.
	for i, v := range a {
		if v != 0 {
			t.Fatalf("a[%v] not cleared: want %#x, got %#x", i, uint32(0), v)
		}
	}
}

func TestStorepNoWB(t *testing.T) {
	var p [2]*int
	for i := range p {
		atomicx.StorepNoWB(unsafe.Pointer(&p[i]), unsafe.Pointer(new(int)))
	}
	if p[0] == p[1] {
		t.Error("Bad escape analysis of StorepNoWB")
	}
}
