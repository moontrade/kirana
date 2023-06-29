package timex

import "time"

var now int64

const precision = time.Millisecond * 250

func init() {
	now = Now()
	go pacer()
}

func Fastnow() int64 {
	return now
}

func pacer() {
	for {
		time.Sleep(precision)
		now = Now()
	}
}
