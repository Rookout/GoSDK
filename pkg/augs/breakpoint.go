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
}
