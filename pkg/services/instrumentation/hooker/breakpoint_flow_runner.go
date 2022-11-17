package hooker

import (
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/callstack"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_installer"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"
	"github.com/Rookout/GoSDK/pkg/services/suspender"
	"github.com/Rookout/GoSDK/pkg/types"
	"time"
	"unsafe"
)

const (
	installHookNumAttempts       = 10    
	installHookAttemptDelayMS    = 10    
	installHookWatchDogTimeoutMS = 60000 
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
	functionEntry             types.Address
	functionEnd               types.Address
	breakpointAddress         types.Address
	bpCallback                uintptr
	prologueCallback          uintptr
	shouldRunPrologueCallback uintptr
	functionStackFrameSize    int32
}

type baseBreakpointFlowRunner struct {
	installerCreator safe_hook_installer.SafeHookInstallerCreator
	hooker           hookerManipulator
	info             BreakpointFlowRunnerInitializationInfo
	stateId          int
}

func (c *baseBreakpointFlowRunner) GetAddressMapping() (unsafe.Pointer, error) {
	mapping, err := c.hooker.getNativeAPI().GetInstructionMapping(c.info.functionEntry, c.info.functionEnd, c.stateId)
	return unsafe.Pointer(mapping), err
}

func (c *baseBreakpointFlowRunner) GetUnpatchedAddressMapping() (unsafe.Pointer, error) {
	mapping, err := c.hooker.getNativeAPI().GetUnpatchedInstructionMapping(c.info.functionEntry, c.info.functionEnd)
	return unsafe.Pointer(mapping), err
}

func (c *baseBreakpointFlowRunner) applyBreakpointsState() error {
	var err error = nil
	var shi safe_hook_installer.SafeHookInstaller
	var numGoroutines int
	for attempt := 1; attempt <= installHookNumAttempts; attempt++ {
		if attempt > 1 {
			
			time.Sleep(installHookAttemptDelayMS * time.Millisecond)
		}
		shi, err = c.installerCreator(c.info.functionEntry, c.info.functionEnd, c.stateId,
			c.hooker.getNativeAPI(), suspender.GetSuspender(), callstack.NewStackTraceBuffer(), &safe_hook_validator.ValidatorFactoryImpl{})
		if err != nil {
			
			break
		}
		err = c.hooker.getNativeAPI().TriggerWatchDog(installHookWatchDogTimeoutMS)
		if err != nil {
			logger.Logger().Warningf("Failed to trigger the watchdog on attempt #%d. Reason: %v", attempt, err)
			continue
		}
		numGoroutines, err = shi.InstallHook()
		c.hooker.getNativeAPI().DefuseWatchDog()
		if err != nil {
			logger.Logger().Warningf("Failed to install the hook on attempt #%d. Detected %d non-system goroutines. Reason: %v", attempt, numGoroutines, err)
			continue
		}
		
		break
	}
	if err == nil {
		
		notifyErr := c.hooker.getNativeAPI().ApplyBreakpointsState(c.info.functionEntry, c.info.functionEnd, c.stateId)
		if notifyErr != nil {
			logger.Logger().Warningf("Failed to notify the native on installing the breakpoint (%v). However, the breakpoint was installed successfully.", notifyErr)
			
		}
	}
	return err
}

func (c *baseBreakpointFlowRunner) IsPatched() bool {
	return c.stateId > 0
}

func (c *baseBreakpointFlowRunner) ID() int {
	return c.stateId
}

func (c *baseBreakpointFlowRunner) DefaultID() int {
	return 0
}

func newFlowRunner(hooker hookerManipulator, initInfo BreakpointFlowRunnerInitializationInfo, requiredBreakpoints []types.Address, installerFactory safe_hook_installer.SafeHookInstallerCreator) (*baseBreakpointFlowRunner, error) {
	stateId, err := hooker.getNativeAPI().RegisterFunctionBreakpointsState(initInfo.functionEntry, initInfo.functionEnd, requiredBreakpoints, initInfo.bpCallback, initInfo.prologueCallback, initInfo.shouldRunPrologueCallback, initInfo.functionStackFrameSize)
	if err != nil {
		return nil, err
	}
	return &baseBreakpointFlowRunner{
		installerCreator: installerFactory,
		hooker:           hooker,
		info:             initInfo,
		stateId:          stateId,
	}, nil
}

type breakpointAdder struct {
	*baseBreakpointFlowRunner
}

type breakpointRemover struct {
	*baseBreakpointFlowRunner
}

func startAddingBreakpoint(hooker hookerManipulator, initInfo BreakpointFlowRunnerInitializationInfo, installerCreator safe_hook_installer.SafeHookInstallerCreator) (BreakpointFlowRunner, error) {
	allBPs := hooker.getActiveBreakpointsWithNew(initInfo.functionEntry, initInfo.breakpointAddress)
	baseCtxt, err := newFlowRunner(hooker, initInfo, allBPs, installerCreator)
	if err != nil {
		return nil, err
	}
	return &breakpointAdder{baseCtxt}, nil
}

func startRemovingBreakpoint(hooker hookerManipulator, initInfo BreakpointFlowRunnerInitializationInfo, installerCreator safe_hook_installer.SafeHookInstallerCreator) (BreakpointFlowRunner, error) {
	allBPs := hooker.getActiveBreakpointsWithoutOld(initInfo.functionEntry, initInfo.breakpointAddress)
	baseCtxt, err := newFlowRunner(hooker, initInfo, allBPs, installerCreator)
	if err != nil {
		return nil, err
	}
	return &breakpointRemover{baseCtxt}, nil
}

func (c *breakpointAdder) ApplyBreakpointsState() error {
	if err := c.applyBreakpointsState(); err != nil {
		return err
	}
	c.hooker.addBreakpoint(c.info.breakpointAddress, c.info.functionEntry, c.info.functionEnd)
	return nil
}

func (c *breakpointRemover) ApplyBreakpointsState() error {
	if err := c.applyBreakpointsState(); err != nil {
		return err
	}
	c.hooker.removeBreakpoint(c.info.breakpointAddress, c.info.functionEntry)
	return nil
}
