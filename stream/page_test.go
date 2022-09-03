package stream

import (
	"encoding/binary"
	"fmt"
	"github.com/moontrade/kirana/pkg/uid"
	"reflect"
	"testing"
)

func TestGenerateMagic(t *testing.T) {
	for i := 0; i < 5; i++ {
		v, err := uid.CryptoU64()
		if err != nil {
			t.Fatal(err)
		}
		printMagic(v)
	}
}

func printMagic(value uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[0:8], value)
	fmt.Println(b[0:8], value)
}

// BenchmarkDeref
// BenchmarkDeref/Static
// BenchmarkDeref/Static-8         	1000000000	         0.3142 ns/op
// BenchmarkDeref/Generic
// BenchmarkDeref/Generic-8        	1000000000	         0.9401 ns/op
func BenchmarkDeref(b *testing.B) {
	var (
		y pageTailPtr
		z deref[PageTail]
	)

	b.Run("Static", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			y.Deref()
		}
	})

	b.Run("Generic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			z.Deref()
		}
	})
}

func TestSegment(t *testing.T) {
	printSizeOf(Trade{})
	printSizeOf(Candle{})
	printSizeOf(CompactCandle{})
	printSizeOf(CompactCandleBA{})
}

func printSizeOf(v any) {
	s := reflect.ValueOf(v).Type().Size()
	fmt.Println(reflect.TypeOf(v).Name(), uint(s))
}

type Trade struct {
	Time        int64
	ExchangeID  [16]byte
	Price       float64
	Quantity    float64
	BidQuantity float64
	AskQuantity float64
	Aggressor   byte
	_           [7]byte
}

type OHLC struct {
	Open, High, Low, Close float64
}

type CompactOHLC struct {
	Open, High, Low, Close float32
}

type Candle struct {
	Time int64
	OHLC
	Volume, Buys, Sells float64
	NumTradesOrOI       int32
}

type CompactCandle struct {
	Time int64
	CompactOHLC
	Volume, Buys, Sells float32
	NumTradesOrOI       int32
}

type CompactCandleBA struct {
	Time                int64
	Bid                 CompactOHLC
	Ask                 CompactOHLC
	Volume, Buys, Sells float32
	NumTradesOrOI       int32
}

type CandleBA struct{}
