package aof

import (
	"fmt"
	"github.com/moontrade/wormhole/reactor"
	"testing"
	"time"
)

func init() {
	reactor.Init(1, reactor.Millis250, 8192, 8192)
}

func TestTailer(t *testing.T) {
	//u, err := uid.CryptoU64()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//b := make([]byte, 8)
	//binary.LittleEndian.PutUint64(b, u)
	//fmt.Println(b)
	m, err := NewManager("testdata", 0755)
	if err != nil {
		t.Fatal(err)
	}
	f, err := m.Open("db.txt", OpenFile, RecoveryDefault)
	if err != nil {
		t.Fatal(err)
	}
	tailer, err := f.Subscribe(&Reader{})
	if err != nil {
		t.Fatal(err)
	}
	_ = tailer

	for {
		time.Sleep(time.Second)
		_, _ = f.Write([]byte{'a', 'b', 'c', 'd'})
		_ = f.Flush()

		err = f.Finish()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(m.stats)
		break
	}

	f.Close()

	time.Sleep(time.Hour)
}

type Reader struct {
	Consumer
}

func (r *Reader) PollRead(event ReadEvent) (int64, error) {
	fmt.Println("read:", event.Begin, event.End, event.EOF)
	if event.EOF {
		return 0, reactor.ErrStop
	}
	return event.End, nil
}

func (r *Reader) PollReadClosed(reason error) {
	fmt.Println("reader closed", reason)
}
