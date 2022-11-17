package types

type Address = uint64
type FunctionType int

const (
	FunctionType0 FunctionType = iota
	FunctionType1
)

func (f FunctionType) String() string {
	switch f {
	case FunctionType0:
		return "Type0"
	case FunctionType1:
		return "Type1"
	default:
		return "Illegal"
	}
}

type NativeHookerAPI interface {
	RegisterFunctionBreakpointsState(functionEntry Address, functionEnd Address, breakpoints []uint64, bpCallback uintptr, prologueCallback uintptr, shouldRunPrologue uintptr, functionStackUsage int32) (stateId int, err error)
	GetInstructionMapping(functionEntry Address, functionEnd Address, stateId int) (rawAddressMapping uintptr, err error)
	GetUnpatchedInstructionMapping(functionEntry uint64, functionEnd uint64) (uintptr, error)
	GetPrologueStackUsage() int32
	GetPrologueAfterUsingStackOffset() int
	GetBreakpointStackUsage() int32
	GetBreakpointTrampolineSizeInBytes() int
	ApplyBreakpointsState(functionEntry Address, functionEnd Address, stateId int) (err error)
	GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error)
	GetHookSizeBytes(functionEntry uint64, functionEnd uint64, stateId int) (int, error)
	GetHookBytes(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error)
	GetFunctionType(functionEntry uint64, functionEnd uint64) (FunctionType, error)
	GetDangerZoneStartAddress(functionEntry uint64, functionEnd uint64) (uint64, error)
	GetDangerZoneEndAddress(functionEntry uint64, functionEnd uint64) (uint64, error)
	TriggerWatchDog(timeoutMS uint64) error
	DefuseWatchDog()
}
