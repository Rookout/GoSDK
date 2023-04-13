package conditions

import (
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/processor/paths"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type ConditionCreatorFunc func(string) (Condition, rookoutErrors.RookoutError)

type Condition interface {
	Evaluate(namespace namespaces.Namespace) (bool, rookoutErrors.RookoutError)
}

type condition struct {
	path paths.Path
}

func NewCondition(conditionString string) (Condition, rookoutErrors.RookoutError) {
	path, err := paths.NewArithmeticPath(conditionString)
	if err != nil {
		return nil, err
	}
	return &condition{path: path}, nil
}

func (c *condition) Evaluate(namespace namespaces.Namespace) (bool, rookoutErrors.RookoutError) {
	res, err := c.path.ReadFrom(namespace)
	if err != nil {
		return false, err
	}
	return res.GetObject().(bool), nil
}
