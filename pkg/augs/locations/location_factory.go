package locations

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/augs/actions"
	"github.com/Rookout/GoSDK/pkg/augs/conditions"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"sync/atomic"
	"unsafe"
)

type LocationFactory struct {
	config                    *config.LocationsConfiguration
	output                    com_ws.Output
	processorFactory          actions.ProcessorFactory
	ConditionCreator          types.ConditionCreatorFunc
	AugCreator                AugCreatorFunc
	LocationFileLineCreator   LocationFileLineCreatorFunc
	ActionRunProcessorCreator ActionRunProcessorCreatorFunc
}
type ActionRunProcessorCreatorFunc func(configuration types.AugConfiguration, factory actions.ProcessorFactory) (actions.Action, rookoutErrors.RookoutError)
type AugCreatorFunc func(types.AugId, actions.Action, com_ws.Output) augs.Aug
type LocationFileLineCreatorFunc func(types.AugConfiguration, com_ws.Output, augs.Aug) (Location, rookoutErrors.RookoutError)

func NewLocationFactory(output com_ws.Output, processorFactory actions.ProcessorFactory, config config.LocationsConfiguration) *LocationFactory {
	augs.GetLimitProvider().UpdateConfig(config)
	return &LocationFactory{
		config:                    &config,
		output:                    output,
		processorFactory:          processorFactory,
		ConditionCreator:          conditions.NewCondition,
		AugCreator:                augs.NewAug,
		LocationFileLineCreator:   NewLocationFileLine,
		ActionRunProcessorCreator: actions.NewActionRunProcessor,
	}
}

func (l *LocationFactory) UpdateConfig(config config.LocationsConfiguration) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&l.config)), unsafe.Pointer(&config))
}

func (l *LocationFactory) GetLocation(configuration types.AugConfiguration) (Location, rookoutErrors.RookoutError) {
	var err error

	augId, ok := configuration["id"].(string)
	if !ok {
		return nil, rookoutErrors.NewRookAugInvalidKey("id", configuration)
	}

	actionConfig, ok := configuration["action"].(map[string]interface{})
	if !ok {
		return nil, rookoutErrors.NewRookAugInvalidKey("action", configuration)
	}

	action, err := l.ActionRunProcessorCreator(actionConfig, l.processorFactory)
	if err != nil {
		return nil, err.(rookoutErrors.RookoutError)
	}

	aug := l.AugCreator(augId, action, l.output)

	limitsManager, err := augs.GetLimitProvider().GetLimitManager(configuration, augId, l.output)
	if err != nil {
		return nil, err.(rookoutErrors.RookoutError)
	}

	aug.SetLimitsManager(limitsManager)

	conditionConfig, ok := configuration["conditional"].(string)
	if ok {
		condition, err := l.ConditionCreator(conditionConfig)
		if err != nil {
			return nil, err
		}
		aug.SetCondition(condition)
	}

	locationConfig, ok := configuration["location"].(map[string]interface{})
	if !ok {
		return nil, rookoutErrors.NewRookAugInvalidKey("location", configuration)
	}

	return l.getLocation(locationConfig, aug)
}

func (l *LocationFactory) getLocation(configuration types.AugConfiguration, aug augs.Aug) (Location, rookoutErrors.RookoutError) {
	name, ok := configuration["name"].(string)
	if !ok {
		return nil, rookoutErrors.NewRookObjectNameMissing(configuration)
	}

	switch name {
	case "file_line":
		return l.LocationFileLineCreator(configuration, l.output, aug)
	
	default:
		return nil, rookoutErrors.NewRookUnsupportedLocation(name)
	}
}
