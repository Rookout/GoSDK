package safe_hook_validator

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/types"
	"strings"
)

type ValidationErrorFlags int

const (
	NoError                    ValidationErrorFlags = 0
	IllegalPcValue             ValidationErrorFlags = 1 << (iota - 1) 
	PcInDangerZoneEntry                                               
	PcInDangerZoneAfterEntry                                          
	PcInFunction                                                      
	DeepStackDidntResolveAllPc                                        
)

func (f ValidationErrorFlags) String() string {
	if f == NoError {
		return "No error"
	}
	msgs := make([]string, 0)
	if f&IllegalPcValue != 0 {
		f ^= IllegalPcValue
		msgs = append(msgs, "Had PC in function entry which is illegal!")
	}
	if f&PcInDangerZoneEntry != 0 {
		f ^= PcInDangerZoneEntry
		msgs = append(msgs, "Had PC at function entry+1. Not sure if dangerous or not!")
	}
	if f&PcInDangerZoneAfterEntry != 0 {
		f ^= PcInDangerZoneAfterEntry
		msgs = append(msgs, "Had PC in the danger zone!")
	}
	if f&PcInFunction != 0 {
		f ^= PcInFunction
		msgs = append(msgs, "Had PC in the function!")
	}
	if f&DeepStackDidntResolveAllPc != 0 {
		f ^= DeepStackDidntResolveAllPc
		msgs = append(msgs, "Failed to retrieve an entire stack trace, possible PC in function!")
	}
	if f != 0 {
		msgs = append(msgs, "Illegal validation result!")
	}
	return strings.Join(msgs, " | ")
}

type AddressRange struct {
	Start uintptr 
	End   uintptr 
}



type ValidatorFactory interface {
	GetValidator(funcType types.FunctionType, functionRange, dangerRange AddressRange) (Validator, error)
}

type Validator interface {
	Validate(buffer callstack.IStackTraceBuffer) ValidationErrorFlags
}

type ValidatorFactoryImpl struct {
}

func (v *ValidatorFactoryImpl) GetValidator(funcType types.FunctionType, functionRange, dangerRange AddressRange) (Validator, error) {
	switch funcType {
	case types.FunctionType0:
		if dangerRange.Start != functionRange.Start+1 {
			return nil, fmt.Errorf("Danger zone should start after the first byte of the function!")
		}
		return &type0Validator{
			functionRange: functionRange,
			dangerRange:   dangerRange,
		}, nil
	case types.FunctionType1:
		return &type1Validator{
			functionRange: functionRange,
		}, nil
	case types.FunctionType2:
		return &type2Validator{}, nil
	}
	return nil, fmt.Errorf("Illegal function type! Got %d=%s", int(funcType), funcType.String())
}

type type0Validator struct {
	functionRange AddressRange
	dangerRange   AddressRange
}

type type1Validator struct {
	functionRange AddressRange
}

type type2Validator struct {
}

func (r *AddressRange) contains(pc uintptr) bool {
	return (r.Start <= pc) && (pc < r.End)
}

func (v *type0Validator) Validate(buffer callstack.IStackTraceBuffer) ValidationErrorFlags {
	ret := NoError
	totalGoroutines := buffer.Size()
	for gr := 0; gr < totalGoroutines; gr++ {
		depth, _ := buffer.GetDepth(gr)
		for d := 0; d < depth; d++ {
			pc := buffer.GetPC(gr, d)
			if pc == v.functionRange.Start {
				ret |= IllegalPcValue
			} else if pc == v.dangerRange.Start {
				ret |= PcInDangerZoneEntry
			} else if v.dangerRange.contains(pc) {
				ret |= PcInDangerZoneAfterEntry
			}
		}
	}
	return ret
}

func (v *type1Validator) Validate(buffer callstack.IStackTraceBuffer) ValidationErrorFlags {
	ret := NoError
	totalGoroutines := buffer.Size()
	for gr := 0; gr < totalGoroutines; gr++ {
		depth, allFrames := buffer.GetDepth(gr)
		if !allFrames {
			ret |= DeepStackDidntResolveAllPc
		}
		for d := 0; d < depth; d++ {
			pc := buffer.GetPC(gr, d)
			if pc == v.functionRange.Start {
				ret |= IllegalPcValue
			} else if v.functionRange.contains(pc) {
				ret |= PcInFunction
			}
		}
	}
	return ret
}

func (v *type2Validator) Validate(buffer callstack.IStackTraceBuffer) ValidationErrorFlags {
	
	return NoError
}
