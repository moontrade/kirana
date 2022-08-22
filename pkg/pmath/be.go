//go:build arm64be || armbe || mips || mips64 || ppc || ppc64 || s390 || s390x || sparc || sparc64

package pmath

import (
	"math/bits"
)

func PowerOf2Index(size int) int {
	return bits.LeadingZeros64(uint64(CeilToPowerOf2(size)))
}
