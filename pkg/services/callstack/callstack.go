package callstack

import (
	"github.com/Rookout/GoSDK/pkg/services/suspender"
	"math"
	"runtime"
)


const ReallocateSizeAuto int = 0

type IStackTraceBuffer interface {
	Reallocate(n int)                                      
	FillStackTraces() (n int, ok bool)                     
	Size() int                                             
	GetPC(goroutineIdx, depthIdx int) uintptr              
	GetDepth(goroutineIdx int) (depth int, allFrames bool) 
	GetMaxPossibleDepth() int                              
}

type StackTraceBuffer struct {
	buf    []runtime.StackRecord
	filled int
}

func NewStackTraceBuffer() IStackTraceBuffer {
	stb := &StackTraceBuffer{}
	stb.Reallocate(ReallocateSizeAuto)
	return stb
}

func (s *StackTraceBuffer) Reallocate(n int) {
	if suspender.GetSuspender().Stopped() {
		panic("Can't reallocate when the world is stopped! You must call ResumeAll() first!")
	}
	if n == ReallocateSizeAuto || n < 0 {
		n = runtime.NumGoroutine()
	}
	if n < 1 {
		n = 1
	}
	extraFactor := 1.1 
	n = int(math.Ceil(float64(n) * extraFactor))
	s.buf = make([]runtime.StackRecord, n)
	s.filled = 0
}

func (s *StackTraceBuffer) capacity() int {
	return len(s.buf)
}

func (s *StackTraceBuffer) Size() int {
	return s.filled
}

func (s *StackTraceBuffer) GetPC(goroutineIdx, depthIdx int) uintptr {
	if goroutineIdx >= s.Size() || depthIdx >= len(s.buf[goroutineIdx].Stack0) {
		return 0
	}
	return s.buf[goroutineIdx].Stack0[depthIdx]
}

func (s *StackTraceBuffer) GetDepth(goroutineIdx int) (depth int, allFrames bool) {
	if goroutineIdx >= s.Size() {
		return 0, false
	}
	buf := &s.buf[goroutineIdx]
	i := 0
	for ; i < len(buf.Stack0); i++ {
		if buf.Stack0[i] == 0 {
			break
		}
	}
	return i, i < len(buf.Stack0)
}

func (s *StackTraceBuffer) GetMaxPossibleDepth() int {
	if s.capacity() == 0 {
		return 0
	}
	return len(s.buf[0].Stack0)
}
