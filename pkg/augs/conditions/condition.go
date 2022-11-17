package conditions

import (
	"github.com/Rookout/GoSDK/pkg/processor/paths"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type Condition struct {
	path paths.Path
}

func NewCondition(condition string) (types.Condition, rookoutErrors.RookoutError) {
	path, err := paths.NewArithmeticPath(condition)
	if err != nil {
		return nil, err
	}
	return &Condition{path: path}, nil
}

func (c *Condition) Evaluate(namespace types.Namespace) (bool, rookoutErrors.RookoutError) {
	res, err := c.path.ReadFrom(namespace)
	if err != nil {
		return false, err
	}
	return res.GetObject().(bool), nil
}
