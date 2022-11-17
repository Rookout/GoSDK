package processor

import (
	"github.com/Rookout/GoSDK/pkg/augs/actions"
	"github.com/Rookout/GoSDK/pkg/processor/operations"
	"github.com/Rookout/GoSDK/pkg/processor/paths"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

//goland:noinspection GoNameStartsWithPackageName
type processorFactory struct {
}

func NewProcessorFactory() *processorFactory {
	return &processorFactory{}
}

func (p *processorFactory) GetPath(path string) (paths.Path, rookoutErrors.RookoutError) {
	return paths.NewArithmeticPath(path)
}

func (p *processorFactory) GetOperation(configuration types.AugConfiguration) (operations.Operation, rookoutErrors.RookoutError) {
	return operations.NewSet(configuration, p)
}

func (p *processorFactory) GetProcessor(configuration []types.AugConfiguration) (actions.Processor, rookoutErrors.RookoutError) {
	return NewProcessor(configuration, p)
}
