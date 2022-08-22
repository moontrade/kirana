package mmap

import (
	"os"
	"testing"
)

func TestMMap(t *testing.T) {
	_ = os.Remove("testdata/1.dat")
	f, err := os.Create("testdata/1.dat")
	if err != nil {
		t.Fatal(err)
	}
	m, err := MapRegion(f, 1024*128, RDWR, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Truncate(1000)
	if err != nil {
		t.Fatal(err)
	}
	m[0] = 'a'
	m.Flush()
	f.Sync()
	m.Unmap()
	f.Close()
}

func BenchmarkTruncate(b *testing.B) {
	_ = os.Remove("testdata/1.dat")
	f, err := os.Create("testdata/1.dat")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = f.Truncate(int64(i % 16384))
		if err != nil {
			b.Fatal(err)
		}
	}
}
