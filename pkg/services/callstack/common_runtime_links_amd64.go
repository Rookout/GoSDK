//go:build amd64 && go1.16 && !go1.23
// +build amd64,go1.16,!go1.23

package callstack

import (
	"runtime"

	"github.com/Rookout/GoSDK/pkg/services/go_runtime"

	_ "unsafe"
)


const (
	
	
	
	
	
	
	
	
	
	
	
	
	// goroutines found in the run queue, rather than CAS-looping
	
	
	

	
	
	_Gidle = iota 

	
	
	_Grunnable 

	
	
	
	_Grunning 

	
	
	// goroutine. It is not on a run queue. It is assigned an M.
	_Gsyscall 

	
	
	
	
	
	
	
	// goroutine enters _Gwaiting (e.G., it may get moved).
	_Gwaiting 

	
	
	_Gmoribund_unused 

	
	
	
	
	
	
	_Gdead 

	
	_Genqueue_unused 

	
	
	
	_Gcopystack 

	
	
	
	
	
	_Gpreempted 

	
	
	// goroutine is not executing user code and the stack is owned
	
	
	
	
	
	
	
	
	_Gscan          = 0x1000
	_Gscanrunnable  = _Gscan + _Grunnable  
	_Gscanrunning   = _Gscan + _Grunning   
	_Gscansyscall   = _Gscan + _Gsyscall   
	_Gscanwaiting   = _Gscan + _Gwaiting   
	_Gscanpreempted = _Gscan + _Gpreempted 
)

//go:linkname readgstatus runtime.readgstatus
func readgstatus(g go_runtime.GPtr) uint32

//go:linkname isSystemGoroutine runtime.isSystemGoroutine
func isSystemGoroutine(g go_runtime.GPtr, fixed bool) bool

//go:linkname saveg runtime.saveg
func saveg(pc, sp uintptr, g go_runtime.GPtr, r *runtime.StackRecord)


var (
	globN          int
	globOk         bool
	globCurrentG   go_runtime.GPtr
	globStbView    []runtime.StackRecord
	globCurrentStb *StackTraceBuffer
)

func isOK(g1 go_runtime.GPtr) bool {
	return g1 != globCurrentG && readgstatus(g1) != _Gdead && !isSystemGoroutine(g1, false)
}

func countg(g1 go_runtime.GPtr) {
	if isOK(g1) {
		globN++
	}
}

func saveGoroutine(g1 go_runtime.GPtr) {
	if !isOK(g1) {
		return
	}

	if len(globStbView) == 0 {
		
		
		return
	}
	saveg(^uintptr(0), ^uintptr(0), g1, &globStbView[0])
	globCurrentStb.filled++
	globStbView = globStbView[1:]
}
