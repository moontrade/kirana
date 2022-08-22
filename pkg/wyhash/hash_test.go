package wyhash

import (
	"encoding/binary"
	"fmt"
	"github.com/bytedance/gopkg/util/xxhash3"
	"hash/crc32"
	"hash/crc64"
	"math/rand"
	"testing"
	"time"
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
	//println(wymum(5000000000, 11))
	//SetSeed(uint64(time.Now().UnixNano()))
	//fmt.Println(next())
	//for i := 0; i < 10; i++ {
	//	fmt.Println(NextFloat())
	//}
	//fmt.Println(NextGaussian())
	//fmt.Println(wymum(10, 11), wymum2(10, 11))
	//fmt.Println(wymum(192923, 9877732), wymum2(192923, 9877732))
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

func crc64h(v uint32) uint32 {
	h := crc64.New(crc64.MakeTable(crc64.ECMA))
	h.Reset()
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[0:], v)
	h.Write(b[0:4])
	return uint32(h.Sum64())
}

func crc32h(v uint32) uint32 {
	h := crc32.NewIEEE()
	h.Reset()
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[0:], v)
	h.Write(b[0:4])
	return h.Sum32()
}

var crc16tab = [256]uint16{
	0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
	0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
	0x1231, 0x0210, 0x3273, 0x2252, 0x52b5, 0x4294, 0x72f7, 0x62d6,
	0x9339, 0x8318, 0xb37b, 0xa35a, 0xd3bd, 0xc39c, 0xf3ff, 0xe3de,
	0x2462, 0x3443, 0x0420, 0x1401, 0x64e6, 0x74c7, 0x44a4, 0x5485,
	0xa56a, 0xb54b, 0x8528, 0x9509, 0xe5ee, 0xf5cf, 0xc5ac, 0xd58d,
	0x3653, 0x2672, 0x1611, 0x0630, 0x76d7, 0x66f6, 0x5695, 0x46b4,
	0xb75b, 0xa77a, 0x9719, 0x8738, 0xf7df, 0xe7fe, 0xd79d, 0xc7bc,
	0x48c4, 0x58e5, 0x6886, 0x78a7, 0x0840, 0x1861, 0x2802, 0x3823,
	0xc9cc, 0xd9ed, 0xe98e, 0xf9af, 0x8948, 0x9969, 0xa90a, 0xb92b,
	0x5af5, 0x4ad4, 0x7ab7, 0x6a96, 0x1a71, 0x0a50, 0x3a33, 0x2a12,
	0xdbfd, 0xcbdc, 0xfbbf, 0xeb9e, 0x9b79, 0x8b58, 0xbb3b, 0xab1a,
	0x6ca6, 0x7c87, 0x4ce4, 0x5cc5, 0x2c22, 0x3c03, 0x0c60, 0x1c41,
	0xedae, 0xfd8f, 0xcdec, 0xddcd, 0xad2a, 0xbd0b, 0x8d68, 0x9d49,
	0x7e97, 0x6eb6, 0x5ed5, 0x4ef4, 0x3e13, 0x2e32, 0x1e51, 0x0e70,
	0xff9f, 0xefbe, 0xdfdd, 0xcffc, 0xbf1b, 0xaf3a, 0x9f59, 0x8f78,
	0x9188, 0x81a9, 0xb1ca, 0xa1eb, 0xd10c, 0xc12d, 0xf14e, 0xe16f,
	0x1080, 0x00a1, 0x30c2, 0x20e3, 0x5004, 0x4025, 0x7046, 0x6067,
	0x83b9, 0x9398, 0xa3fb, 0xb3da, 0xc33d, 0xd31c, 0xe37f, 0xf35e,
	0x02b1, 0x1290, 0x22f3, 0x32d2, 0x4235, 0x5214, 0x6277, 0x7256,
	0xb5ea, 0xa5cb, 0x95a8, 0x8589, 0xf56e, 0xe54f, 0xd52c, 0xc50d,
	0x34e2, 0x24c3, 0x14a0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
	0xa7db, 0xb7fa, 0x8799, 0x97b8, 0xe75f, 0xf77e, 0xc71d, 0xd73c,
	0x26d3, 0x36f2, 0x0691, 0x16b0, 0x6657, 0x7676, 0x4615, 0x5634,
	0xd94c, 0xc96d, 0xf90e, 0xe92f, 0x99c8, 0x89e9, 0xb98a, 0xa9ab,
	0x5844, 0x4865, 0x7806, 0x6827, 0x18c0, 0x08e1, 0x3882, 0x28a3,
	0xcb7d, 0xdb5c, 0xeb3f, 0xfb1e, 0x8bf9, 0x9bd8, 0xabbb, 0xbb9a,
	0x4a75, 0x5a54, 0x6a37, 0x7a16, 0x0af1, 0x1ad0, 0x2ab3, 0x3a92,
	0xfd2e, 0xed0f, 0xdd6c, 0xcd4d, 0xbdaa, 0xad8b, 0x9de8, 0x8dc9,
	0x7c26, 0x6c07, 0x5c64, 0x4c45, 0x3ca2, 0x2c83, 0x1ce0, 0x0cc1,
	0xef1f, 0xff3e, 0xcf5d, 0xdf7c, 0xaf9b, 0xbfba, 0x8fd9, 0x9ff8,
	0x6e17, 0x7e36, 0x4e55, 0x5e74, 0x2e93, 0x3eb2, 0x0ed1, 0x1ef0}

func crc16a(v uint32) uint32 {
	crc := uint16(0)
	crc = ((crc << 8) & 0xff00) ^ crc16tab[((crc>>8)&0xff)^uint16(byte(v))]
	crc = ((crc << 8) & 0xff00) ^ crc16tab[((crc>>8)&0xff)^uint16(byte(v<<8))]
	crc = ((crc << 8) & 0xff00) ^ crc16tab[((crc>>8)&0xff)^uint16(byte(v<<16))]
	crc = ((crc << 8) & 0xff00) ^ crc16tab[((crc>>8)&0xff)^uint16(byte(v<<24))]
	return uint32(crc)
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

BenchmarkXXHash
BenchmarkXXHash/Hash_3
BenchmarkXXHash/Hash_3-8         	361316500	         3.037 ns/op
BenchmarkXXHash/Hash_5
BenchmarkXXHash/Hash_5-8         	346691133	         3.440 ns/op
BenchmarkXXHash/Hash_8
BenchmarkXXHash/Hash_8-8         	338533942	         3.454 ns/op
BenchmarkXXHash/Hash_16
BenchmarkXXHash/Hash_16-8        	362716063	         3.327 ns/op
BenchmarkXXHash/Hash_32
BenchmarkXXHash/Hash_32-8        	265428247	         4.546 ns/op
BenchmarkXXHash/Hash_64
BenchmarkXXHash/Hash_64-8        	197559765	         6.129 ns/op
BenchmarkXXHash/Hash_128
BenchmarkXXHash/Hash_128-8       	129044548	         9.265 ns/op
BenchmarkXXHash/Hash_129
BenchmarkXXHash/Hash_129-8       	122969744	         9.946 ns/op
BenchmarkXXHash/Hash_256
BenchmarkXXHash/Hash_256-8       	48000000	        25.27 ns/op
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
	//b.Run("Hash U64", func(b *testing.B) {
	//	b.ResetTimer()
	//	for i := 0; i < b.N; i++ {
	//		U64(uint64(i + 1))
	//	}
	//})
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
