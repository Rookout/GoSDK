package instrumentation

import (
	"context"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/locations_set"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/callback"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/google/uuid"
)

const maxStackFrames = 128 

type ProcessManager struct {
	hooker            hooker.Hooker
	binaryInfo        *binary_info.BinaryInfo
	locations         *locations_set.LocationsSet
	trampolineManager *trampolineManager
	cancelGCControl   context.CancelFunc
	cancelBPMonitor   context.CancelFunc
}

func NewProcessManager(locations *locations_set.LocationsSet, breakpointMonitorInterval time.Duration) (pm *ProcessManager, err rookoutErrors.RookoutError) {
	bi, err := createBinaryInfo()
	if err != nil {
		return nil, err
	}

	h, err := createHooker(bi)
	defer func() {
		if err != nil {
			_ = hooker.Destroy()
		}
	}()
	if err != nil {
		return nil, err
	}

	module.Init()
	p := &ProcessManager{hooker: h, binaryInfo: bi, locations: locations, trampolineManager: newTrampolineManager()}

	ctxGCControl, cancelGCControl := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelGCControl()
		}
	}()
	p.cancelGCControl = cancelGCControl
	triggerChan := make(chan bool, 10000)
	gcController := newGCController()
	utils.CreateGoroutine(func() {
		gcController.start(ctxGCControl, triggerChan)
	})
	callback.SetTriggerChan(triggerChan)

	ctxBPMonitor, cancelBPMonitor := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelBPMonitor()
		}
	}()
	p.cancelBPMonitor = cancelBPMonitor
	utils.CreateGoroutine(func() {
		p.monitorBreakpoints(ctxBPMonitor, breakpointMonitorInterval)
	})

	return p, nil
}

func (p *ProcessManager) monitorBreakpoints(ctx context.Context, breakpointMonitorInterval time.Duration) {
	for {
		monitorTimeout := time.NewTimer(breakpointMonitorInterval)
		select {
		case <-ctx.Done():
			return

		case <-monitorTimeout.C:
			bpInstances := p.locations.GetBreakpointInstances()
			for _, bpInstance := range bpInstances {
				failedCounter := atomic.SwapUint64(bpInstance.FailedCounter, 0)
				if failedCounter > 0 {
					locations, ok := p.locations.FindLocationsByBreakpointName(bpInstance.Breakpoint.Name)
					if !ok {
						continue
					}

					for i := range locations {
						locations[i].SetError(rookoutErrors.NewFailedToExecuteBreakpoint(failedCounter))
					}
				}
			}
		}
	}
}

func createHooker(bi *binary_info.BinaryInfo) (hooker.Hooker, rookoutErrors.RookoutError) {
	shouldRunPrologue, err := bi.GetUnwrappedFuncPointer(callback.ShouldRunPrologue)
	if err != nil {
		return nil, err
	}
	moreStack, err := bi.GetUnwrappedFuncPointer(callback.MoreStack)
	if err != nil {
		return nil, err
	}

	h := hooker.New(unsafe.Pointer(reflect.ValueOf(callback.Callback).Pointer()),
		unsafe.Pointer(moreStack),
		unsafe.Pointer(shouldRunPrologue))

	err = hooker.Init(funcForInit)
	if err != nil {
		logger.Logger().WithError(err).Errorf("Unable to start hooker")
		return nil, err
	}

	return h, nil
}

func createBinaryInfo() (*binary_info.BinaryInfo, rookoutErrors.RookoutError) {
	bi := binary_info.NewBinaryInfo()
	exec, err := os.Executable()
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetExecutable(err)
	}
	err = bi.LoadBinaryInfo(exec, binary_info.GetEntrypoint(exec), nil)
	if err != nil {
		return nil, rookoutErrors.NewFailedToLoadBinaryInfo(err)
	}
	bi.Dwarf = bi.Images[0].Dwarf 

	callback.SetBinaryInfo(bi)
	return bi, nil
}

func (p *ProcessManager) LineToPC(filename string, lineno int) ([]uint64, string, *binary_info.Function, rookoutErrors.RookoutError) {
	return p.binaryInfo.FindFileLocation(filename, lineno)
}

func (p *ProcessManager) WriteBreakpoint(filename string, lineno int, function *binary_info.Function, addrs []uint64) (*augs.Breakpoint, rookoutErrors.RookoutError) {
	if bp, ok := p.locations.FindBreakpointByAddrs(addrs); ok {
		return bp, nil
	}

	breakpointID, rookErr := createBreakpointID()
	if rookErr != nil {
		return nil, rookErr
	}
	bp := &augs.Breakpoint{
		FunctionName: function.Name,
		File:         filename,
		Line:         lineno,
		Stacktrace:   maxStackFrames,
		Name:         "rookout" + breakpointID,
	}

	for _, addr := range addrs {
		bpInstance, err := p.writeBreakpointInstance(bp, addr)
		if err != nil {
			return nil, err
		}
		bp.Instances = append(bp.Instances, bpInstance)
	}

	return bp, nil
}

func (p *ProcessManager) getFunction(biFunction *binary_info.Function) (*augs.Function, rookoutErrors.RookoutError) {
	if function, ok := p.locations.FindFunctionByEntry(biFunction.Entry); ok {
		return function, nil
	}

	finalTrampolinePointer, middleTrampolineAddress, err := p.trampolineManager.getTrampolineAddress()
	if err != nil {
		return nil, err
	}
	function := augs.NewFunction(biFunction.Entry, biFunction.End, module.FindFuncMaxSPDelta(biFunction.Entry), middleTrampolineAddress, finalTrampolinePointer)

	p.locations.AddFunction(function)
	return function, nil
}

func (p *ProcessManager) writeBreakpointInstance(bp *augs.Breakpoint, addr uint64) (*augs.BreakpointInstance, rookoutErrors.RookoutError) {
	if bpInstance, ok := p.locations.FindBreakpointByAddr(addr); ok {
		return bpInstance, nil
	}

	filename, lineno, biFunction := p.binaryInfo.PCToLine(addr)
	function, rookErr := p.getFunction(biFunction)
	if rookErr != nil {
		return nil, rookErr
	}
	bpInstance := augs.NewBreakpointInstance(addr, bp)
	bpInstance.Function = function

	logger.Logger().Debugf("Adding breakpoint in %s:%d address=0x%x", filename, lineno, addr)

	flowRunner, err := p.hooker.StartWritingBreakpoint(bpInstance)
	if err != nil {
		return nil, rookoutErrors.NewFailedToAddBreakpoint(filename, lineno, err)
	}

	addressMappings, offsetMappings, err := flowRunner.GetAddressMapping()
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetAddressMapping(filename, lineno, err)
	}

	unpatchedAddressMappings, unpatchedOffsetMappings, err := flowRunner.GetUnpatchedAddressMapping()
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetUnpatchedAddressMapping(filename, lineno, err)
	}

	
	
	
	if err = module.PatchModuleData(unpatchedAddressMappings, unpatchedOffsetMappings, flowRunner.DefaultID()); err != nil {
		return nil, rookoutErrors.NewFailedToPatchModule(filename, lineno, err)
	}
	if err = module.PatchModuleData(addressMappings, offsetMappings, flowRunner.ID()); err != nil {
		return nil, rookoutErrors.NewFailedToPatchModule(filename, lineno, err)
	}

	err = flowRunner.ApplyBreakpointsState()
	if err != nil {
		return nil, rookoutErrors.NewFailedToApplyBreakpointState(filename, lineno, err)
	}

	variableLocators, err := variable.GetVariableLocators(addr, lineno, biFunction, p.binaryInfo)
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetVariableLocators(filename, lineno, err)
	}
	bpInstance.VariableLocators = variableLocators

	p.locations.AddBreakpointInstance(bpInstance)
	return bpInstance, nil
}

func (p *ProcessManager) EraseBreakpoint(bp *augs.Breakpoint) rookoutErrors.RookoutError {
	var remainingInstances []*augs.BreakpointInstance
	for _, bpInstance := range bp.Instances {
		err := p.eraseBreakpointInstance(bp, bpInstance)
		if err != nil {
			logger.Logger().WithError(err).Warningf("Failed to erase an instance of the breakpoint")
			remainingInstances = append(remainingInstances, bpInstance)
		}
	}

	if remainingInstances != nil {
		bp.Instances = remainingInstances
		return rookoutErrors.NewFailedToEraseAllBreakpointInstances()
	}
	return nil
}

func (p *ProcessManager) eraseBreakpointInstance(bp *augs.Breakpoint, bpInstance *augs.BreakpointInstance) rookoutErrors.RookoutError {
	flowRunner, err := p.hooker.StartErasingBreakpoint(bpInstance)
	if err != nil {
		return rookoutErrors.NewFailedToRemoveBreakpoint(bp.File, bp.Line, err)
	}

	if !flowRunner.IsDefaultState() {
		addressMappings, offsetMappings, err := flowRunner.GetAddressMapping()
		if err != nil {
			return rookoutErrors.NewFailedToGetAddressMapping(bp.File, bp.Line, err)
		}

		err = module.PatchModuleData(addressMappings, offsetMappings, flowRunner.ID())
		if err != nil {
			return rookoutErrors.NewFailedToPatchModule(bp.File, bp.Line, err)
		}
	}

	if err = flowRunner.ApplyBreakpointsState(); err != nil {
		return rookoutErrors.NewFailedToApplyBreakpointState(bp.File, bp.Line, err)
	}

	p.locations.RemoveBreakpointInstance(bpInstance)

	return nil
}

func createBreakpointID() (string, rookoutErrors.RookoutError) {
	breakpointID, err := uuid.NewUUID()
	if err != nil {
		return "", rookoutErrors.NewFailedToCreateID(err)
	}
	return strings.ReplaceAll(breakpointID.String(), "-", ""), nil
}


func (_ *ProcessManager) Clean() rookoutErrors.RookoutError {
	err := hooker.Destroy()
	if err != nil {
		return rookoutErrors.NewFailedToDestroyNative(err)
	}
	return nil
}
