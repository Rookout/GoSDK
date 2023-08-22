//go:build go1.21 && !go1.22
// +build go1.21,!go1.22

package go_runtime

import (
	_ "unsafe"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
)

type GPtr uintptr


type guintptr uintptr
type unwindFlags uint8

const (
	
	
	
	
	
	
	
	
	
	
	
	
	unwindPrintErrors unwindFlags = 1 << iota

	
	unwindSilentErrors

	
	
	
	
	
	
	
	
	
	
	unwindTrap

	
	
	
	unwindJumpStack
)

type stkframe struct {
	
	
	fn module.FuncInfo

	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	pc uintptr

	
	
	
	
	
	
	
	
	continpc uintptr

	lr   uintptr 
	sp   uintptr 
	fp   uintptr 
	varp uintptr 
	argp uintptr 
}
type unwinder struct {
	
	
	frame stkframe

	
	
	
	g guintptr

	
	
	cgoCtxt int

	
	
	calleeFuncID module.FuncID

	
	
	flags unwindFlags

	
	cache module.PCValueCache
}


func Getg() GPtr

//go:linkname initAt runtime.(*unwinder).initAt
func initAt(u *unwinder, pc0, sp0, lr0 uintptr, gp GPtr, flags unwindFlags)

//go:linkname tracebackPCs runtime.tracebackPCs
func tracebackPCs(u *unwinder, skip int, pcBuf []uintptr) int


func Callers(pc uintptr, sp uintptr, gp GPtr, pcbuf []uintptr) int {
	var u unwinder
	initAt(&u, pc, sp, 0, gp, unwindSilentErrors)
	return tracebackPCs(&u, 0, pcbuf)
}
