package instrumentation

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker/prologue"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
)

type functionCreator struct {
	trampolineManager *trampolineManager
	hooker            hooker.Hooker
	filename          string
	lineno            int
	biFunction        *binary_info.Function
	binaryInfo        *binary_info.BinaryInfo
}

func NewFunction(function *binary_info.Function, filename string, lineno int, binaryInfo *binary_info.BinaryInfo, hooker hooker.Hooker, trampolineManager *trampolineManager) (*augs.Function, rookoutErrors.RookoutError) {
	functionCreator := &functionCreator{
		trampolineManager: trampolineManager,
		hooker:            hooker,
		biFunction:        function,
		filename:          filename,
		lineno:            lineno,
		binaryInfo:        binaryInfo,
	}
	return functionCreator.createFunction()
}

func (f *functionCreator) createFunction() (*augs.Function, rookoutErrors.RookoutError) {
	finalTrampolinePointer, middleTrampolineAddress, err := f.trampolineManager.getTrampolineAddress()
	if err != nil {
		return nil, err
	}
	stackFrameSize := module.FindFuncMaxSPDelta(f.biFunction.Entry)
	function := augs.NewFunction(f.biFunction.Entry, f.biFunction.End, stackFrameSize, middleTrampolineAddress, finalTrampolinePointer)
	function.FunctionCopyStateID, function.Prologue, err = f.getPrologue(function)
	if err != nil {
		return nil, err
	}

	return function, nil
}

func (f *functionCreator) getPrologue(function *augs.Function) (int, []byte, rookoutErrors.RookoutError) {
	functionCopyStateID, functionCopyEntry, err := f.createFunctionCopy(function)
	if err != nil {
		return -1, nil, err
	}

	getRegsUsed := func() ([]assembler.Reg, rookoutErrors.RookoutError) {
		variableLocators, err := variable.GetVariableLocators(function.Entry, f.lineno, f.biFunction, f.binaryInfo)
		if err != nil {
			return nil, rookoutErrors.NewFailedToGetVariableLocators(f.filename, f.lineno, err)
		}

		var regsUsed []assembler.Reg
		regsUsedMap := make(map[assembler.Reg]struct{})
		for _, locator := range variableLocators {
			regs, err := locator.GetRegsUsed()
			if err != nil {
				logger.Logger().WithError(err).Warningf("Failed to get regs used by variable %s", locator.VariableName)
			}

			
			for _, reg := range regs {
				if _, ok := regsUsedMap[reg]; !ok {
					regsUsedMap[reg] = struct{}{}
					regsUsed = append(regsUsed, reg)
				}
			}
		}
		return regsUsed, nil
	}

	prologueGenerator, err := prologue.NewGenerator(uintptr(f.biFunction.Entry), uintptr(f.biFunction.End), int(function.StackFrameSize), functionCopyEntry, getRegsUsed)
	if err != nil {
		return -1, nil, err
	}
	p, err := prologueGenerator.Generate()
	return functionCopyStateID, p, err
}

func (f *functionCreator) createFunctionCopy(function *augs.Function) (int, uintptr, rookoutErrors.RookoutError) {
	runner, err := f.hooker.StartCopyingFunction(function)
	if err != nil {
		return -1, 0, rookoutErrors.NewFailedToStartCopyingFunction(err)
	}
	addressMappings, offsetMappings, err := runner.GetAddressMapping()
	if err != nil {
		return -1, 0, rookoutErrors.NewFailedToGetAddressMapping(f.filename, f.lineno, err)
	}

	
	
	
	if err = module.PatchModuleData(addressMappings, offsetMappings, runner.ID()); err != nil {
		return -1, 0, rookoutErrors.NewFailedToPatchModule(f.filename, f.lineno, err)
	}

	return runner.ID(), addressMappings[0].NewAddress, nil
}
