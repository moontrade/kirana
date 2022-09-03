package stream

type Options struct {
	SegmentSize int32
	PageSize    int32
}

type Stream struct {
	Head *Segment
	Tail *Segment
}
