package threadx

/*
#include <sys/types.h>
#include <pthread.h>

void kirana_get_thread_id(uint64_t thread_id, uint64_t b) {
	*(uint64_t*)(void*)thread_id = (uint64_t)pthread_getthreadid_np();
}
*/
import "C"
import (
	"github.com/moontrade/unsafe/cgo"
	"unsafe"
)

func CurrentThreadID() uint64 {
	tid := uint64(0)
	cgo.NonBlocking((*byte)(C.kirana_get_thread_id), uintptr(unsafe.Pointer(&tid)), 0)
	return tid
}
