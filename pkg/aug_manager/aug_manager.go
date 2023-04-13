package aug_manager

import (
	"sync"

	"github.com/Rookout/GoSDK/pkg/augs/locations"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/processor"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation"
	"github.com/Rookout/GoSDK/pkg/types"
)

type AugManager interface {
	InitializeAugs(map[types.AugID]types.AugConfiguration)
	AddAug(types.AugConfiguration)
	RemoveAug(types.AugID) error
	ClearAugs()
}

type augManager struct {
	triggerServices *instrumentation.TriggerServices
	output          com_ws.Output
	factory         *locations.LocationFactory

	augIDs     map[types.AugID]interface{}
	augIDsLock sync.Mutex
}

func NewAugManager(triggerServices *instrumentation.TriggerServices, output com_ws.Output) *augManager {
	augFactory := locations.NewLocationFactory(output, processor.NewProcessorFactory())
	return &augManager{triggerServices: triggerServices, output: output, factory: augFactory, augIDs: make(map[types.AugID]interface{}), augIDsLock: sync.Mutex{}}
}

func (a *augManager) InitializeAugs(augConfigs map[types.AugID]types.AugConfiguration) {
	a.augIDsLock.Lock()
	defer a.augIDsLock.Unlock()

	leftovers := make(map[types.AugID]struct{})
	for k := range a.augIDs {
		leftovers[k] = struct{}{}
	}

	for augID, augConf := range augConfigs {
		if _, ok := leftovers[augID]; ok {
			delete(leftovers, augID)
		} else {
			a.addAug(augConf)
		}
	}

	for augID := range leftovers {
		err := a.removeAug(augID)
		if err != nil {
			logger.Logger().WithError(err).Errorf("failed to remove leftover aug (%s)", augID)
		}
	}
}

func (a *augManager) addAug(configuration types.AugConfiguration) {
	aug, err := a.factory.GetLocation(configuration)
	if err != nil {
		logger.Logger().WithError(err).Errorln("Failed to parse aug")

		if augID, ok := configuration["id"].(types.AugID); ok {
			
			_ = a.output.SendRuleStatus(augID, "Error", err)
		}
		return
	}

	if _, exists := a.augIDs[aug.GetAugID()]; exists {
		logger.Logger().Debugf("Aug already exists - %s\n", aug.GetAugID())
		return
	}

	a.triggerServices.GetInstrumentation().AddAug(aug)
	a.augIDs[aug.GetAugID()] = struct{}{}
}

func (a *augManager) AddAug(configuration types.AugConfiguration) {
	a.augIDsLock.Lock()
	defer a.augIDsLock.Unlock()

	a.addAug(configuration)
}

func (a *augManager) removeAug(augID types.AugID) error {
	logger.Logger().Debugf("Removing aug - %s\n", augID)

	err := a.triggerServices.RemoveAug(augID)
	if err != nil {
		return err
	}

	delete(a.augIDs, augID)
	return nil
}

func (a *augManager) RemoveAug(augID types.AugID) error {
	a.augIDsLock.Lock()
	defer a.augIDsLock.Unlock()

	return a.removeAug(augID)
}

func (a *augManager) ClearAugs() {
	logger.Logger().Debugf("Clearing all augs\n")

	var idsCopy []types.AugID
	for k := range a.augIDs {
		idsCopy = append(idsCopy, k)
	}

	for _, augID := range idsCopy {
		err := a.RemoveAug(augID)
		if err != nil {
			logger.Logger().WithError(err).Errorf("failed to clear aug (%s)", augID)
		}
	}

	a.triggerServices.ClearAugs()
}
