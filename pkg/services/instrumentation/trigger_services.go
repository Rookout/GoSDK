package instrumentation

import (
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/go-errors/errors"
)

type TriggerServices struct {
	instrumentationService *InstrumentationService
}

func NewTriggerServices() (*TriggerServices, error) {
	if inst, err := NewInstrumentationService(); err == nil {
		return &TriggerServices{instrumentationService: inst}, nil
	} else {
		return nil, err
	}
}

func (t TriggerServices) GetInstrumentation() *InstrumentationService {
	if t.instrumentationService != nil {
		return t.instrumentationService
	}

	
	return nil
}

func (t TriggerServices) RemoveAug(augId types.AugId) error {
	if t.instrumentationService != nil {
		return t.instrumentationService.RemoveAug(augId)
	}

	return errors.Errorf("Couldn't remove aug (%s), instrumentationService is nil", augId)
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
