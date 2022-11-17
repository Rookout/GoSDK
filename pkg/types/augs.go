package types

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type ConditionCreatorFunc func(string) (Condition, rookoutErrors.RookoutError)

type Condition interface {
	Evaluate(namespace Namespace) (bool, rookoutErrors.RookoutError)
}
