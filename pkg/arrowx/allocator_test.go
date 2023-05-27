package arrowx

import (
	"fmt"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/math"
	"testing"
)

func TestOffHeap_Allocate(t *testing.T) {
	fb := array.NewFloat64Builder(OffHeap)

	fb.AppendValues([]float64{1, 3, 5, 7, 9, 11}, nil)

	vec := fb.NewFloat64Array()

	fmt.Println(math.Float64.Sum(vec))
}
