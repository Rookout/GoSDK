//go:build go1.15 && !go1.21
// +build go1.15,!go1.21

package suspender

import (
	"sync"
	_ "unsafe"
)

//go:linkname stopWorld runtime.stopTheWorldGC
func stopWorld(reason string)

//go:linkname startWorld runtime.startTheWorldGC
func startWorld()

type suspender struct {
	isSuspended bool
}


func (s *suspender) StopAll() {
	if s.isSuspended {
		
		return
	}
	
	stopWorld("")
	s.isSuspended = true
}

func (s *suspender) ResumeAll() {
	if !s.isSuspended {
		return
	}
	
	s.isSuspended = false
	startWorld()
}

func (s *suspender) Stopped() bool {
	
	return s.isSuspended
}

var creationLock sync.Mutex
var theOnlySuspender *suspender

func GetSuspender() Suspender {
	if theOnlySuspender == nil {
		
		creationLock.Lock()
		defer creationLock.Unlock()
		if theOnlySuspender == nil {
			
			theOnlySuspender = &suspender{isSuspended: false}
		}
	}

	return theOnlySuspender
}
