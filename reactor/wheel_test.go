package reactor

import (
	"fmt"
	"testing"
)

func TestWheel(t *testing.T) {
	w := NewWheel(Millis250)
	fmt.Println(w)
	sw := NewWheel(Seconds)
	fmt.Println(sw)
}
