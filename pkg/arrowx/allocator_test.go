package arrowx

import (
	"context"
	"fmt"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/ipc"
	"github.com/apache/arrow/go/v13/arrow/math"
	"github.com/apache/arrow/go/v13/parquet"
	"github.com/apache/arrow/go/v13/parquet/compress"
	"github.com/apache/arrow/go/v13/parquet/file"
	"github.com/apache/arrow/go/v13/parquet/pqarrow"
	"os"
	"testing"
)

func TestOffHeap_Allocate(t *testing.T) {
	fb := array.NewFloat64Builder(OffHeap)

	fb.AppendValues([]float64{1, 3, 5, 7, 9, 11}, nil)

	vec := fb.NewFloat64Array()

	fmt.Println(math.Float64.Sum(vec))
}

func TestParquet(t *testing.T) {
	f, err := os.OpenFile("/Users/cmo/data/tickers/ES/ESZ22_TZ.ipc", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ipc.NewFileReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	s := r.Schema()
	fmt.Println(s)

	out, err := os.OpenFile("/Users/cmo/data/tickers/ES/ESZ22_TZ.parquet", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		out.Sync()
		out.Close()
	}()

	w, err := pqarrow.NewFileWriter(s, out,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Zstd),
			//parquet.WithCompression(compress.Codecs.Brotli),
			parquet.WithAllocator(OffHeap),
			parquet.WithCompressionLevel(5),
		),
		pqarrow.DefaultWriterProps())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	//w.NewRowGroup()
	numRecords := r.NumRecords()
	for i := 0; i < numRecords; i++ {
		rec, err := r.Record(i)
		if err != nil {
			t.Fatal(err)
		}
		err = w.Write(rec)
		if err != nil {
			t.Fatal(err)
		}
	}

	fmt.Println("Wrote", numRecords, "Records")
}

func TestParquetReader(t *testing.T) {
	pf, err := file.OpenParquetFile("/Users/cmo/data/tickers/ES/ESZ22_TZ.parquet", true)
	if err != nil {
		t.Fatal(err)
	}
	defer pf.Close()

	fmt.Println("Rows", pf.NumRows())
	fmt.Println("RowGroups", pf.NumRowGroups())

	rg := pf.RowGroup(0)
	fmt.Println("ByteSize", rg.ByteSize())

	fr, err := pqarrow.NewFileReader(pf, pqarrow.ArrowReadProperties{
		Parallel:  false,
		BatchSize: 1024 * 512,
	}, OffHeap)

	if err != nil {
		t.Fatal(err)
	}

	cr, err := fr.GetColumn(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := cr.NextBatch(900)
	if err != nil {
		t.Fatal(err)
	}

	c := chunk.Chunk(0)
	fmt.Println(c.GetOneForMarshal(0))
	chunk, err = cr.NextBatch(900)
	c = chunk.Chunk(0)
	fmt.Println(c.GetOneForMarshal(0))

	//table, err := fr.ReadTable(context.Background())
	//if err != nil {
	//	t.Fatal(err)
	//}
	//_ = table

	//cr, err := rg.Column(0)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(rg.MetaData())
}
