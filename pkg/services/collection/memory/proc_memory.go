package memory

import (
	"errors"
	"fmt"
	"unsafe"
)

type ProcMemory struct {
}

func (m *ProcMemory) ReadMemory(buf []byte, addr uint64) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			if panicErr, ok := r.(error); ok {
				err = panicErr
				return
			}
			err = fmt.Errorf("recovered from panic: %v", r)
			return
		}
	}()

	nMemcpy := len(buf)

	if nMemcpy == 0 {
		return 0, errors.New("buf size is 0")
	}

	//goland:noinspection GoVetUnsafePointer
	memcpy(unsafe.Pointer(&buf[0]), unsafe.Pointer(uintptr(addr)), nMemcpy)
	return nMemcpy, nil
}

func memcpy(dest, src unsafe.Pointer, len int) unsafe.Pointer {
	cnt := len >> 3
	var i = 0
	for i = 0; i < cnt; i++ {
		var pdest = (*uint64)(unsafe.Pointer(uintptr(dest) + uintptr(8*i)))
		var psrc = (*uint64)(unsafe.Pointer(uintptr(src) + uintptr(8*i)))
		*pdest = *psrc
	}
	left := len & 7
	for i = 0; i < left; i++ {
		var pdest = (*uint8)(unsafe.Pointer(uintptr(dest) + uintptr(8*cnt+i)))
		var psrc = (*uint8)(unsafe.Pointer(uintptr(src) + uintptr(8*cnt+i)))

		*pdest = *psrc
	}
	return dest
}

func (m *ProcMemory) ID() string {
	return "procmemory"
}
