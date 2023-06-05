package wyhash

import (
	crand "crypto/rand"
	"encoding/binary"
	"time"
)

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
