package augs

import (
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
)



type Breakpoint struct {
	
	Name string `json:"name"`
	
	File string `json:"file"`
	
	Line int `json:"line"`

	
	Stacktrace int `json:"stacktrace"`
	Instances  []*BreakpointInstance
}

type BreakpointInstance struct {
	Addr             uint64
	VariableLocators []*variable.VariableLocator
	Breakpoint       *Breakpoint
	Function         *Function
	FailedCounter    *uint64
}

func NewBreakpointInstance(addr uint64, breakpoint *Breakpoint, function *Function) *BreakpointInstance {
	b := &BreakpointInstance{
		Addr:       addr,
		Breakpoint: breakpoint,
		Function:   function,
	}
	return b
}

type Function struct {
	Entry                   uint64
	End                     uint64
	StackFrameSize          int32
	GetBreakpointInstances  func() []*BreakpointInstance
	MiddleTrampolineAddress unsafe.Pointer
	FinalTrampolinePointer  *uint64
	PatchedBytes            []byte
	Hooked                  bool
	Prologue                []byte
	FunctionCopyStateID     int
}

func NewFunction(entry uint64, end uint64, stackFrameSize int32, middleTrampolineAddress unsafe.Pointer, finalTrampolinePointer *uint64) *Function {
	return &Function{
		Entry:          entry,
		End:            end,
		StackFrameSize: stackFrameSize,
		GetBreakpointInstances: func() []*BreakpointInstance {
			return []*BreakpointInstance{}
		},
		MiddleTrampolineAddress: middleTrampolineAddress,
		FinalTrampolinePointer:  finalTrampolinePointer,
		Hooked:                  false,
		FunctionCopyStateID:     -1, 
	}
}
