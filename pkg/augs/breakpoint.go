package augs

import (
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
)



type Breakpoint struct {
	
	Name string `json:"name"`
	
	File string `json:"file"`
	
	Line int `json:"line"`
	
	
	FunctionName string `json:"functionName,omitempty"`

	
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

func NewBreakpointInstance(addr uint64, breakpoint *Breakpoint) *BreakpointInstance {
	b := &BreakpointInstance{
		Addr:       addr,
		Breakpoint: breakpoint,
	}
	return b
}

type Function struct {
	Entry                  uint64
	End                    uint64
	StackFrameSize         int32
	GetBreakpointInstances func() []*BreakpointInstance
}

func NewFunction(entry uint64, end uint64, stackFrameSize int32) *Function {
	return &Function{
		Entry:          entry,
		End:            end,
		StackFrameSize: stackFrameSize,
		GetBreakpointInstances: func() []*BreakpointInstance {
			return []*BreakpointInstance{}
		},
	}
}
