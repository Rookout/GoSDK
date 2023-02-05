package instrumentation

import (
	"strings"
	"sync"
	"time"

	"github.com/Rookout/GoSDK/pkg/augs/locations"
	"github.com/Rookout/GoSDK/pkg/locations_set"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/callback"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/pkg/errors"
)

const staleBreakpointClearInterval = 10 * time.Second

type InstrumentationService struct {
	processManager *ProcessManager

	locations           *locations_set.LocationsSet
	staleBreakpointsGC  *ZombieCollector
	instrumentationLock *sync.Mutex

	breakpointFailedCounter uint64
}

func NewInstrumentationService(breakpointMonitorInterval time.Duration) (*InstrumentationService, rookoutErrors.RookoutError) {
	locationsSet := locations_set.NewLocationsSet()
	callback.SetLocationsSet(locationsSet)

	processManager, rookErr := NewProcessManager(locationsSet, breakpointMonitorInterval)
	if rookErr != nil {
		return nil, rookErr
	}

	instrumentationLock := &sync.Mutex{}
	bpGC := NewZombieCollector(staleBreakpointClearInterval, locationsSet, instrumentationLock, processManager.EraseBreakpoint)
	bpGC.Start()
	return &InstrumentationService{locations: locationsSet, staleBreakpointsGC: bpGC, instrumentationLock: instrumentationLock, processManager: processManager}, nil
}

func funcForInit() {
	_ = 3
}

func (i *InstrumentationService) AddAug(location locations.Location) {
	i.instrumentationLock.Lock()
	defer i.instrumentationLock.Unlock()

	i.addAug(location)
}


func (i *InstrumentationService) addAug(location locations.Location) {
	if !strings.HasSuffix(location.GetFileName(), ".go") {
		return
	}

	logger.Logger().Debugf("Attempting to add aug (id=%s) on file %s line %d",
		location.GetAugId(), location.GetFileName(), location.GetLineno())

	if err := location.SetPending(); err != nil {
		logger.Logger().WithError(err).Errorf("Unable to set status of location %s to pending", location.GetAugId())
	}

	if rookErr := i.setBreakpoint(location); rookErr != nil {
		logger.Logger().WithError(rookErr).Errorf("Failed to add aug: %s", location.GetAugId())
		err := location.SetError(rookErr)
		if err != nil {
			logger.Logger().WithError(err).Errorf("Unable to set status of location %s to error", location.GetAugId())
		}
		return
	}

	if err := location.SetActive(); err != nil {
		logger.Logger().WithError(err).Errorf("Unable to set status of location %s to active", location.GetAugId())
	}
}

func (i *InstrumentationService) setBreakpoint(location locations.Location) rookoutErrors.RookoutError {
	filename := location.GetFileName()
	lineno := location.GetLineno()

	addrs, filename, function, err := i.processManager.LineToPC(filename, lineno)
	if err != nil {
		return err
	}

	breakpoint, rookErr := i.processManager.WriteBreakpoint(filename, lineno, function, addrs)
	if rookErr != nil {
		return rookErr
	}

	i.locations.AddLocation(location, breakpoint)
	logger.Logger().Infof("Successfully placed breakpoint on file %s line %d", filename, lineno)

	return nil
}

func (i *InstrumentationService) RemoveAug(augID types.AugId) error {
	i.instrumentationLock.Lock()
	defer i.instrumentationLock.Unlock()

	return i.removeAug(augID)
}


func (i *InstrumentationService) removeAug(augID types.AugId) error {
	logger.Logger().Debugf("Attempting to remove aug %s", augID)
	bp, exists := i.locations.FindBreakpointByAugId(augID)
	if !exists {
		return errors.Errorf("no aug found with id %s", augID)
	}

	i.locations.RemoveLocation(augID)
	shouldClear, err := i.locations.ShouldClearBreakpoint(bp)
	if err != nil {
		return err
	}
	if shouldClear {
		logger.Logger().Debugf("Clearing breakpoint (name=%s) on file %s line %d", bp.Name, bp.File, bp.Line)
		err = i.processManager.EraseBreakpoint(bp)
		if err != nil {
			return err
		}
		i.locations.RemoveBreakpoint(bp)
	}
	logger.Logger().Infof("Successfully removed aug ID %s", augID)
	return nil
}

func (i *InstrumentationService) ReplaceAllRules(newAugs map[types.AugId]locations.Location) error {
	i.instrumentationLock.Lock()
	defer i.instrumentationLock.Unlock()

	var augIDsToRemove []types.AugId
	for _, location := range i.locations.Locations() {
		if _, exists := newAugs[location.GetAug().GetAugId()]; exists {
			
			delete(newAugs, location.GetAug().GetAugId())
		} else {
			augIDsToRemove = append(augIDsToRemove, location.GetAug().GetAugId())
		}
	}

	for _, augID := range augIDsToRemove {
		err := i.removeAug(augID)
		if err != nil {
			logger.Logger().WithError(err).Errorf("Failed to clear aug %s", augID)
			
			continue
		}
	}
	for _, aug := range newAugs {
		i.addAug(aug)
	}

	return nil
}

func (i *InstrumentationService) ClearAugs() {
	i.instrumentationLock.Lock()
	defer i.instrumentationLock.Unlock()

	for _, loc := range i.locations.Locations() {
		if err := i.removeAug(loc.GetAugId()); err != nil {
			logger.Logger().WithError(err).Errorf("Unable to remove aug: %s\n", loc.GetAugId())
		}
	}
}

func (i *InstrumentationService) Stop() {
	if i.staleBreakpointsGC != nil {
		i.staleBreakpointsGC.Stop()
	}

	err := i.processManager.Clean()
	if err != nil {
		logger.Logger().WithError(err).Errorln("Failed to clean process manager")
	}
}
