package safe_hook_installer

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
	"github.com/Rookout/GoSDK/pkg/types"
	"unsafe"
)

type SafeHookInstallerCreator func(functionEntry, functionEnd uint64, stateId int,
	nativeApi types.NativeHookerAPI, goroutineSuspender suspender.Suspender,
	stb callstack.IStackTraceBuffer, validatorFactory safe_hook_validator.ValidatorFactory) (SafeHookInstaller, error)

type SafeHookInstaller interface {
	InstallHook() (int, error)
}

type RealSafeHookInstaller struct {
	goroutineSuspender suspender.Suspender
	stb                callstack.IStackTraceBuffer
	validator          safe_hook_validator.Validator
	hookSizeBytes      int
	hookDstAddr        uintptr
	hookSrcAddr        uintptr
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
		err = fmt.Errorf("Detected it's unsafe to insatll hook at this time. Reason: %s", validationRes.String())
		return n, err
	}
	
	for i := 0; i < s.hookSizeBytes; i++ {
		*(*uint8)(unsafe.Pointer(s.hookDstAddr + uintptr(i))) = *(*uint8)(unsafe.Pointer(s.hookSrcAddr + uintptr(i)))
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
