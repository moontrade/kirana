package aof

import (
	"fmt"
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/moontrade/wormhole/pkg/uid"
	"github.com/pidato/unsafe/memory/hash"
	"testing"
)

func BenchmarkUID(b *testing.B) {
	b.Run("Crypto", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = uid.CryptoU64()
		}
	})

	b.Run("Fastrand 32bit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastrand.Uint32()
		}
	})

	b.Run("Fastrand 64bit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fastrand.Uint64()
		}
	})

	b.Run("Wyhash 64bit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uid.UID()
		}
	})
}

func cryptoUID() uint64 {
	r, _ := uid.CryptoU64()
	return r
}

func TestCollisions(t *testing.T) {
	//fmt.Println("Crypto 1m", collisions(10000000, cryptoUID))
	fmt.Println("Wyhash", collisions(10000000, hash.Next))
}

func collisions(c int, fn func() uint64) int {
	m := make(map[uint64]struct{})
	r := 0
	for i := 0; i < c; i++ {
		h := fn()
		if _, ok := m[h]; ok {
			r++
		} else {
			m[h] = struct{}{}
		}
	}
	return r
}
