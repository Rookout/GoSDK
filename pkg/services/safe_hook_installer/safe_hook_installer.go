package safe_hook_installer

import (
	"syscall"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
)

type HookManager interface {
	GetFunctionType(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error)
	GetDangerZoneStartAddress() uint64
	GetDangerZoneEndAddress() uint64
}

type SafeHookInstallerCreator func(functionEntry, functionEnd uint64, stateID int, hookWriter *HookWriter,
	nativeApi HookManager, goroutineSuspender suspender.Suspender,
	stb callstack.IStackTraceBuffer, validatorFactory safe_hook_validator.ValidatorFactory) (SafeHookInstaller, error)

type SafeHookInstaller interface {
	InstallHook() (int, rookoutErrors.RookoutError)
}

type RealSafeHookInstaller struct {
	goroutineSuspender suspender.Suspender
	stb                callstack.IStackTraceBuffer
	validator          safe_hook_validator.Validator
	hookWriter         *HookWriter
}

type HookWriter struct {
	Hook                     []byte
	HookAddr                 uintptr
	originalMemoryProtection int
	hookPageAlignedStart     uintptr
	hookPageAlignedEnd       uintptr
}

func NewHookWriter(hookAddr uintptr, hook []byte) *HookWriter {
	
	hookPageAlignedStart := hookAddr & (^uintptr(syscall.Getpagesize() - 1))
	hookPageAlignedEnd := ((hookAddr + uintptr(len(hook)) - 1) & (^uintptr(syscall.Getpagesize() - 1))) + uintptr(syscall.Getpagesize())

	return &HookWriter{
		hookPageAlignedStart: hookPageAlignedStart,
		hookPageAlignedEnd:   hookPageAlignedEnd,
		Hook:                 hook,
		HookAddr:             hookAddr,
	}
}

func NewSafeHookInstaller(functionEntry, functionEnd uint64, stateID int,
	hookWriter *HookWriter, hookManager HookManager, goroutineSuspender suspender.Suspender,
	stb callstack.IStackTraceBuffer, validatorFactory safe_hook_validator.ValidatorFactory) (SafeHookInstaller, error) {
	shi := &RealSafeHookInstaller{goroutineSuspender: goroutineSuspender, stb: stb, hookWriter: hookWriter}
	err := shi.initValidator(hookManager, functionEntry, functionEnd, validatorFactory)
	if err != nil {
		return nil, err
	}
	return shi, nil
}

func (s *RealSafeHookInstaller) InstallHook() (n int, err rookoutErrors.RookoutError) {
	
	var ok bool
	var validationRes safe_hook_validator.ValidationErrorFlags

	s.goroutineSuspender.StopAll()
	n, ok = s.stb.FillStackTraces()
	if !ok {
		s.goroutineSuspender.ResumeAll()
		return n, rookoutErrors.NewFailedToCollectGoroutinesInfo(n)
	}
	
	validationRes = s.validator.Validate(s.stb)
	if validationRes != safe_hook_validator.NoError {
		s.goroutineSuspender.ResumeAll()
		return n, rookoutErrors.NewUnsafeToInstallHook(validationRes.String())
	}
	
	errno := s.hookWriter.write()
	if errno != 0 {
		s.goroutineSuspender.ResumeAll()
		return n, rookoutErrors.NewFailedToWriteBytes(errno)
	}
	s.goroutineSuspender.ResumeAll()
	return n, nil
}

func (s *RealSafeHookInstaller) initValidator(hookManager HookManager, functionEntry, functionEnd uint64,
	validatorFactory safe_hook_validator.ValidatorFactory) error {
	funcType, err := hookManager.GetFunctionType(functionEntry, functionEnd)
	if err != nil {
		return err
	}

	dangerZoneStart := hookManager.GetDangerZoneStartAddress()
	dangerZoneEnd := hookManager.GetDangerZoneEndAddress()

	functionRange := safe_hook_validator.AddressRange{Start: uintptr(functionEntry), End: uintptr(functionEnd)}
	dangerRange := safe_hook_validator.AddressRange{Start: uintptr(dangerZoneStart), End: uintptr(dangerZoneEnd)}
	s.validator, err = validatorFactory.GetValidator(funcType, functionRange, dangerRange)
	if err != nil {
		return err
	}
	return nil
}
