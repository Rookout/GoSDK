package hooker

import (
	"time"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_installer"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
)

const (
	InstallHookNumAttempts       = 10    
	InstallHookAttemptDelayMS    = 10    
	InstallHookWatchDogTimeoutMS = 60000 
)

type BreakpointFlowRunner interface {
	GetAddressMapping() (unsafe.Pointer, error)
	GetUnpatchedAddressMapping() (unsafe.Pointer, error)
	ApplyBreakpointsState() error
	IsPatched() bool
	ID() int
	DefaultID() int
}

type BreakpointFlowRunnerInitializationInfo struct {
	Function                  *augs.Function
	BPCallback                uintptr
	PrologueCallback          uintptr
	ShouldRunPrologueCallback uintptr
}

type breakpointFlowRunner struct {
	installerCreator safe_hook_installer.SafeHookInstallerCreator
	nativeAPI        NativeHookerAPI
	info             BreakpointFlowRunnerInitializationInfo
	stateID          int
	functionEntry    uint64
	functionEnd      uint64
}

func (c *breakpointFlowRunner) GetAddressMapping() (unsafe.Pointer, error) {
	mapping, err := c.nativeAPI.GetInstructionMapping(c.functionEntry, c.functionEnd, c.stateID)
	return unsafe.Pointer(mapping), err
}

func (c *breakpointFlowRunner) GetUnpatchedAddressMapping() (unsafe.Pointer, error) {
	mapping, err := c.nativeAPI.GetUnpatchedInstructionMapping(c.functionEntry, c.functionEnd)
	return unsafe.Pointer(mapping), err
}

func (c *breakpointFlowRunner) ApplyBreakpointsState() error {
	var err error = nil
	var shi safe_hook_installer.SafeHookInstaller
	var numGoroutines int
	for attempt := 1; attempt <= InstallHookNumAttempts; attempt++ {
		if attempt > 1 {
			
			time.Sleep(InstallHookAttemptDelayMS * time.Millisecond)
		}
		shi, err = c.installerCreator(c.functionEntry, c.functionEnd, c.stateID, c.nativeAPI,
			suspender.GetSuspender(), callstack.NewStackTraceBuffer(), &safe_hook_validator.ValidatorFactoryImpl{})
		if err != nil {
			
			break
		}
		err = c.nativeAPI.TriggerWatchDog(InstallHookWatchDogTimeoutMS)
		if err != nil {
			logger.Logger().Warningf("Failed to trigger the watchdog on attempt #%d. Reason: %v", attempt, err)
			continue
		}
		numGoroutines, err = shi.InstallHook()
		c.nativeAPI.DefuseWatchDog()
		if err != nil {
			logger.Logger().Warningf("Failed to install the hook on attempt #%d. Detected %d non-system goroutines. Reason: %v", attempt, numGoroutines, err)
			continue
		}
		
		break
	}
	if err == nil {
		
		notifyErr := c.nativeAPI.ApplyBreakpointsState(c.functionEntry, c.functionEnd, c.stateID)
		if notifyErr != nil {
			logger.Logger().Warningf("Failed to notify the native on installing the breakpoint (%v). However, the breakpoint was installed successfully.", notifyErr)
			
		}
	}
	return err
}

func (c *breakpointFlowRunner) IsPatched() bool {
	return c.stateID > 0
}

func (c *breakpointFlowRunner) ID() int {
	return c.stateID
}

func (c *breakpointFlowRunner) DefaultID() int {
	return 0
}

func NewFlowRunner(nativeAPI NativeHookerAPI, initInfo BreakpointFlowRunnerInitializationInfo, requiredBreakpoints []*augs.BreakpointInstance, installerFactory safe_hook_installer.SafeHookInstallerCreator) (*breakpointFlowRunner, error) {
	stateID, err := nativeAPI.RegisterFunctionBreakpointsState(initInfo.Function.Entry, initInfo.Function.End, requiredBreakpoints, initInfo.BPCallback, initInfo.PrologueCallback, initInfo.ShouldRunPrologueCallback, initInfo.Function.StackFrameSize)
	if err != nil {
		return nil, err
	}
	return &breakpointFlowRunner{
		installerCreator: installerFactory,
		info:             initInfo,
		stateID:          stateID,
		functionEntry:    initInfo.Function.Entry,
		functionEnd:      initInfo.Function.End,
		nativeAPI:        nativeAPI,
	}, nil
}
