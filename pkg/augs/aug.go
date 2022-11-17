package augs

import (
	"github.com/Rookout/GoSDK/pkg/augs/actions"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/google/uuid"
)

type Aug interface {
	Execute(collectionService *collection.CollectionService)
	GetAugId() types.AugId
	SetCondition(condition types.Condition)
	SetLimitsManager(manager LimitsManager)
}

type aug struct {
	augId         types.AugId
	action        actions.Action
	output        com_ws.Output
	condition     types.Condition
	limitsManager LimitsManager
}

func NewAug(augId types.AugId, action actions.Action, output com_ws.Output) Aug {
	return &aug{
		augId:  augId,
		action: action,
		output: output,
	}
}

func (a *aug) GetAugId() types.AugId {
	return a.augId
}

func (a *aug) SetCondition(condition types.Condition) {
	a.condition = condition
}

func (a *aug) SetLimitsManager(manager LimitsManager) {
	a.limitsManager = manager
}

func (a *aug) execute(collectionService *collection.CollectionService, reportId string) {
	namespace, err := newAugNamespace(collectionService)
	if err != nil {
		logger.Logger().WithError(err).Warningf("Error while executing aug: %s\n", a.augId)
		_ = a.output.SendWarning(a.augId, err)
		return
	}

	if a.condition != nil {
		res, err := a.condition.Evaluate(namespace.GetAugNamespace())
		if err != nil {
			logger.Logger().WithError(err).Warningf("Error while executing condition on aug: %s", a.augId)
			_ = a.output.SendWarning(a.augId, err)
		}

		if !res {
			return
		}
	}

	err = a.action.Execute(a.augId, reportId, namespace.GetAugNamespace(), a.output)
	if err != nil {
		logger.Logger().WithError(err).Warningf("Error while executing aug: %s", a.augId)
		_ = a.output.SendWarning(a.augId, err)
		return
	}
}

func (a *aug) Execute(collectionService *collection.CollectionService) {
	executionId := uuid.NewString()
	if a.limitsManager == nil {
		a.execute(collectionService, executionId)
		return
	}

	if len(a.limitsManager.GetAllLimiters()) == 0 {
		logger.Logger().Warningf("Aug (%s) has no limiters", a.augId)
	}

	if ok := a.limitsManager.BeforeRun(executionId); ok {
		a.execute(collectionService, executionId)
		a.limitsManager.AfterRun(executionId)
	}
}
