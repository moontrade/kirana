package threadx

import "golang.org/x/sys/windows"

func CurrentThreadID() uint64 {
	return uint64(windows.GetCurrentThreadId())
}
