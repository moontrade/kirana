package reactor

import (
	"fmt"
	"github.com/moontrade/kirana/pkg/counter"
)

type SimpleTask struct {
	c *counter.Counter
}

func (t *SimpleTask) Poll(ctx Context) error {
	t.c.Incr()
	return nil
}

func (t *SimpleTask) PollClose(event CloseEvent) error {
	fmt.Println("SimpleTask closed")
	return nil
}

type OneShot struct {
	c *counter.Counter
}

func (t *OneShot) init() {}

func (t *OneShot) Poll(event Context) error {
	fmt.Println("one shot")
	return nil
}

func (t *OneShot) PollClose(event CloseEvent) error {
	fmt.Println("one shot closed")
	return nil
}
