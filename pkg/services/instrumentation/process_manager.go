package instrumentation

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/callback"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
	"github.com/google/uuid"
	"os"
	"reflect"
	"strings"
	"unsafe"
)


type ProcessManager struct {
	hooker     hooker.Hooker
	binaryInfo *binary_info.BinaryInfo
}

func NewProcessManager() (pm *ProcessManager, err rookoutErrors.RookoutError) {
	h := hooker.New(unsafe.Pointer(reflect.ValueOf(callback.PrepForCallback).Pointer()),
		unsafe.Pointer(reflect.ValueOf(callback.MoreStack).Pointer()),
		unsafe.Pointer(reflect.ValueOf(callback.ShouldRunPrologue).Pointer()),
		module.FindFuncMaxSPDelta)
	defer func() {
		if err != nil {
			_ = hooker.Destroy()
		}
	}()

	err = hooker.Init(funcForInit)
	if err != nil {
		logger.Logger().WithError(err).Errorf("Unable to start hooker")
		return nil, err
	}

	stackUsageMap, err := h.GetStackUsageMap()
	if err != nil {
		return nil, err
	}
	module.Init(stackUsageMap)
	bi, err := createBinaryInfo()
	if err != nil {
		return nil, err
	}

	return &ProcessManager{hooker: h, binaryInfo: bi}, nil
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
	breakpointId, rookErr := createBreakpointId()
	if rookErr != nil {
		return nil, rookErr
	}
	bp := &augs.Breakpoint{
		FunctionName: function.Name,
		File:         filename,
		Line:         lineno,
		Stacktrace:   callback.MaxStacktrace,
		Name:         "rookout" + breakpointId,
	}

	for _, addr := range addrs {
		bpInstance, err := p.writeBreakpointInstance(filename, lineno, function, addr)
		if err != nil {
			return nil, err
		}
		bp.Instances = append(bp.Instances, bpInstance)
	}

	return bp, nil
}

func (p *ProcessManager) writeBreakpointInstance(filename string, lineno int, function *binary_info.Function, addr uint64) (*augs.BreakpointInstance, rookoutErrors.RookoutError) {
	filename, lineno, function = p.binaryInfo.PCToLine(addr)
	logger.Logger().Debugf("Adding breakpoint in %s:%d address=0x%x", filename, lineno, addr)
	flowRunner, err := p.hooker.StartWritingBreakpoint(addr, function.Entry, function.End)
	if err != nil {
		return nil, rookoutErrors.NewFailedToAddBreakpoint(filename, lineno, err)
	}

	rawAddressMapping, err := flowRunner.GetAddressMapping()
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetAddressMapping(filename, lineno, err)
	}

	rawUnpatchedAddressMapping, err := flowRunner.GetUnpatchedAddressMapping()
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetUnpatchedAddressMapping(filename, lineno, err)
	}

	
	
	
	if err = module.PatchModuleData(addr, rawUnpatchedAddressMapping, flowRunner.DefaultID()); err != nil {
		return nil, rookoutErrors.NewFailedToPatchModule(filename, lineno, err)
	}
	if err = module.PatchModuleData(addr, rawAddressMapping, flowRunner.ID()); err != nil {
		return nil, rookoutErrors.NewFailedToPatchModule(filename, lineno, err)
	}

	err = flowRunner.ApplyBreakpointsState()
	if err != nil {
		return nil, rookoutErrors.NewFailedToApplyBreakpointState(filename, lineno, err)
	}

	variableLocators, err := variable.GetVariableLocators(addr, lineno, function, p.binaryInfo)
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetVariableLocators(filename, lineno, err)
	}

	return &augs.BreakpointInstance{
		Addr:             addr,
		VariableLocators: variableLocators,
	}, nil
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
	flowRunner, err := p.hooker.StartErasingBreakpoint(bpInstance.Addr)
	if err != nil {
		return rookoutErrors.NewFailedToRemoveBreakpoint(bp.File, bp.Line, err)
	}

	if flowRunner.IsPatched() {
		rawAddressMapping, err := flowRunner.GetAddressMapping()
		if err != nil {
			return rookoutErrors.NewFailedToGetAddressMapping(bp.File, bp.Line, err)
		}

		err = module.PatchModuleData(bpInstance.Addr, rawAddressMapping, flowRunner.ID())
		if err != nil {
			return rookoutErrors.NewFailedToPatchModule(bp.File, bp.Line, err)
		}
	}

	if err = flowRunner.ApplyBreakpointsState(); err != nil {
		return rookoutErrors.NewFailedToApplyBreakpointState(bp.File, bp.Line, err)
	}

	return nil
}

func createBreakpointId() (string, rookoutErrors.RookoutError) {
	breakpointId, err := uuid.NewUUID()
	if err != nil {
		return "", rookoutErrors.NewFailedToCreateID(err)
	}
	return strings.ReplaceAll(breakpointId.String(), "-", ""), nil
}


func (p *ProcessManager) Clean() rookoutErrors.RookoutError {
	err := hooker.Destroy()
	if err != nil {
		return rookoutErrors.NewFailedToDestroyNative(err)
	}
	return nil
}
