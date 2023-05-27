package wyhash

import (
	"encoding/hex"
	"fmt"
	"github.com/bytedance/gopkg/util/xxhash3"
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/minio/highwayhash"
)

func print_hash(s string) {
	fmt.Printf("%s: %d\n", s, String(s))
}

func TestMul64(t *testing.T) {
	print_hash("h")
	print_hash("he")
	print_hash("hel")
	print_hash("hell")
	print_hash("hello")
	print_hash("hellonow")
	print_hash("hellonowhellonow")
	print_hash("hellonowhellonowhellonowhellonow")
	print_hash("hellonowhellonowhellonowhellonowhellonowhellonowhellonowhellonow")

	//println(U64(10))
	//println(U64(11))
	//println(wymix(5000000000, 11))
	//SetSeed(uint64(time.Now().UnixNano()))
	//fmt.Println(next())
	//for i := 0; i < 10; i++ {
	//	fmt.Println(NextFloat())
	//}
	//fmt.Println(NextGaussian())
	//fmt.Println(wymix(10, 11), wymum2(10, 11))
	//fmt.Println(wymix(192923, 9877732), wymum2(192923, 9877732))
	//fmt.Println(1 ^ uint64(0xe7037ed1a0b428db))
	//fmt.Println(99 ^ uint64(0xe7037ed1a0b428db))
	//fmt.Println(HashString("hel"))
	//fmt.Println(HashString("hell"))
	//fmt.Println(HashString("hello"))
	//fmt.Println(HashString("hello there today ok"))
	//fmt.Println(String("hello there today ok hello there today ok hello there today ok o hello there today ok hello there today ok hello there today ok o"))
	//fmt.Println(HashString("hello"))
	//fmt.Println(HashString("hello123"))
	//fmt.Println(uint64(10) >> 2)
}

func TestHashCollisions(t *testing.T) {
	var (
		//c32   int
		//a32   int
		//c16   int
		fn           int
		xx           int
		wyf3         int
		wyf3string3  int
		wyf3string4  int
		wyf3string5  int
		wyf3string8  int
		wyf3string20 int
		wyf3string32 int
		wyf3string64 int
		total        int
	)

	type config struct {
		low            int
		high           int
		adder          int
		factor         float64
		addressStart   int
		multiplierLow  int
		multiplierHigh int
		multiplierAdd  int
	}
	for _, cfg := range []config{
		//{2, 10, 4, 8, 65580, 64, 4096, 128},
		// WASM like pointer values
		{40, 1024, 88, 3, 65580, 512, 256000, 96},
		{1024, 4096, 512, 3, 65580, 56, 52050, 512},
	} {
		for i := cfg.low; i < cfg.high; i += cfg.adder {
			var (
				entries = i
				slots   = int(float64(i) * cfg.factor)
			)
			for multiplier := cfg.multiplierLow; multiplier < cfg.multiplierHigh; multiplier += cfg.multiplierAdd {
				total += entries
				//c32 += testCollisions(entries, multiplier, slots, crc32h)
				//c16 += testCollisions(entries, multiplier, slots, crc16a)
				fn += testCollisions(entries, multiplier, slots, FNV32)
				wyf3 += testCollisions64(entries, multiplier, slots, U64)
				xx += testCollisionsString(entries, 8, slots, String)
				wyf3string3 += testCollisionsString(entries, 3, slots, String)
				wyf3string4 += testCollisionsString(entries, 4, slots, String)
				wyf3string5 += testCollisionsString(entries, 5, slots, String)
				wyf3string8 += testCollisionsString(entries, 8, slots, String)
				wyf3string20 += testCollisionsString(entries, 20, slots, String)
				wyf3string32 += testCollisionsString(entries, 32, slots, String)
				wyf3string64 += testCollisionsString(entries, 64, slots, String)
			}
		}
	}

	println("")
	println("total		", total)
	//println("\tcrc32		", c32)
	//println("\tcrc16		", c16)
	println("\tfnv			", fn)
	println("\txx			", xx)
	println("\tWYF3		", wyf3)
	println("\tWYF3Str 3		", wyf3string3)
	println("\tWYF3Str 4		", wyf3string4)
	println("\tWYF3Str 5		", wyf3string5)
	println("\tWYF3Str 8		", wyf3string8)
	println("\tWYF3Str 20		", wyf3string20)
	println("\tWYF3Str 32		", wyf3string32)
	println("\tWYF3Str 64		", wyf3string64)
}

func rangeRandom(min, max uint32) uint32 {
	return uint32(rand.Int31n(int32(max-min)) + int32(min))
}

func testCollisions(entries, multiplier, slots int, hasher func(uint32) uint32) int {
	m := make(map[uint32]struct{})
	count := 0

	ptr := 65680

	for i := 0; i < entries; i++ {
		v := hasher(uint32(ptr))
		ptr += multiplier
		index := v % uint32(slots)

		_, ok := m[index]
		if ok {
			count++
		} else {
			m[index] = struct{}{}
		}
	}
	return count
}

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func testCollisionsString(entries, size, slots int, hasher func(b string) uint64) int {
	m := make(map[uint64]struct{})
	count := 0

	for i := 0; i < entries; i++ {
		s := StringWithCharset(size, charset)
		v := hasher(s)
		index := v % uint64(slots)

		_, ok := m[index]
		if ok {
			count++
		} else {
			m[index] = struct{}{}
		}
	}
	return count
}

func testCollisions64(entries, multiplier, slots int, hasher func(uint64) uint64) int {
	m := make(map[uint64]struct{})
	count := 0

	ptr := 65536

	for i := uint64(0); i < uint64(entries); i++ {
		v := hasher(uint64(ptr))
		ptr += multiplier
		index := v % uint64(slots)

		_, ok := m[index]
		if ok {
			count++
		} else {
			m[index] = struct{}{}
		}
	}
	return count
}

func TestU32(t *testing.T) {
	v := uint32(565498879)
	fmt.Println(U32(v))
	bytes := *(*[4]byte)(unsafe.Pointer(&v))
	fmt.Println(String(string(bytes[:])))
}

func TestU64(t *testing.T) {
	v := uint64(5654654988749)
	fmt.Println(U64(v))
	bytes := *(*[8]byte)(unsafe.Pointer(&v))
	fmt.Println(String(string(bytes[:])))
}

func BenchmarkOptimized(b *testing.B) {
	b.Run("Optimized U8", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			U8(byte(i + 1))
		}
	})
	b.Run("Unoptimized U8", func(b *testing.B) {
		b.ResetTimer()
		var v byte
		for i := 0; i < b.N; i++ {
			v = byte(i)
			Hash(unsafe.Pointer(&v), 1)
		}
	})
	b.Run("Optimized U16", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			U16(uint16(i + 1))
		}
	})
	b.Run("Unoptimized U16", func(b *testing.B) {
		b.ResetTimer()
		var v uint16
		for i := 0; i < b.N; i++ {
			v = uint16(i)
			Hash(unsafe.Pointer(&v), 2)
		}
	})
	b.Run("Optimized U32", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			U32(uint32(i + 1))
		}
	})
	b.Run("Unoptimized U32", func(b *testing.B) {
		b.ResetTimer()
		for i := int32(0); i < int32(b.N); i++ {
			Hash(unsafe.Pointer(&i), 4)
		}
	})
	b.Run("Optimized U64", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			U64(uint64(i + 1))
		}
	})
	b.Run("Unoptimized U64", func(b *testing.B) {
		b.ResetTimer()
		for i := int64(0); i < int64(b.N); i++ {
			Hash(unsafe.Pointer(&i), 8)
		}
	})
}

func BenchmarkHash(b *testing.B) {
	const multiply = uint64(1)
	seed := rand.Uint64()

	//FNV64(FNV64((11 + 1) * seed))
	String(StringWithCharset(128, charset))
	String(StringWithCharset(3, charset))
	String(StringWithCharset(4, charset))
	String(StringWithCharset(16, charset))
	String(StringWithCharset(32, charset))
	String(StringWithCharset(51, charset))
	String(StringWithCharset(64, charset))
	String(StringWithCharset(96, charset))
	String(StringWithCharset(128, charset))
	String(StringWithCharset(256, charset))
	U64(U64((11 + 1) * seed))

	//b.Invoke("crc32", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc32h(23)
	//	}
	//})
	//b.Invoke("crc16", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc16a(23)
	//	}
	//})
	//b.Invoke("crc16", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc16a(23)
	//	}
	//})
	//b.Invoke("FNV64a", func(b *testing.B) {
	//	for i := uint64(0); i < uint64(b.N)*multiply; i++ {
	//		FNV32a(uint32((i + 1) * seed))
	//	}
	//})
	b.Run("Hash U64", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			U64(uint64(i + 1))
		}
	})
	b.Run("Hash 3", func(b *testing.B) {
		str := "hel"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 5", func(b *testing.B) {
		str := "hello"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 8", func(b *testing.B) {
		str := "hellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 9", func(b *testing.B) {
		str := "hellobyeh"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 11", func(b *testing.B) {
		str := "hellobyehel"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 13", func(b *testing.B) {
		str := "hellobyehello"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 16", func(b *testing.B) {
		str := "hellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 32", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 64", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 128", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 129", func(b *testing.B) {
		str := "hello there today ok hello there today ok hello there today ok o hello there today ok hello there today ok hello there today ok o"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
	b.Run("Hash 256", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			String(str)
		}
	})
}

/*
WyhashF3
BenchmarkHash/Hash_U64
BenchmarkHash/Hash_U64-8         	1000000000	         0.3251 ns/op
BenchmarkHash/Hash_3
BenchmarkHash/Hash_3-8           	512882978	         2.314 ns/op
BenchmarkHash/Hash_5
BenchmarkHash/Hash_5-8           	557430321	         2.170 ns/op
BenchmarkHash/Hash_8
BenchmarkHash/Hash_8-8           	557687008	         2.150 ns/op
BenchmarkHash/Hash_16
BenchmarkHash/Hash_16-8          	555317316	         2.160 ns/op
BenchmarkHash/Hash_32
BenchmarkHash/Hash_32-8          	410573053	         2.942 ns/op
BenchmarkHash/Hash_64
BenchmarkHash/Hash_64-8          	305601321	         3.987 ns/op
BenchmarkHash/Hash_128
BenchmarkHash/Hash_128-8         	190449877	         6.760 ns/op
BenchmarkHash/Hash_129
BenchmarkHash/Hash_129-8         	157599220	         7.098 ns/op
BenchmarkHash/Hash_256
BenchmarkHash/Hash_256-8         	100000000	        10.83 ns/op

WyhashF4
BenchmarkHash
BenchmarkHash/Hash_U64
BenchmarkHash/Hash_U64-8         	1000000000	         0.3195 ns/op
BenchmarkHash/Hash_3
BenchmarkHash/Hash_3-8           	389462864	         3.056 ns/op
BenchmarkHash/Hash_5
BenchmarkHash/Hash_5-8           	376605081	         3.170 ns/op
BenchmarkHash/Hash_8
BenchmarkHash/Hash_8-8           	373160917	         3.177 ns/op
BenchmarkHash/Hash_9
BenchmarkHash/Hash_9-8           	375572551	         3.184 ns/op
BenchmarkHash/Hash_11
BenchmarkHash/Hash_11-8          	375357028	         3.184 ns/op
BenchmarkHash/Hash_13
BenchmarkHash/Hash_13-8          	375741795	         3.188 ns/op
BenchmarkHash/Hash_16
BenchmarkHash/Hash_16-8          	353291739	         3.181 ns/op
BenchmarkHash/Hash_32
BenchmarkHash/Hash_32-8          	328011498	         3.647 ns/op
BenchmarkHash/Hash_64
BenchmarkHash/Hash_64-8          	254052578	         4.729 ns/op
BenchmarkHash/Hash_128
BenchmarkHash/Hash_128-8         	175286341	         6.813 ns/op
BenchmarkHash/Hash_129
BenchmarkHash/Hash_129-8         	160516773	         7.454 ns/op
BenchmarkHash/Hash_256
BenchmarkHash/Hash_256-8         	100000000	        11.08 ns/op

WyhashF4 Default Optimized
BenchmarkHash
BenchmarkHash/Hash_U64
BenchmarkHash/Hash_U64-8         	541850214	         2.227 ns/op
BenchmarkHash/Hash_3
BenchmarkHash/Hash_3-8           	439639683	         2.712 ns/op
BenchmarkHash/Hash_5
BenchmarkHash/Hash_5-8           	425250298	         2.842 ns/op
BenchmarkHash/Hash_8
BenchmarkHash/Hash_8-8           	424785966	         2.849 ns/op
BenchmarkHash/Hash_9
BenchmarkHash/Hash_9-8           	424775502	         2.842 ns/op
BenchmarkHash/Hash_11
BenchmarkHash/Hash_11-8          	420177139	         2.854 ns/op
BenchmarkHash/Hash_13
BenchmarkHash/Hash_13-8          	421709115	         2.844 ns/op
BenchmarkHash/Hash_16
BenchmarkHash/Hash_16-8          	421790827	         2.845 ns/op
BenchmarkHash/Hash_32
BenchmarkHash/Hash_32-8          	349303792	         3.433 ns/op
BenchmarkHash/Hash_64
BenchmarkHash/Hash_64-8          	248358012	         4.831 ns/op
BenchmarkHash/Hash_128
BenchmarkHash/Hash_128-8         	174078080	         6.945 ns/op
BenchmarkHash/Hash_129
BenchmarkHash/Hash_129-8         	155284183	         7.707 ns/op
BenchmarkHash/Hash_256
BenchmarkHash/Hash_256-8         	100000000	        11.39 ns/op

BenchmarkXXHash
BenchmarkXXHash/Hash_3
BenchmarkXXHash/Hash_3-8         	356664517	         3.063 ns/op
BenchmarkXXHash/Hash_5
BenchmarkXXHash/Hash_5-8         	367344314	         3.250 ns/op
BenchmarkXXHash/Hash_8
BenchmarkXXHash/Hash_8-8         	372347170	         3.209 ns/op
BenchmarkXXHash/Hash_16
BenchmarkXXHash/Hash_16-8        	372898899	         3.201 ns/op
BenchmarkXXHash/Hash_32
BenchmarkXXHash/Hash_32-8        	270837207	         4.435 ns/op
BenchmarkXXHash/Hash_64
BenchmarkXXHash/Hash_64-8        	200117971	         6.011 ns/op
BenchmarkXXHash/Hash_128
BenchmarkXXHash/Hash_128-8       	128223158	         9.281 ns/op
BenchmarkXXHash/Hash_129
BenchmarkXXHash/Hash_129-8       	120736381	         9.899 ns/op
BenchmarkXXHash/Hash_256
BenchmarkXXHash/Hash_256-8       	51440235	        22.55 ns/op
*/

func BenchmarkXXHash(b *testing.B) {
	const multiply = uint64(1)
	seed := rand.Uint64()

	//FNV64(FNV64((11 + 1) * seed))
	String(StringWithCharset(128, charset))
	String(StringWithCharset(3, charset))
	String(StringWithCharset(4, charset))
	String(StringWithCharset(16, charset))
	String(StringWithCharset(32, charset))
	String(StringWithCharset(51, charset))
	String(StringWithCharset(64, charset))
	String(StringWithCharset(96, charset))
	String(StringWithCharset(128, charset))
	String(StringWithCharset(256, charset))
	U64(U64((11 + 1) * seed))

	//b.Invoke("crc32", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc32h(23)
	//	}
	//})
	//b.Invoke("crc16", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc16a(23)
	//	}
	//})
	//b.Invoke("crc16", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc16a(23)
	//	}
	//})
	//b.Invoke("FNV64a", func(b *testing.B) {
	//	for i := uint64(0); i < uint64(b.N)*multiply; i++ {
	//		FNV32a(uint32((i + 1) * seed))
	//	}
	//})
	b.Run("Hash U64", func(b *testing.B) {
		b.ResetTimer()
		for i := uint64(0); i < uint64(b.N); i++ {
			buf := *(*[8]byte)(unsafe.Pointer(&i))
			xxhash3.Hash(buf[:])
		}
	})
	b.Run("Hash 3", func(b *testing.B) {
		str := "hel"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 5", func(b *testing.B) {
		str := "hello"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 8", func(b *testing.B) {
		str := "hellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 16", func(b *testing.B) {
		str := "hellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 32", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 64", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 128", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 129", func(b *testing.B) {
		str := "hello there today ok hello there today ok hello there today ok o hello there today ok hello there today ok hello there today ok o"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
	b.Run("Hash 256", func(b *testing.B) {
		str := "hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xxhash3.HashString(str)
		}
	})
}

var hhkey []byte

func init() {
	var err error
	hhkey, err = hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	if err != nil {
		panic(err)
	}
}

func BenchmarkHighwayHash(b *testing.B) {
	const multiply = uint64(1)
	seed := rand.Uint64()

	//FNV64(FNV64((11 + 1) * seed))
	String(StringWithCharset(128, charset))
	String(StringWithCharset(3, charset))
	String(StringWithCharset(4, charset))
	String(StringWithCharset(16, charset))
	String(StringWithCharset(32, charset))
	String(StringWithCharset(51, charset))
	String(StringWithCharset(64, charset))
	String(StringWithCharset(96, charset))
	String(StringWithCharset(128, charset))
	String(StringWithCharset(256, charset))
	U64(U64((11 + 1) * seed))

	//b.Invoke("crc32", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc32h(23)
	//	}
	//})
	//b.Invoke("crc16", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc16a(23)
	//	}
	//})
	//b.Invoke("crc16", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		crc16a(23)
	//	}
	//})
	//b.Invoke("FNV64a", func(b *testing.B) {
	//	for i := uint64(0); i < uint64(b.N)*multiply; i++ {
	//		FNV32a(uint32((i + 1) * seed))
	//	}
	//})
	//b.Run("Hash U64", func(b *testing.B) {
	//	b.ResetTimer()
	//	for i := 0; i < b.N; i++ {
	//		U64(uint64(i + 1))
	//	}
	//})
	b.Run("Hash 3", func(b *testing.B) {
		str := []byte("hel")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 5", func(b *testing.B) {
		str := []byte("hello")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 8", func(b *testing.B) {
		str := []byte("hellobye")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 16", func(b *testing.B) {
		str := []byte("hellobyehellobye")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 32", func(b *testing.B) {
		str := []byte("hellobyehellobyehellobyehellobye")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 64", func(b *testing.B) {
		str := []byte("hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 128", func(b *testing.B) {
		str := []byte("hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 129", func(b *testing.B) {
		str := []byte("hello there today ok hello there today ok hello there today ok o hello there today ok hello there today ok hello there today ok o")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
	b.Run("Hash 256", func(b *testing.B) {
		str := []byte("hellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobyehellobye")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			highwayhash.Sum64(str, hhkey)
		}
	})
}
