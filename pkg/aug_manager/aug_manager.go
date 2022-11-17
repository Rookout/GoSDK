package aug_manager

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/augs/locations"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/processor"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation"
	"github.com/Rookout/GoSDK/pkg/types"
	"sync"
)

type AugManager interface {
	UpdateConfig(config.LocationsConfiguration)
	InitializeAugs(map[types.AugId]types.AugConfiguration)
	AddAug(types.AugConfiguration)
	RemoveAug(types.AugId) error
	ClearAugs()
}

type augManager struct {
	triggerServices *instrumentation.TriggerServices
	output          com_ws.Output
	factory         *locations.LocationFactory

	augIds     map[types.AugId]interface{}
	augIdsLock sync.Mutex
}

func NewAugManager(triggerServices *instrumentation.TriggerServices, output com_ws.Output, config config.LocationsConfiguration) *augManager {
	augFactory := locations.NewLocationFactory(output, processor.NewProcessorFactory(), config)
	return &augManager{triggerServices: triggerServices, output: output, factory: augFactory, augIds: make(map[types.AugId]interface{}), augIdsLock: sync.Mutex{}}
}

func (a *augManager) UpdateConfig(config config.LocationsConfiguration) {
	a.factory.UpdateConfig(config)
	augs.GetLimitProvider().UpdateConfig(config)
}

func (a *augManager) InitializeAugs(augConfigs map[types.AugId]types.AugConfiguration) {
	a.augIdsLock.Lock()
	defer a.augIdsLock.Unlock()

	leftovers := make(map[types.AugId]struct{})
	for k := range a.augIds {
		leftovers[k] = struct{}{}
	}

	for augId, augConf := range augConfigs {
		if _, ok := leftovers[augId]; ok {
			delete(leftovers, augId)
		} else {
			a.addAug(augConf)
		}
	}

	for augId := range leftovers {
		err := a.removeAug(augId)
		if err != nil {
			logger.Logger().WithError(err).Errorf("failed to remove leftover aug (%s)", augId)
		}
	}
}

func (a *augManager) addAug(configuration types.AugConfiguration) {
	aug, err := a.factory.GetLocation(configuration)
	if err != nil {
		logger.Logger().WithError(err).Errorln("Failed to parse aug")

		if augId, ok := configuration["id"].(types.AugId); ok {
			
			_ = a.output.SendRuleStatus(augId, "Error", err)
		}
		return
	}

	if _, exists := a.augIds[aug.GetAugId()]; exists {
		logger.Logger().Debugf("Aug already exists - %s\n", aug.GetAugId())
		return
	}

	a.triggerServices.GetInstrumentation().AddAug(aug)
	a.augIds[aug.GetAugId()] = struct{}{}
}

func (a *augManager) AddAug(configuration types.AugConfiguration) {
	a.augIdsLock.Lock()
	defer a.augIdsLock.Unlock()

	a.addAug(configuration)
}

func (a *augManager) removeAug(augId types.AugId) error {
	logger.Logger().Debugf("Removing aug - %s\n", augId)

	err := a.triggerServices.RemoveAug(augId)
	if err != nil {
		return err
	}

	delete(a.augIds, augId)
	return nil
}

func (a *augManager) RemoveAug(augId types.AugId) error {
	a.augIdsLock.Lock()
	defer a.augIdsLock.Unlock()

	return a.removeAug(augId)
}

func (a *augManager) ClearAugs() {
	logger.Logger().Debugf("Clearing all augs\n")

	var idsCopy []types.AugId
	for k := range a.augIds {
		idsCopy = append(idsCopy, k)
	}

	for _, augId := range idsCopy {
		err := a.RemoveAug(augId)
		if err != nil {
			logger.Logger().WithError(err).Errorf("failed to clear aug (%s)", augId)
		}
	}

	a.triggerServices.ClearAugs()
}
