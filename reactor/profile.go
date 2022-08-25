package reactor

import (
	"github.com/moontrade/kirana/pkg/counter"
	"github.com/moontrade/kirana/pkg/cow"
)

type FuncStats struct {
	receiver     string
	name         string
	invokes      counter.Counter
	invokesDur   counter.TimeCounter
	panics       counter.Counter
	panicsBuffer cow.Slice[error]
}

type Func struct {
}

type FuncMap struct{}

type ObjectMap struct{}
