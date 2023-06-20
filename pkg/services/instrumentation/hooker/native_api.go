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
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
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



func cArrayToUint64Slice(arrayPointer unsafe.Pointer, count int) []uint64 {
	return (*[1 << 28]uint64)(arrayPointer)[:count:count]
}




func bufferToAddressMapping(addressMappingsBufferPointer unsafe.Pointer, origFuncAddress uintptr) (addressMappings []module.AddressMapping,
	offsetMappings []module.AddressMapping) {
	addressMappingsCount := *(*uint64)(addressMappingsBufferPointer)
	
	addressMappingsBufferPointer = unsafe.Pointer(uintptr(addressMappingsBufferPointer) + unsafe.Sizeof(uint64(0)))
	addressMappingsSlice := cArrayToUint64Slice(addressMappingsBufferPointer, int(addressMappingsCount)*2)

	newFuncAddress := uintptr(addressMappingsSlice[0])

	for i := 0; i < len(addressMappingsSlice); i += 2 {
		newAddress := uintptr(addressMappingsSlice[i])
		origAddress := uintptr(addressMappingsSlice[i+1])
		newOffset := newAddress - newFuncAddress
		origOffset := origAddress - origFuncAddress

		
		
		if _, ok := module.CallbacksMarkers[origAddress]; ok {
			origOffset = origAddress
		}

		addressMappings = append(addressMappings, module.AddressMapping{newAddress, origAddress})
		offsetMappings = append(offsetMappings, module.AddressMapping{newOffset, origOffset})
	}
	return addressMappings, offsetMappings
}

func getUnsafePointer(value uint64) unsafe.Pointer {
	//goland:noinspection GoVetUnsafePointer
	return unsafe.Pointer(uintptr(value))
}

func (a *nativeAPIImpl) RegisterFunctionBreakpointsState(functionEntry, functionEnd uint64, breakpoints []*augs.BreakpointInstance, bpCallback uintptr, prologue []byte, functionStackUsage int32) (int, error) {
	
	var breakpointAddrs []uint64
	var bpAddr unsafe.Pointer
	var failedCounters []*uint64
	var failedCountersAddr unsafe.Pointer
	bpCallbackPtr := unsafe.Pointer(bpCallback)
	var prologueAddr unsafe.Pointer
	prologueLen := 0

	if len(breakpoints) > 0 {
		breakpointAddrs = make([]uint64, len(breakpoints))
		for i, bp := range breakpoints {
			breakpointAddrs[i] = bp.Addr
		}
		bpAddr = unsafe.Pointer(&breakpointAddrs[0])

		failedCounters = make([]*uint64, len(breakpoints))
		failedCountersAddr = unsafe.Pointer(&failedCounters[0])

		if len(prologue) > 0 {
			prologueAddr = unsafe.Pointer(&prologue[0])
			prologueLen = len(prologue)
		}
	}

	stateID := int(C.RookoutRegisterFunctionBreakpointsState(
		getUnsafePointer(functionEntry),
		getUnsafePointer(functionEnd),
		C.int(len(breakpoints)),
		bpAddr,
		failedCountersAddr,
		bpCallbackPtr,
		prologueAddr,
		C.int(prologueLen),
		C.uint(functionStackUsage)))

	if stateID < 0 {
		return stateID, fmt.Errorf("couldn't set new function breakpoint state (%v) (%s)", breakpoints, C.GoString(C.RookoutGetHookerLastError()))
	}

	for i := range failedCounters {
		breakpoints[i].FailedCounter = failedCounters[i]
	}

	return stateID, nil
}

func (a *nativeAPIImpl) GetInstructionMapping(functionEntry uint64, functionEnd uint64, stateId int) (addressMappings []module.AddressMapping, offsetMappings []module.AddressMapping, err error) {
	rawAddressMapping := C.RookoutGetInstructionMapping(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd), C.int(stateId))
	if rawAddressMapping == nil {
		return nil, nil, fmt.Errorf("Couldn't get instruction mapping (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	addressMappings, offsetMappings = bufferToAddressMapping(rawAddressMapping, uintptr(functionEntry))
	return addressMappings, offsetMappings, nil
}


func (a *nativeAPIImpl) GetStateEntryAddr(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error) {
	addressMappings, _, err := a.GetInstructionMapping(functionEntry, functionEnd, stateId)
	if err != nil {
		return 0, err
	}
	return addressMappings[0].NewAddress, nil
}

func (a *nativeAPIImpl) GetUnpatchedInstructionMapping(functionEntry uint64, functionEnd uint64) (addressMappings []module.AddressMapping, offsetMappings []module.AddressMapping, err error) {
	rawUnpatchedAddressMapping := C.RookoutGetUnpatchedInstructionMapping(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd))
	if rawUnpatchedAddressMapping == nil {
		return nil, nil, fmt.Errorf("Couldn't get unpatched instruction mapping (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	addressMappings, offsetMappings = bufferToAddressMapping(rawUnpatchedAddressMapping, uintptr(functionEntry))
	return addressMappings, offsetMappings, nil
}

func (a *nativeAPIImpl) ApplyBreakpointsState(functionEntry uint64, functionEnd uint64, stateId int) error {
	ret := int(C.RookoutApplyBreakpointsState(getUnsafePointer(functionEntry), getUnsafePointer(functionEnd), C.int(stateId)))
	if ret != 0 {
		return fmt.Errorf("Couldn't apply breakpoint state (%s)\n", C.GoString(C.RookoutGetHookerLastError()))
	}

	return nil
}

func (a *nativeAPIImpl) GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, rookoutErrors.RookoutError) {
	funcEntry := getUnsafePointer(functionEntry)
	funcEnd := getUnsafePointer(functionEnd)
	hookAddr := uint64(C.RookoutGetHookAddress(funcEntry, funcEnd, C.int(stateId)))
	if hookAddr == 0 {
		return 0, rookoutErrors.NewFailedToGetHookAddress(C.GoString(C.RookoutGetHookerLastError()))
	}
	return uintptr(hookAddr), nil
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
