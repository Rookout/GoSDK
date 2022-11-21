package safe_hook_installer

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
	"github.com/Rookout/GoSDK/pkg/types"
	"syscall"
	"unsafe"
)

type SafeHookInstallerCreator func(functionEntry, functionEnd uint64, stateId int,
	nativeApi types.NativeHookerAPI, goroutineSuspender suspender.Suspender,
	stb callstack.IStackTraceBuffer, validatorFactory safe_hook_validator.ValidatorFactory) (SafeHookInstaller, error)

type SafeHookInstaller interface {
	InstallHook() (int, error)
}

type RealSafeHookInstaller struct {
	goroutineSuspender   suspender.Suspender
	stb                  callstack.IStackTraceBuffer
	validator            safe_hook_validator.Validator
	hookSizeBytes        int
	hookDstAddr          uintptr
	hookSrcAddr          uintptr
	hookPageStartAddress uintptr
	hookTotalPagesBytes  int
}

func NewSafeHookInstaller(functionEntry, functionEnd uint64, stateId int,
	nativeApi types.NativeHookerAPI, goroutineSuspender suspender.Suspender,
	stb callstack.IStackTraceBuffer, validatorFactory safe_hook_validator.ValidatorFactory) (SafeHookInstaller, error) {
	shi := &RealSafeHookInstaller{goroutineSuspender: goroutineSuspender, stb: stb}
	err := shi.initHookInfo(nativeApi, functionEntry, functionEnd, stateId)
	if err != nil {
		return nil, err
	}

	err = shi.initValidator(nativeApi, functionEntry, functionEnd, validatorFactory)
	if err != nil {
		return nil, err
	}
	return shi, nil
}

func (s *RealSafeHookInstaller) InstallHook() (int, error) {
	
	n := 0
	var ok bool
	var err error = nil
	var validationRes safe_hook_validator.ValidationErrorFlags
	var modifyPermissionsRes int
	s.goroutineSuspender.StopAll()
	n, ok = s.stb.FillStackTraces()
	if !ok {
		s.goroutineSuspender.ResumeAll()
		err = fmt.Errorf("Wasn't able to collect all goroutines info. Needed to collect %d goroutines", n)
		return n, err
	}
	
	validationRes = s.validator.Validate(s.stb)
	if validationRes != safe_hook_validator.NoError {
		s.goroutineSuspender.ResumeAll()
		err = fmt.Errorf("Detected it's unsafe to install hook at this time. Reason: %s", validationRes.String())
		return n, err
	}
	
	modifyPermissionsRes = setWritable(s.hookPageStartAddress, s.hookTotalPagesBytes)
	if modifyPermissionsRes != 0 {
		s.goroutineSuspender.ResumeAll()
		err = fmt.Errorf("Failed to set write permissions before setting the hook. Error code: %d", modifyPermissionsRes)
		return n, err
	}
	for i := 0; i < s.hookSizeBytes; i++ {
		*(*uint8)(unsafe.Pointer(s.hookDstAddr + uintptr(i))) = *(*uint8)(unsafe.Pointer(s.hookSrcAddr + uintptr(i)))
	}
	modifyPermissionsRes = setExecutable(s.hookPageStartAddress, s.hookTotalPagesBytes)
	if modifyPermissionsRes != 0 {
		
		s.goroutineSuspender.ResumeAll()
		err = fmt.Errorf("Failed to set execute permissions after setting the hook. Error code: %d", modifyPermissionsRes)
		return n, err
	}
	s.goroutineSuspender.ResumeAll()
	return n, nil
}

func (s *RealSafeHookInstaller) initHookInfo(nativeApi types.NativeHookerAPI, functionEntry, functionEnd uint64, stateId int) error {
	var err error = nil
	s.hookSizeBytes, err = nativeApi.GetHookSizeBytes(functionEntry, functionEnd, stateId)
	if err != nil {
		return err
	}

	s.hookDstAddr, err = nativeApi.GetHookAddress(functionEntry, functionEnd, stateId)
	if err != nil {
		return err
	}

	s.hookSrcAddr, err = nativeApi.GetHookBytes(functionEntry, functionEnd, stateId)
	if err != nil {
		return err
	}

	pageSize := uintptr(syscall.Getpagesize())
	pageMask := ^(pageSize - 1)                                                              
	s.hookPageStartAddress = s.hookDstAddr & pageMask                                        
	hookEndPageStartAddress := (s.hookDstAddr + uintptr(s.hookSizeBytes)) & pageMask         
	s.hookTotalPagesBytes = int(pageSize + hookEndPageStartAddress - s.hookPageStartAddress) 

	return nil
}

func (s *RealSafeHookInstaller) initValidator(nativeApi types.NativeHookerAPI, functionEntry, functionEnd uint64,
	validatorFactory safe_hook_validator.ValidatorFactory) error {
	var funcType types.FunctionType
	var err error = nil
	funcType, err = nativeApi.GetFunctionType(functionEntry, functionEnd)
	if err != nil {
		return err
	}

	var dangerZoneStart uint64
	dangerZoneStart, err = nativeApi.GetDangerZoneStartAddress(functionEntry, functionEnd)
	if err != nil {
		return err
	}

	var dangerZoneEnd uint64
	dangerZoneEnd, err = nativeApi.GetDangerZoneEndAddress(functionEntry, functionEnd)
	if err != nil {
		return err
	}

	functionRange := safe_hook_validator.AddressRange{Start: uintptr(functionEntry), End: uintptr(functionEnd)}
	dangerRange := safe_hook_validator.AddressRange{Start: uintptr(dangerZoneStart), End: uintptr(dangerZoneEnd)}
	s.validator, err = validatorFactory.GetValidator(funcType, functionRange, dangerRange)
	if err != nil {
		return err
	}
	return nil
}
