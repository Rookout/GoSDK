package augs

import (
	"github.com/Rookout/GoSDK/pkg/augs/actions"
	"github.com/Rookout/GoSDK/pkg/augs/conditions"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/google/uuid"
)

type Aug interface {
	Execute(collectionService *collection.CollectionService)
	GetAugID() types.AugID
	SetCondition(condition conditions.Condition)
	SetLimitsManager(manager LimitsManager)
}

type aug struct {
	augID         types.AugID
	action        actions.Action
	output        com_ws.Output
	condition     conditions.Condition
	limitsManager LimitsManager
	executed      bool
}

func NewAug(augID types.AugID, action actions.Action, output com_ws.Output) Aug {
	return &aug{
		augID:    augID,
		action:   action,
		output:   output,
		executed: false,
	}
}

func (a *aug) GetAugID() types.AugID {
	return a.augID
}

func (a *aug) SetCondition(condition conditions.Condition) {
	a.condition = condition
}

func (a *aug) SetLimitsManager(manager LimitsManager) {
	a.limitsManager = manager
}

func (a *aug) execute(collectionService *collection.CollectionService, reportID string) {
	namespace, err := newAugNamespace(collectionService)
	if err != nil {
		logger.Logger().WithError(err).Warningf("Error while executing aug: %s\n", a.augID)
		_ = a.output.SendWarning(a.augID, err)
		return
	}

	if a.condition != nil {
		res, err := a.condition.Evaluate(namespace.GetAugNamespace())
		if err != nil {
			logger.Logger().WithError(err).Warningf("Error while executing condition on aug: %s", a.augID)
			_ = a.output.SendWarning(a.augID, err)
		}

		if !res {
			return
		}
	}

	a.executed = true
	err = a.action.Execute(a.augID, reportID, namespace.GetAugNamespace(), a.output)
	if err != nil {
		logger.Logger().WithError(err).Warningf("Error while executing aug: %s", a.augID)
		_ = a.output.SendWarning(a.augID, err)
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
		logger.Logger().Warningf("Aug (%s) has no limiters", a.augID)
	}

	shouldSkipLimiters := (!a.executed) && (a.condition == nil)
	if ok := a.limitsManager.BeforeRun(executionId, shouldSkipLimiters); ok {
		a.execute(collectionService, executionId)
		a.limitsManager.AfterRun(executionId)
	}
}
