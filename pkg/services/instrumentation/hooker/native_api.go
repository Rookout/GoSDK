//go:build !windows && (amd64 || (arm64 && !darwin)) && cgo
// +build !windows
// +build amd64 arm64,!darwin
// +build cgo

package hooker

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)



/* #cgo CFLAGS: -I${SRCDIR}
// Dynamic alpine
#cgo rookout_dynamic,linux,alpine314,amd64 rookout_dynamic,linux,alpine,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_alpine314_x86_64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static alpine
#cgo !rookout_dynamic,linux,alpine314,amd64 !rookout_dynamic,linux,alpine,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_alpine314_x86_64.a -static -lpthread -lrt -ldl -lz -lm -lstdc++
// Dynamic debian
#cgo rookout_dynamic,linux,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_ubuntu18_x86_64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static debian
#cgo !rookout_dynamic,linux,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_ubuntu18_x86_64.a -static -lpthread -lrt -ldl -lz -lm -lstdc++
// Dynamic macos
#cgo darwin,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_macos_x86_64.a -lpthread -lz -lffi -ledit -lm -lc++
// Dynamic linux-arm
#cgo rookout_dynamic,linux,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_linux_arm64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static linux-arm
#cgo !rookout_dynamic,linux,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_linux_arm64.a -static -lpthread -lrt -ldl -lz -lstdc++ -lm
#include <stdlib.h>
#include <hook_api.h>
*/
import "C"

type nativeAPIImpl struct{}

func NewNativeAPI() *nativeAPIImpl {
	return &nativeAPIImpl{}
}

func getUnsafePointer(value uint64) unsafe.Pointer {
	//goland:noinspection GoVetUnsafePointer
	return unsafe.Pointer(uintptr(value))
}

func (a *nativeAPIImpl) RegisterFunctionBreakpointsState(functionEntry, functionEnd uint64, breakpoints []uint64, bpCallback, prologueCallback, shouldRunPrologue uintptr, functionStackUsage int32) (int, error) {
	

	var bpAddr unsafe.Pointer
	bpCallbackPtr := unsafe.Pointer(bpCallback)
	prologueCallbackPtr := unsafe.Pointer(prologueCallback)
	shouldRunProloguePtr := unsafe.Pointer(shouldRunPrologue)

	if len(breakpoints) == 0 {
		bpAddr = nil
		prologueCallbackPtr = nil
		functionStackUsage = -1
		shouldRunProloguePtr = nil
	} else {
		bpAddr = unsafe.Pointer(&breakpoints[0])
	}

	stateId := int(C.RegisterFunctionBreakpointsState(
		getUnsafePointer(functionEntry),
		getUnsafePointer(functionEnd),
		C.int(len(breakpoints)),
		bpAddr,
		bpCallbackPtr,
		prologueCallbackPtr,
		shouldRunProloguePtr,
		C.uint(functionStackUsage),
	))

	if stateId < 0 {
		return stateId, fmt.Errorf("Couldn't set new function breakpoint state (%v) (%s)\n", breakpoints, C.GoString(C.GetHookerLastError()))
	}

	return stateId, nil
}

func (a *nativeAPIImpl) GetInstructionMapping(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	rawAddressMapping := C.GetInstructionMapping(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd), C.int(stateId))
	var err error = nil
	if rawAddressMapping == nil {
		err = fmt.Errorf("Couldn't get instruction mapping (%s)\n", C.GoString(C.GetHookerLastError()))
	}

	return uintptr(rawAddressMapping), err
}

func (a *nativeAPIImpl) GetUnpatchedInstructionMapping(functionEntry uint64, functionEnd uint64) (uintptr, error) {
	rawUnpatchedAddressMapping := C.GetUnpatchedInstructionMapping(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd))
	var err error = nil
	if rawUnpatchedAddressMapping == nil {
		err = fmt.Errorf("Couldn't get unpatched instruction mapping (%s)\n", C.GoString(C.GetHookerLastError()))
	}

	return uintptr(rawUnpatchedAddressMapping), err
}

func (a *nativeAPIImpl) GetStackUsageMap() (map[uint64][]map[string]int64, rookoutErrors.RookoutError) {
	const stackUsageBufferSize = 100000
	stackUsageMap := make(map[uint64][]map[string]int64)
	stackUsageBufferPtr := C.malloc(C.ulong(C.sizeof_char * stackUsageBufferSize))
	defer C.free(stackUsageBufferPtr)
	stackUsageBufferLen := C.GetStackUsageJSON((*C.char)(stackUsageBufferPtr), C.ulong(stackUsageBufferSize))
	if stackUsageBufferLen < 0 {
		return nil, rookoutErrors.NewFailedToGetStackUsageMap(C.GoString(C.GetHookerLastError()))
	}
	stackUsageBuffer := C.GoBytes(stackUsageBufferPtr, stackUsageBufferLen)
	err := json.Unmarshal(stackUsageBuffer, &stackUsageMap)
	if err != nil {
		return nil, rookoutErrors.NewFailedToParseStackUsageMap(string(stackUsageBuffer), err)
	}
	return stackUsageMap, nil
}

func (a *nativeAPIImpl) ApplyBreakpointsState(functionEntry uint64, functionEnd uint64, stateId int) error {
	ret := int(C.ApplyBreakpointsState(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd), C.int(stateId)))
	if ret != 0 {
		return fmt.Errorf("Couldn't apply breakpoint state (%s)\n", C.GoString(C.GetHookerLastError()))
	}

	return nil
}

func (a *nativeAPIImpl) GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookAddr := uint64(C.GetHookAddress(funcEntry, funcEnd, C.int(stateId)))
	if hookAddr == uint64(0) {
		err = fmt.Errorf("Failed to get the hook Address (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return uintptr(hookAddr), err
}

func (a *nativeAPIImpl) GetHookSizeBytes(functionEntry uint64, functionEnd uint64, stateId int) (int, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookSize := int(C.GetHookSizeBytes(funcEntry, funcEnd, C.int(stateId)))
	if hookSize < 0 {
		err = fmt.Errorf("Failed to get the hook size (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return hookSize, err
}

func (a *nativeAPIImpl) GetHookBytes(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookBytes := unsafe.Pointer(C.GetHookBytesView(funcEntry, funcEnd, C.int(stateId)))
	if hookBytes == nil {
		err = fmt.Errorf("Failed to get the hook bytes (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return uintptr(hookBytes), err
}

func (a *nativeAPIImpl) GetFunctionType(functionEntry uint64, functionEnd uint64) (types.FunctionType, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	funcType := int(C.GetFunctionType(funcEntry, funcEnd))
	if funcType < 0 {
		err = fmt.Errorf("Failed to get the function type (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return types.FunctionType(funcType), err
}

func (a *nativeAPIImpl) GetDangerZoneStartAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	dangerZoneStart := uint64(C.GetDangerZoneStartAddress(funcEntry, funcEnd))
	if dangerZoneStart == uint64(0) {
		err = fmt.Errorf("Failed to get the function danger zone start Address (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return dangerZoneStart, err
}

func (a *nativeAPIImpl) GetDangerZoneEndAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	dangerZoneStart := uint64(C.GetDangerZoneEndAddress(funcEntry, funcEnd))
	if dangerZoneStart == uint64(0) {
		err = fmt.Errorf("Failed to get the function danger zone end Address (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return dangerZoneStart, err
}

func (a *nativeAPIImpl) TriggerWatchDog(timeoutMS uint64) error {
	var err error = nil
	res := int(C.TriggerWatchDog(C.ulonglong(timeoutMS)))
	if res < 0 {
		err = fmt.Errorf("Failed to trigger the watchdog (%s)\n", C.GoString(C.GetHookerLastError()))
	}
	return err
}

func (a *nativeAPIImpl) DefuseWatchDog() {
	C.DefuseWatchDog()
}

func Init(someFunc func()) rookoutErrors.RookoutError {
	if C.Init(unsafe.Pointer(reflect.ValueOf(someFunc).Pointer())) != 0 {
		return rookoutErrors.NewFailedToInitNative(C.GoString(C.GetHookerLastError()))
	}

	return nil
}

func Destroy() error {
	if C.Destroy() != 0 {
		return fmt.Errorf("Native `Destroy` failed with error message: %s\n", C.GoString(C.GetHookerLastError()))
	}

	return nil
}
