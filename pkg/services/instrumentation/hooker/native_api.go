//go:build !windows && (amd64 || arm64) && cgo
// +build !windows
// +build amd64 arm64
// +build cgo

package hooker

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
)



/* #cgo CFLAGS: -I${SRCDIR}
// Dynamic alpine X86
#cgo rookout_dynamic,linux,alpine314,amd64 rookout_dynamic,linux,alpine,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_alpine314_x86_64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static alpine X86
#cgo !rookout_dynamic,linux,alpine314,amd64 !rookout_dynamic,linux,alpine,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_alpine314_x86_64.a -static -lpthread -lrt -ldl -lz -lm -lstdc++
// Dynamic debian X86
#cgo rookout_dynamic,linux,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_ubuntu18_x86_64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static debian X86
#cgo !rookout_dynamic,linux,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_ubuntu18_x86_64.a -static -lpthread -lrt -ldl -lz -lm -lstdc++
// Dynamic macos X86
#cgo darwin,amd64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_macos_x86_64.a -lpthread -lz -lffi -ledit -lm -lc++
// Dynamic alpine arm64
#cgo rookout_dynamic,linux,alpine314,arm64 rookout_dynamic,linux,alpine,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_alpine314_arm64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static alpine arm64
#cgo !rookout_dynamic,linux,alpine314,arm64 !rookout_dynamic,linux,alpine,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_alpine314_arm64.a -static -lpthread -lrt -ldl -lz -lm -lstdc++
// Dynamic debian arm64
#cgo rookout_dynamic,linux,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_ubuntu_arm64.a -lpthread -lrt -ldl -lz -lm -lstdc++
// Static debian arm64
#cgo !rookout_dynamic,linux,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_ubuntu_arm64.a -static -lpthread -lrt -ldl -lz -lstdc++ -lm
// Dynamic macos arm64
#cgo darwin,arm64 LDFLAGS: -v ${SRCDIR}/libhook_lib_pack_macos_arm64.a -lpthread -lz -lffi -ledit -lm -lc++
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

func (a *nativeAPIImpl) RegisterFunctionBreakpointsState(functionEntry, functionEnd uint64, breakpoints []*augs.BreakpointInstance, bpCallback, prologueCallback, shouldRunPrologue uintptr, functionStackUsage int32) (int, error) {
	
	var breakpointAddrs []uint64
	var bpAddr unsafe.Pointer
	var failedCounters []*uint64
	var failedCountersAddr unsafe.Pointer
	bpCallbackPtr := unsafe.Pointer(bpCallback)
	prologueCallbackPtr := unsafe.Pointer(prologueCallback)
	shouldRunProloguePtr := unsafe.Pointer(shouldRunPrologue)

	if len(breakpoints) == 0 {
		prologueCallbackPtr = nil
		functionStackUsage = -1
		shouldRunProloguePtr = nil
	} else {
		breakpointAddrs = make([]uint64, len(breakpoints))
		for i, bp := range breakpoints {
			breakpointAddrs[i] = bp.Addr
		}
		bpAddr = unsafe.Pointer(&breakpointAddrs[0])

		failedCounters = make([]*uint64, len(breakpoints))
		failedCountersAddr = unsafe.Pointer(&failedCounters[0])
	}

	stateID := int(C.RookoutRegisterFunctionBreakpointsState(
		getUnsafePointer(functionEntry),
		getUnsafePointer(functionEnd),
		C.int(len(breakpoints)),
		bpAddr,
		failedCountersAddr,
		bpCallbackPtr,
		prologueCallbackPtr,
		shouldRunProloguePtr,
		C.uint(functionStackUsage)))

	if stateID < 0 {
		return stateID, fmt.Errorf("couldn't set new function breakpoint state (%v) (%s)", breakpoints, C.GoString(C.RookoutGetHookerLastError()))
	}

	for i := range failedCounters {
		breakpoints[i].FailedCounter = failedCounters[i]
	}

	return stateID, nil
}

func (a *nativeAPIImpl) GetInstructionMapping(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	rawAddressMapping := C.RookoutGetInstructionMapping(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd), C.int(stateId))
	var err error = nil
	if rawAddressMapping == nil {
		err = fmt.Errorf("Couldn't get instruction mapping (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	return uintptr(rawAddressMapping), err
}

func (a *nativeAPIImpl) GetUnpatchedInstructionMapping(functionEntry uint64, functionEnd uint64) (uintptr, error) {
	rawUnpatchedAddressMapping := C.RookoutGetUnpatchedInstructionMapping(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd))
	var err error = nil
	if rawUnpatchedAddressMapping == nil {
		err = fmt.Errorf("Couldn't get unpatched instruction mapping (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	return uintptr(rawUnpatchedAddressMapping), err
}

func (a *nativeAPIImpl) ApplyBreakpointsState(functionEntry uint64, functionEnd uint64, stateId int) error {
	ret := int(C.RookoutApplyBreakpointsState(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd), C.int(stateId)))
	if ret != 0 {
		return fmt.Errorf("Couldn't apply breakpoint state (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	return nil
}

func (a *nativeAPIImpl) GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookAddr := uint64(C.RookoutGetHookAddress(funcEntry, funcEnd, C.int(stateId)))
	if hookAddr == uint64(0) {
		err = fmt.Errorf("Failed to get the hook Address (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return uintptr(hookAddr), err
}

func (a *nativeAPIImpl) GetHookSizeBytes(functionEntry uint64, functionEnd uint64, stateId int) (int, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookSize := int(C.RookoutGetHookSizeBytes(funcEntry, funcEnd, C.int(stateId)))
	if hookSize < 0 {
		err = fmt.Errorf("Failed to get the hook size (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return hookSize, err
}

func (a *nativeAPIImpl) GetHookBytes(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookBytes := unsafe.Pointer(C.RookoutGetHookBytesView(funcEntry, funcEnd, C.int(stateId)))
	if hookBytes == nil {
		err = fmt.Errorf("Failed to get the hook bytes (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return uintptr(hookBytes), err
}

func (a *nativeAPIImpl) GetFunctionType(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error) {
	var err error
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	funcType := int(C.RookoutGetFunctionType(funcEntry, funcEnd))
	if funcType < 0 {
		err = fmt.Errorf("Failed to get the function type (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return safe_hook_validator.FunctionType(funcType), err
}

func (a *nativeAPIImpl) GetDangerZoneStartAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	dangerZoneStart := uint64(C.RookoutGetDangerZoneStartAddress(funcEntry, funcEnd))
	if dangerZoneStart == uint64(0) {
		err = fmt.Errorf("Failed to get the function danger zone start Address (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return dangerZoneStart, err
}

func (a *nativeAPIImpl) GetDangerZoneEndAddress(functionEntry uint64, functionEnd uint64) (uint64, error) {
	var err error = nil
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	dangerZoneStart := uint64(C.RookoutGetDangerZoneEndAddress(funcEntry, funcEnd))
	if dangerZoneStart == uint64(0) {
		err = fmt.Errorf("Failed to get the function danger zone end Address (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return dangerZoneStart, err
}

func (a *nativeAPIImpl) TriggerWatchDog(timeoutMS uint64) error {
	var err error = nil
	res := int(C.RookoutTriggerWatchDog(C.ulonglong(timeoutMS)))
	if res < 0 {
		err = fmt.Errorf("Failed to trigger the watchdog (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}
	return err
}

func (a *nativeAPIImpl) DefuseWatchDog() {
	C.RookoutDefuseWatchDog()
}

func Init(someFunc func()) rookoutErrors.RookoutError {
	if C.RookoutInit(unsafe.Pointer(reflect.ValueOf(someFunc).Pointer())) != 0 {
		return rookoutErrors.NewFailedToInitNative(C.GoString(C.RookoutGetHookerLastError()))
	}

	return nil
}

func Destroy() error {
	if C.RookoutDestroy() != 0 {
		return fmt.Errorf("Native `Destroy` failed with error message: %s\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	return nil
}
