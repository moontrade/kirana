package logger

import (
	"testing"
	"time"
)
import "fmt"
import "unsafe"

func TestRecordBuilder(t *testing.T) {
	AllocRecordBuilder().
		TimeHF().
		Str("field", "value").
		Msgf("test %s", "me")
}

type Field[T AttrType] struct {
	Name  string
	Value T
}

func (f *Field[T]) Kind() Kind {
	switch ((any)(f.Value)).(type) {
	case int64:
		return KindInt64
	}
	return 0
}

type AttrAny interface{ Field[string] | Field[int64] }

func field[T AttrType](f Field[T], count int) int {
	switch ((any)(f)).(type) {
	case Field[string]:
		count += 1
	case Field[int8]:
		count += 6
	case Field[int16]:
		count += 6
	case Field[int32]:
		count += 6
	case Field[time.Duration]:
		count += 9
	case Field[int64]:
		count += 3
	case Field[uint8]:
		count += 6
	case Field[uint16]:
		count += 6
	case Field[uint32]:
		count += 6
	case Field[uint64]:
		count += 3
	case Field[float32]:
		count += 6
	case Field[float64]:
		count += 3
	}
	return count
}

func sizeofField[T AttrType](f Field[T]) int {
	return int(unsafe.Sizeof(f))
}

func TestAttr(t *testing.T) {
	fmt.Println(sizeofField(Field[time.Duration]{}))
	fmt.Println(sizeofField(Field[string]{}))
	fmt.Println(unsafe.Sizeof(Attr[string]{}))
}

func BenchmarkField(b *testing.B) {
	f := Field[time.Duration]{}
	count := 0

	//field(f)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sizeofField(f)
		//count = field(f, count)
		//switch ((any)(f)).(type) {
		//case Field[string]:
		//	count += 1
		//case Field[int8]:
		//	count += 6
		//case Field[int16]:
		//	count += 6
		//case Field[int32]:
		//	count += 6
		//case Field[time.Duration]:
		//	count += 9
		//case Field[int64]:
		//	count += 3
		//case Field[uint8]:
		//	count += 6
		//case Field[uint16]:
		//	count += 6
		//case Field[uint32]:
		//	count += 6
		//case Field[uint64]:
		//	count += 3
		//case Field[float32]:
		//	count += 6
		//case Field[float64]:
		//	count += 3
		//}
	}
	b.StopTimer()
	fmt.Println(count)
}
