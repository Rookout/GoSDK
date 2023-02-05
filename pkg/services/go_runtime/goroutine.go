package go_runtime

import (
	"unsafe"
)

type GPtr uintptr


type stkframe struct{}
type g struct{}


func Getg() GPtr

//go:linkname gentraceback runtime.gentraceback
func gentraceback(pc0, sp0, lr0 uintptr, gp *g, skip int, pcbuf *uintptr, max int, callback func(*stkframe, unsafe.Pointer) bool, v unsafe.Pointer, flags uint) int


func Callers(pc uintptr, sp uintptr, gp GPtr, pcbuf []uintptr) int {
	return gentraceback(pc, sp, 0, (*g)(unsafe.Pointer(gp)), 0, &pcbuf[0], len(pcbuf), nil, nil, 0)
}
