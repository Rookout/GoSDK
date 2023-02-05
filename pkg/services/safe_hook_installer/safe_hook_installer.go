package safe_hook_installer

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
	"syscall"
)

type HookManager interface {
	GetHookAddress(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error)
	GetHookSizeBytes(functionEntry uint64, functionEnd uint64, stateId int) (int, error)
	GetHookBytes(functionEntry uint64, functionEnd uint64, stateId int) (uintptr, error)
	GetFunctionType(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error)
	GetDangerZoneStartAddress(functionEntry uint64, functionEnd uint64) (uint64, error)
	GetDangerZoneEndAddress(functionEntry uint64, functionEnd uint64) (uint64, error)
}

type SafeHookInstallerCreator func(functionEntry, functionEnd uint64, stateId int,
	nativeApi HookManager, goroutineSuspender suspender.Suspender,
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
	hookManager HookManager, goroutineSuspender suspender.Suspender,
	stb callstack.IStackTraceBuffer, validatorFactory safe_hook_validator.ValidatorFactory) (SafeHookInstaller, error) {
	shi := &RealSafeHookInstaller{goroutineSuspender: goroutineSuspender, stb: stb}
	err := shi.initHookInfo(hookManager, functionEntry, functionEnd, stateId)
	if err != nil {
		return nil, err
	}

	err = shi.initValidator(hookManager, functionEntry, functionEnd, validatorFactory)
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
	var writeBytesRes int
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
	
	writeBytesRes = writeBytes(s.hookDstAddr, s.hookSrcAddr, s.hookSizeBytes)
	if writeBytesRes != 0 {
		
		s.goroutineSuspender.ResumeAll()
		err = fmt.Errorf("Failed to set the hook bytes. Error code: %d", writeBytesRes)
		return n, err
	}
	s.goroutineSuspender.ResumeAll()
	return n, nil
}

func (s *RealSafeHookInstaller) initHookInfo(hookManager HookManager, functionEntry, functionEnd uint64, stateId int) error {
	var err error = nil
	s.hookSizeBytes, err = hookManager.GetHookSizeBytes(functionEntry, functionEnd, stateId)
	if err != nil {
		return err
	}

	s.hookDstAddr, err = hookManager.GetHookAddress(functionEntry, functionEnd, stateId)
	if err != nil {
		return err
	}

	s.hookSrcAddr, err = hookManager.GetHookBytes(functionEntry, functionEnd, stateId)
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

func (s *RealSafeHookInstaller) initValidator(hookManager HookManager, functionEntry, functionEnd uint64,
	validatorFactory safe_hook_validator.ValidatorFactory) error {
	var funcType safe_hook_validator.FunctionType
	var err error = nil
	funcType, err = hookManager.GetFunctionType(functionEntry, functionEnd)
	if err != nil {
		return err
	}

	var dangerZoneStart uint64
	dangerZoneStart, err = hookManager.GetDangerZoneStartAddress(functionEntry, functionEnd)
	if err != nil {
		return err
	}

	var dangerZoneEnd uint64
	dangerZoneEnd, err = hookManager.GetDangerZoneEndAddress(functionEntry, functionEnd)
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
