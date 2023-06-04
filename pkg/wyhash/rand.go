package wyhash

import (
	crand "crypto/rand"
	"encoding/binary"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/moontrade/kirana/pkg/pmath"
	"github.com/moontrade/kirana/pkg/runtimex"
)

var (
	randShards     []AtomicRand
	randShardsMask uint64 = 0
)

func init() {
	shards := pmath.CeilToPowerOf2(runtime.GOMAXPROCS(0) * 64)
	randShardsMask = uint64(shards - 1)
	randShards = make([]AtomicRand, shards)
	for i := 0; i < len(randShards); i++ {
		randShards[i].seed = CryptoSeed()
	}
}

func CryptoSeed() uint64 {
	seed, err := cryptoU64()
	if err != nil {
		seed = uint64(time.Now().UnixNano())
	}
	return seed
}

// cryptoU64 returns cryptographic random uint64.
func cryptoU64() (uint64, error) {
	var v uint64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func Next() uint64 {
	return randShards[uint64(runtimex.ProcID())&randShardsMask].Next()
}

func NextFloat() float64 {
	return randShards[runtimex.ProcID()&randShardsMask].NextFloat()
}

func NextGaussian() float64 {
	return randShards[runtimex.ProcID()&randShardsMask].NextGaussian()
}

type Rand struct {
	seed uint64
}

func NewRand() *Rand {
	return &Rand{CryptoSeed()}
}

func NewRandWithSeed(seed uint64) *Rand {
	return &Rand{seed}
}

func (r *Rand) Next() uint64 {
	return wyrand(&r.seed)
}

func (r *Rand) NextFloat() float64 {
	return wy2u01(wyrand(&r.seed))
}

func (r *Rand) NextGaussian() float64 {
	return wy2gau(wyrand(&r.seed))
}

type AtomicRand struct {
	seed uint64
}

func NewAtomicRand() *AtomicRand {
	return &AtomicRand{CryptoSeed()}
}

func NewAtomicRandWithSeed(seed uint64) *AtomicRand {
	return &AtomicRand{seed}
}

func (r *AtomicRand) Next() uint64 {
	//seed := atomicx.Xadd64(&r.seed, int64(uint64(0xa0761d64bd642f)))
	seed := atomic.AddUint64(&r.seed, uint64(0xa0761d6478bd642f))
	return wymix(seed, seed^0xe7037ed1a0b428db)
}

func (r *AtomicRand) NextFloat() float64 {
	seed := atomic.AddUint64(&r.seed, uint64(0xa0761d6478bd642f))
	return wy2u01(wymix(seed, seed^0xe7037ed1a0b428db))
}

func (r *AtomicRand) NextGaussian() float64 {
	seed := atomic.AddUint64(&r.seed, uint64(0xa0761d6478bd642f))
	return wy2gau(wymix(seed, seed^0xe7037ed1a0b428db))
}
