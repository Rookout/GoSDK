package instrumentation

import (
	"time"

	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/go-errors/errors"
)

type TriggerServices struct {
	instrumentationService *InstrumentationService
}

const breakpointMonitorInterval = 10 * time.Second

func NewTriggerServices() (*TriggerServices, error) {
	inst, err := NewInstrumentationService(breakpointMonitorInterval)
	if err != nil {
		return nil, err
	}

	return &TriggerServices{instrumentationService: inst}, nil
}

func (t TriggerServices) GetInstrumentation() *InstrumentationService {
	if t.instrumentationService != nil {
		return t.instrumentationService
	}

	
	return nil
}

func (t TriggerServices) RemoveAug(augID types.AugID) error {
	if t.instrumentationService != nil {
		return t.instrumentationService.RemoveAug(augID)
	}

	return errors.Errorf("Couldn't remove aug (%s), instrumentationService is nil", augID)
}

func (t TriggerServices) ClearAugs() {
	if t.instrumentationService != nil {
		t.instrumentationService.ClearAugs()
	}
}

func (t TriggerServices) Close() {
	t.ClearAugs()

	if t.instrumentationService != nil {
		t.instrumentationService.Stop()
	}
}
