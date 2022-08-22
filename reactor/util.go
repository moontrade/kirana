package reactor

import "unsafe"

func PollToPollFnPointer(future Future) unsafe.Pointer {
	var fn = future.Poll
	return *(*unsafe.Pointer)(unsafe.Pointer(&fn))
}

func PollFnPointer(poll func(event Context) error) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&poll))
}
