package wyhash

import (
	"math/bits"
	"unsafe"
)

const (
	DefaultSeed     uint64 = 16194578278573448045 // s0
	defaultSeedInit        = uint64(4668172604561817410)
	s0              uint64 = 0xa0761d6478bd642f // s0
	s1              uint64 = 0xe7037ed1a0b428db
	s2              uint64 = 0x8ebc6af09c88c6e3
	s3              uint64 = 0x589965cc75374cc3
	s4              uint64 = 0x1d8e4e27c47d124f
)

func wyrand(seed *uint64) uint64 {
	*seed += uint64(0xa0761d6478bd642f)
	return wymix(*seed, *seed^0xe7037ed1a0b428db)
}

func wy2u01(r uint64) float64 {
	const norm = float64(1.0) / float64(uint64(1)<<52)
	return float64(r>>12) * norm
}

func wy2gau(r uint64) float64 {
	const norm = float64(1.0) / float64(uint64(1)<<20)
	return float64((r&0x1fffff)+((r>>21)&0x1fffff)+((r>>42)&0x1fffff))*norm - 3.0
}

func wymum(A *uint64, B *uint64) {
	hi, lo := bits.Mul64(*A, *B)
	*A ^= hi
	*B ^= lo
}

func wymixSlow(a, b uint64) uint64 {
	var (
		hh = (a >> 32) * (b >> 32)
		hl = (a >> 32) * (b & 0xFFFF_FFFF)
		lh = (a & 0xFFFF_FFFF) * (b >> 32)
		ll = (a & 0xFFFF_FFFF) * (b & 0xFFFF_FFFF)
	)
	//a = wyrotate(hl) ^ hh
	//b = wyrotate(lh) ^ ll
	a = ((hl >> 32) | (hl << 32)) ^ hh
	b = ((lh >> 32) | (lh << 32)) ^ ll
	return a ^ b
}

func wymix(a, b uint64) uint64 {
	a, b = bits.Mul64(a, b)
	return a ^ b
}

func Bytes(b []byte) uint64 {
	return Hash(*(*unsafe.Pointer)(unsafe.Pointer(&b)), uint64(len(b)))
}

func String(s string) uint64 {
	//return Hash(*(*unsafe.Pointer)(unsafe.Pointer(&s)), uint64(len(s)), DefaultSeed)
	return Hash(*(*unsafe.Pointer)(unsafe.Pointer(&s)), uint64(len(s)))
}

func HashBytes(b []byte) uint64 {
	return Hash(unsafe.Pointer(&b[0]), uint64(len(b)))
}

func HashBytesWithSeed(b []byte, seed uint64) uint64 {
	return HashWithSeed(unsafe.Pointer(&b[0]), uint64(len(b)), seed)
}

func Hash(bytes unsafe.Pointer, length uint64) uint64 {
	var (
		a    uint64
		b    uint64
		seed = defaultSeedInit
	)
	if length <= 16 {
		if length >= 4 {
			a = read32(bytes)<<32 | read32(unsafe.Add(bytes, (length>>3)<<2))
			b = read32(unsafe.Add(bytes, length-4))<<32 |
				read32(unsafe.Add(bytes, length-4-((length>>3)<<2)))
		} else if length > 0 {
			a = uint64(*(*byte)(bytes))<<16 |
				uint64(*(*byte)(unsafe.Add(bytes, length>>1)))<<8 |
				uint64(*(*byte)(unsafe.Add(bytes, length-1)))
		} else {
			a = 0
			b = 0
		}
	} else {
		var (
			index = length
		)
		if length > 48 {
			var (
				see1 = seed
				see2 = seed
			)
			for index > 48 {
				seed = wymix(read64(bytes)^s1, read64(unsafe.Add(bytes, 8))^seed)
				see1 = wymix(read64(unsafe.Add(bytes, 16))^s2, read64(unsafe.Add(bytes, 24))^see1)
				see2 = wymix(read64(unsafe.Add(bytes, 32))^s3, read64(unsafe.Add(bytes, 40))^see2)
				index -= 48
				bytes = unsafe.Add(bytes, 48)
			}
			seed ^= see1 ^ see2
		}

		for index > 16 {
			seed = wymix(read64(bytes)^s1, read64(unsafe.Add(bytes, 8))^seed)
			index -= 16
			bytes = unsafe.Add(bytes, 16)
		}
		a = read64(unsafe.Add(bytes, index-16))
		b = read64(unsafe.Add(bytes, index-8))
	}

	//a ^= s1
	//b ^= seed
	//wymum(&a, &b)
	//return wymix(a^s0^length, b^s1)
	return wymix(s1^length, wymix(a^s1, b^seed))
}

func HashWithSeed(bytes unsafe.Pointer, length uint64, seed uint64) uint64 {
	seed ^= s0
	var (
		a uint64
		b uint64
	)
	if length <= 16 {
		if length >= 4 {
			a = read32(bytes)<<32 | read32(unsafe.Add(bytes, (length>>3)<<2))
			b = read32(unsafe.Add(bytes, length-4))<<32 |
				read32(unsafe.Add(bytes, length-4-((length>>3)<<2)))
		} else if length > 0 {
			a = uint64(*(*byte)(bytes))<<16 |
				uint64(*(*byte)(unsafe.Add(bytes, length>>1)))<<8 |
				uint64(*(*byte)(unsafe.Add(bytes, length-1)))
		}
	} else {
		var (
			index = length
		)
		if length > 48 {
			var (
				see1 = seed
				see2 = seed
			)
			for index > 48 {
				seed = wymix(read64(bytes)^s1, read64(unsafe.Add(bytes, 8))^seed)
				see1 = wymix(read64(unsafe.Add(bytes, 16))^s2, read64(unsafe.Add(bytes, 24))^see1)
				see2 = wymix(read64(unsafe.Add(bytes, 32))^s3, read64(unsafe.Add(bytes, 40))^see2)
				index -= 48
				bytes = unsafe.Add(bytes, 48)
			}
			seed ^= see1 ^ see2
		}

		for index > 16 {
			seed = wymix(read64(bytes)^s1, read64(unsafe.Add(bytes, 8))^seed)
			index -= 16
			bytes = unsafe.Add(bytes, 16)
		}
		a = read64(unsafe.Add(bytes, index-16))
		b = read64(unsafe.Add(bytes, index-8))
	}

	//a ^= s1
	//b ^= seed
	//wymum(&a, &b)
	//return wymix(a^s0^length, b^s1)
	return wymix(s1^length, wymix(a^s1, b^seed))
}

func WithSecret(bytes unsafe.Pointer, length uint64, seed uint64, secret *[4]uint64) uint64 {
	seed ^= wymix(seed^secret[0], secret[1])
	var (
		a uint64
		b uint64
	)
	if length <= 16 {
		if length >= 4 {
			a = read32(bytes)<<32 | read32(unsafe.Add(bytes, (length>>3)<<2))
			b = read32(unsafe.Add(bytes, length-4))<<32 |
				read32(unsafe.Add(bytes, length-4-((length>>3)<<2)))
		} else if length > 0 {
			a = uint64(*(*byte)(bytes))<<16 |
				uint64(*(*byte)(unsafe.Add(bytes, length>>1)))<<8 |
				uint64(*(*byte)(unsafe.Add(bytes, length-1)))
		}
	} else {
		var (
			index = length
		)
		if length > 48 {
			var (
				see1 = seed
				see2 = seed
			)
			for index > 48 {
				seed = wymix(read64(bytes)^secret[1], read64(unsafe.Add(bytes, 8))^seed)
				see1 = wymix(read64(unsafe.Add(bytes, 16))^secret[2], read64(unsafe.Add(bytes, 24))^see1)
				see2 = wymix(read64(unsafe.Add(bytes, 32))^secret[3], read64(unsafe.Add(bytes, 40))^see2)
				index -= 48
				bytes = unsafe.Add(bytes, 48)
			}
			seed ^= see1 ^ see2
		}

		for index > 16 {
			seed = wymix(read64(bytes)^secret[1], read64(unsafe.Add(bytes, 8))^seed)
			index -= 16
			bytes = unsafe.Add(bytes, 16)
		}
		a = read64(unsafe.Add(bytes, index-16))
		b = read64(unsafe.Add(bytes, index-8))
	}

	a ^= secret[1]
	b ^= seed
	wymum(&a, &b)
	return wymix(a^secret[0]^length, b^secret[1])
}

var wyf3Secret = [...]byte{
	15, 23, 27, 29, 30, 39, 43, 45, 46, 51, 53, 54, 57, 58, 60, 71, 75,
	77, 78, 83, 85, 86, 89, 90, 92, 99, 101, 102, 105, 106, 108, 113,
	114, 116, 120, 135, 139, 141, 142, 147, 149, 150, 153, 154, 156, 163,
	165, 166, 169, 170, 172, 177, 178, 180, 184, 195, 197, 198, 201, 202,
	204, 209, 210, 212, 216, 225, 226, 228, 232, 240}

func MakeSecret(seed uint64) [4]uint64 {
	var secret [4]uint64
	for i := 0; i < 4; i++ {
		var ok bool

		for !ok {
			ok = true
			secret[i] = 0

			for j := 0; j < 64; j += 8 {
				secret[i] |= uint64(wyf3Secret[wyrand(&seed)%uint64(len(wyf3Secret))]) << j
			}
			if secret[i]%2 == 0 {
				ok = false
				continue
			}
			for j := 0; j < i; j++ {
				if popcnt64(secret[j]^secret[i]) != 32 {
					ok = false
					break
				}

				//x := secret[j] ^ secret[i]
				//x -= (x >> 1) & 0x5555555555555555
				//x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
				//x = (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
				//x = (x * 0x0101010101010101) >> 56
				//if x != 32 {
				//	ok = false
				//	break
				//}
			}
		}
	}
	return secret
}

func popcnt64(x uint64) uint64 {
	x = (x & 0x5555555555555555) + ((x & 0xAAAAAAAAAAAAAAAA) >> 1)
	x = (x & 0x3333333333333333) + ((x & 0xCCCCCCCCCCCCCCCC) >> 2)
	x = (x & 0x0F0F0F0F0F0F0F0F) + ((x & 0xF0F0F0F0F0F0F0F0) >> 4)
	x *= 0x0101010101010101
	return (x >> 56) & 0xFF
}
