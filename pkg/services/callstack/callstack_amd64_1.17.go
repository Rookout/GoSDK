//go:build amd64 && go1.17 && !go1.23
// +build amd64,go1.17,!go1.23

package callstack

import (
	_ "unsafe"

	"github.com/Rookout/GoSDK/pkg/services/go_runtime"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
)

//go:linkname forEachGRace runtime.forEachGRace
func forEachGRace(fn func(g go_runtime.GPtr))

func (s *StackTraceBuffer) FillStackTraces() (int, bool) {
	if !suspender.GetSuspender().Stopped() {
		panic("You can't use this function while the world is not stopped! You must call StopAll() first!")
	}

	globCurrentG = go_runtime.Getg()
	globCurrentStb = s
	
	globCurrentStb.filled = 0
	globN = 0
	globOk = false

	
	forEachGRace(countg)
	if globN <= globCurrentStb.capacity() {
		globOk = true
		globStbView = globCurrentStb.buf

		
		
		forEachGRace(saveGoroutine)
	}
	return globN, globOk
}




















































