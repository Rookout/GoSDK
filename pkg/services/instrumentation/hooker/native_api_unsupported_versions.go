//go:build windows || !amd64 || !cgo
// +build windows !amd64 !cgo

package hooker

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"runtime"
)

type NativeAPI struct{}

func NewNativeAPI() *NativeAPI {
	return &NativeAPI{}
}

var unsupportedPlatformErr = func() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("unsupported platform - Windows")
	} else if runtime.GOARCH != "amd64" {
		return fmt.Errorf("unsupported platform architecture - %s", runtime.GOARCH)
	} else {
		return fmt.Errorf("unsupported platform - compiling without CGO enabled")
	}
}()

func (n *NativeAPI) RegisterFunctionBreakpointsState(functionEntry types.Address, functionEnd types.Address, breakpoints []uint64, bpCallback uintptr, prologueCallback uintptr, shouldRunPrologue uintptr, functionStackUsage int32) (stateId int, err error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetInstructionMapping(functionEntry types.Address, functionEnd types.Address, stateId int) (rawAddressMapping uintptr, err error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetUnpatchedInstructionMapping(functionEntry uint64, functionEnd uint64) (uintptr, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetHookSizeBytes(functionEntry uint64, functionEnd uint64, stateId int) (int, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetHookBytes(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetFunctionType(functionEntry uint64, functionEnd uint64) (types.FunctionType, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetDangerZoneStartAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) GetDangerZoneEndAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	return 0, unsupportedPlatformErr
}

func (n *NativeAPI) TriggerWatchDog(timeoutMS uint64) error {
	return unsupportedPlatformErr
}

func (n *NativeAPI) DefuseWatchDog() {
}

func (n *NativeAPI) ApplyBreakpointsState(functionEntry uint64, functionEnd uint64, stateId int) error {
	return unsupportedPlatformErr
}

func (n *NativeAPI) GetPrologueStackUsage() int32 {
	return 0
}

func (n *NativeAPI) GetPrologueAfterUsingStackOffset() int {
	return 0
}

func (n *NativeAPI) GetBreakpointStackUsage() int32 {
	return 0
}

func (n *NativeAPI) GetBreakpointTrampolineSizeInBytes() int {
	return 0
}

func Init(_ func()) rookoutErrors.RookoutError {
	if runtime.GOOS == "windows" {
		return rookoutErrors.NewUnsupportedPlatform("Windows")
	} else if runtime.GOARCH != "amd64" {
		return rookoutErrors.NewUnsupportedPlatform("Non AMD64 or unsupported OS")
	}
	return rookoutErrors.NewCompiledWithoutCGO()
}

func Destroy() error {
	return unsupportedPlatformErr
}
