package stream

import "github.com/moontrade/kirana/aof"

const (
	SegmentSequenceBegin = 1000000
)

type Segment struct {
	streamID uint64
	sequence int64
	aof      *aof.AOF
}
