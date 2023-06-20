package instrumentation

import (
	"strings"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/locations_set"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
	"github.com/google/uuid"
)

type functionGetter func(addr uint64) (*binary_info.Function, *augs.Function, rookoutErrors.RookoutError)

type breakpointWriter struct {
	filename    string
	lineno      int
	getFunction functionGetter
	storage     *locations_set.BreakpointStorage
	binaryInfo  *binary_info.BinaryInfo
	hooker      hooker.Hooker
	addrs       []uint64
}

func writeBreakpoint(filename string, lineno int, getFunction functionGetter, addrs []uint64, storage *locations_set.BreakpointStorage, binaryInfo *binary_info.BinaryInfo, hooker hooker.Hooker) (*augs.Breakpoint, rookoutErrors.RookoutError) {
	breakpointWriter := breakpointWriter{
		filename:    filename,
		lineno:      lineno,
		getFunction: getFunction,
		storage:     storage,
		binaryInfo:  binaryInfo,
		hooker:      hooker,
		addrs:       addrs,
	}
	return breakpointWriter.writeBreakpoint()
}

func (b *breakpointWriter) writeBreakpoint() (*augs.Breakpoint, rookoutErrors.RookoutError) {
	breakpointID, rookErr := createBreakpointID()
	if rookErr != nil {
		return nil, rookErr
	}
	bp := &augs.Breakpoint{
		File:       b.filename,
		Line:       b.lineno,
		Stacktrace: maxStackFrames,
		Name:       "rookout" + breakpointID,
	}

	for _, addr := range b.addrs {
		bpInstance, err := b.writeBreakpointInstance(bp, addr)
		if err != nil {
			return nil, err
		}
		bp.Instances = append(bp.Instances, bpInstance)
	}
	return bp, nil
}

func (b *breakpointWriter) writeBreakpointInstance(bp *augs.Breakpoint, addr uint64) (*augs.BreakpointInstance, rookoutErrors.RookoutError) {
	if bpInstance, ok := b.storage.FindBreakpointByAddr(addr); ok {
		return bpInstance, nil
	}

	biFunction, function, rookErr := b.getFunction(addr)
	if rookErr != nil {
		return nil, rookErr
	}
	bpInstance := augs.NewBreakpointInstance(addr, bp, function)

	logger.Logger().Debugf("Adding breakpoint in %s:%d address=0x%x", b.filename, b.lineno, addr)

	flowRunner, err := b.hooker.StartWritingBreakpoint(bpInstance)
	if err != nil {
		return nil, rookoutErrors.NewFailedToAddBreakpoint(b.filename, b.lineno, err)
	}

	rookErr = applyBreakpointState(flowRunner, b.filename, b.lineno)
	if rookErr != nil {
		return nil, rookErr
	}

	variableLocators, err := variable.GetVariableLocators(addr, b.lineno, biFunction, b.binaryInfo)
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetVariableLocators(b.filename, b.lineno, err)
	}
	bpInstance.VariableLocators = variableLocators

	b.storage.AddBreakpointInstance(bpInstance)
	return bpInstance, nil
}

type breakpointEraser struct {
	breakpoint *augs.Breakpoint
	hooker     hooker.Hooker
	storage    *locations_set.BreakpointStorage
}

func eraseBreakpoint(breakpoint *augs.Breakpoint, hooker hooker.Hooker, storage *locations_set.BreakpointStorage) rookoutErrors.RookoutError {
	breakpointEraser := &breakpointEraser{
		breakpoint: breakpoint,
		hooker:     hooker,
		storage:    storage,
	}
	return breakpointEraser.eraseBreakpoint()
}

func (b *breakpointEraser) eraseBreakpoint() rookoutErrors.RookoutError {
	var remainingInstances []*augs.BreakpointInstance
	for _, bpInstance := range b.breakpoint.Instances {
		err := b.eraseBreakpointInstance(bpInstance)
		if err != nil {
			logger.Logger().WithError(err).Warningf("Failed to erase an instance of the breakpoint")
			remainingInstances = append(remainingInstances, bpInstance)
		}
	}

	if remainingInstances != nil {
		b.breakpoint.Instances = remainingInstances
		return rookoutErrors.NewFailedToEraseAllBreakpointInstances()
	}
	return nil
}

func (b *breakpointEraser) eraseBreakpointInstance(bpInstance *augs.BreakpointInstance) rookoutErrors.RookoutError {
	flowRunner, err := b.hooker.StartErasingBreakpoint(bpInstance)
	if err != nil {
		return rookoutErrors.NewFailedToRemoveBreakpoint(b.breakpoint.File, b.breakpoint.Line, err)
	}

	rookErr := applyBreakpointState(flowRunner, b.breakpoint.File, b.breakpoint.Line)
	if rookErr != nil {
		return rookErr
	}
	b.storage.RemoveBreakpointInstance(bpInstance)

	return nil
}

func createBreakpointID() (string, rookoutErrors.RookoutError) {
	breakpointID, err := uuid.NewUUID()
	if err != nil {
		return "", rookoutErrors.NewFailedToCreateID(err)
	}
	return strings.ReplaceAll(breakpointID.String(), "-", ""), nil
}


func applyBreakpointState(runner hooker.BreakpointFlowRunner, filename string, lineno int) rookoutErrors.RookoutError {
	if !runner.IsDefaultState() {
		addressMappings, offsetMappings, err := runner.GetAddressMapping()
		if err != nil {
			return rookoutErrors.NewFailedToGetAddressMapping(filename, lineno, err)
		}

		
		
		
		if err = module.PatchModuleData(addressMappings, offsetMappings, runner.ID()); err != nil {
			return rookoutErrors.NewFailedToPatchModule(filename, lineno, err)
		}
	}

	err := runner.ApplyBreakpointsState()
	if err != nil {
		return rookoutErrors.NewFailedToApplyBreakpointState(filename, lineno, err)
	}

	return nil
}
