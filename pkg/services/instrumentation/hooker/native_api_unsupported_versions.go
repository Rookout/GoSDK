//go:build windows || (!amd64 && !arm64) || !cgo
// +build windows !amd64,!arm64 !cgo

package hooker

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
)

type NativeAPI struct{}

func NewNativeAPI() *NativeAPI {
	return &NativeAPI{}
}

func Init(_ func()) rookoutErrors.RookoutError {
	return rookoutErrors.NewUnsupportedPlatform()
}

func Destroy() error {
	return rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) RegisterFunctionBreakpointsState(functionEntry Address, functionEnd Address, breakpoints []*augs.BreakpointInstance, bpCallback uintptr, prologue []byte, functionStackUsage int32) (stateId int, err error) {
	return 0, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) GetInstructionMapping(_ uint64, _ uint64, _ int) ([]module.AddressMapping, []module.AddressMapping, error) {
	return nil, nil, rookoutErrors.NewUnsupportedPlatform()
}
func (n *NativeAPI) GetStateEntryAddr(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	return 0, rookoutErrors.NewUnsupportedPlatform()
}
func (n *NativeAPI) GetUnpatchedInstructionMapping(functionEntry uint64, functionEnd uint64) (addressMappings []module.AddressMapping, offsetMappings []module.AddressMapping, err error) {
	return nil, nil, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) GetStackUsageMap() (map[uint64][]map[string]int64, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) ApplyBreakpointsState(functionEntry Address, functionEnd Address, stateId int) (err error) {
	return rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, rookoutErrors.RookoutError) {
	return 0, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) GetFunctionType(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error) {
	return 0, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) GetDangerZoneStartAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	return 0, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) GetDangerZoneEndAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	return 0, rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) TriggerWatchDog(timeoutMS uint64) error {
	return rookoutErrors.NewUnsupportedPlatform()
}

func (n *NativeAPI) DefuseWatchDog() {
}
