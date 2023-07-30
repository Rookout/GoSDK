package hooker

import (
	"time"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
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
	GetAddressMapping() ([]module.AddressMapping, []module.AddressMapping, error)
	ApplyBreakpointsState() error
	IsUnhookedState() bool
	ID() int
	UnhookedFunctionStateID() int
}

type BreakpointFlowRunnerInitializationInfo struct {
	Function   *augs.Function
	BPCallback uintptr
}

type breakpointFlowRunner struct {
	installerCreator safe_hook_installer.SafeHookInstallerCreator
	nativeAPI        NativeHookerAPI
	info             BreakpointFlowRunnerInitializationInfo
	stateID          int
	function         *augs.Function
	jumpDestination  uintptr
}

func (c *breakpointFlowRunner) GetAddressMapping() ([]module.AddressMapping, []module.AddressMapping, error) {
	return c.nativeAPI.GetInstructionMapping(c.function.Entry, c.function.End, c.stateID)
}

func (c *breakpointFlowRunner) installHook() (err error) {
	var shi safe_hook_installer.SafeHookInstaller
	var numGoroutines int
	hookWriter, err := c.getHookWriter()
	if err != nil {
		return err
	}
	hookManager := safe_hook_installer.NewHookManager(hookWriter.HookAddr, hookWriter.Hook, c.nativeAPI.GetFunctionType)

	err = hookWriter.AddWritePermission()
	if err != nil {
		return err
	}
	defer func() {
		restoreErr := hookWriter.RestorePermissions()
		if restoreErr != nil {
			logger.Logger().WithError(restoreErr).Warning("Unable to restore memory section to original permissions")
		}

		
		if err == nil {
			err = restoreErr
		}
	}()

	for attempt := 1; attempt <= InstallHookNumAttempts; attempt++ {
		if attempt > 1 {
			
			time.Sleep(InstallHookAttemptDelayMS * time.Millisecond)
		}
		shi, err = c.installerCreator(c.function.Entry, c.function.End, c.stateID, hookWriter, hookManager,
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
		
		c.function.Hooked = !c.IsUnhookedState()
		break
	}
	if err == nil {
		
		notifyErr := c.nativeAPI.ApplyBreakpointsState(c.function.Entry, c.function.End, c.stateID)
		if notifyErr != nil {
			logger.Logger().Warningf("Failed to notify the native on installing the breakpoint (%v). However, the breakpoint was installed successfully.", notifyErr)
			
		}
	}
	return err
}

func (c *breakpointFlowRunner) getHookWriter() (*safe_hook_installer.HookWriter, rookoutErrors.RookoutError) {
	hookAddr, err := c.nativeAPI.GetHookAddress(c.function.Entry, c.function.End, c.stateID)
	if err != nil {
		return nil, err
	}
	hook, err := c.buildHook(hookAddr)
	if err != nil {
		return nil, err
	}
	return safe_hook_installer.NewHookWriter(hookAddr, hook), nil
}

func (c *breakpointFlowRunner) buildHook(hookAddr uintptr) ([]byte, rookoutErrors.RookoutError) {
	if c.IsUnhookedState() {
		return c.function.PatchedBytes, nil
	}

	hook, err := assembler.EncodeJmp(hookAddr, c.jumpDestination)
	if err != nil {
		return nil, err
	}
	if c.function.PatchedBytes == nil {
		
		c.function.PatchedBytes = make([]byte, len(hook))
		for i := range hook {
			c.function.PatchedBytes[i] = *((*byte)(unsafe.Pointer(hookAddr + uintptr(i))))
		}
	}
	return hook, nil
}

func (c *breakpointFlowRunner) IsUnhookedState() bool {
	return c.stateID == c.UnhookedFunctionStateID()
}

func (c *breakpointFlowRunner) ID() int {
	return c.stateID
}

func (c *breakpointFlowRunner) UnhookedFunctionStateID() int {
	return 0
}

func NewFlowRunner(nativeAPI NativeHookerAPI, initInfo BreakpointFlowRunnerInitializationInfo, requiredBreakpoints []*augs.BreakpointInstance, installerFactory safe_hook_installer.SafeHookInstallerCreator) (*breakpointFlowRunner, error) {
	stateID, err := nativeAPI.RegisterFunctionBreakpointsState(initInfo.Function.Entry, initInfo.Function.End, requiredBreakpoints, initInfo.BPCallback, initInfo.Function.Prologue, initInfo.Function.StackFrameSize > 0)
	if err != nil {
		return nil, err
	}

	runner := &breakpointFlowRunner{
		installerCreator: installerFactory,
		info:             initInfo,
		stateID:          stateID,
		function:         initInfo.Function,
		nativeAPI:        nativeAPI,
	}
	if runner.stateID == initInfo.Function.FunctionCopyStateID {
		
		
		runner.stateID = runner.UnhookedFunctionStateID()
	}
	return runner, nil
}
