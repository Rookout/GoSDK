package processor

import (
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/processor/operations"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type operationFactory interface {
	GetOperation(configuration types.AugConfiguration) (operations.Operation, rookoutErrors.RookoutError)
}

type processor struct {
	operationList []operations.Operation
}

func NewProcessor(configuration []types.AugConfiguration, factory operationFactory) (*processor, rookoutErrors.RookoutError) {
	var operationList []operations.Operation
	for _, rawOperation := range configuration {
		operation, err := factory.GetOperation(rawOperation)
		if err != nil {
			return nil, err
		}
		operationList = append(operationList, operation)
	}
	return &processor{operationList: operationList}, nil
}

func (p *processor) Process(namespace namespaces.Namespace) {
	for _, operation := range p.operationList {
		operation.Execute(namespace)
	}
}
