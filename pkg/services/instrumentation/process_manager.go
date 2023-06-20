package instrumentation

import (
	"context"
	"os"
	"reflect"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/locations_set"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/callback"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker/prologue"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
	"github.com/Rookout/GoSDK/pkg/utils"
)

const maxStackFrames = 128 

type ProcessManager struct {
	hooker            hooker.Hooker
	binaryInfo        *binary_info.BinaryInfo
	locations         *locations_set.LocationsSet
	trampolineManager *trampolineManager
	cancelGCControl   context.CancelFunc
	cancelBPMonitor   context.CancelFunc
}

func NewProcessManager(locations *locations_set.LocationsSet, breakpointMonitorInterval time.Duration) (pm *ProcessManager, err rookoutErrors.RookoutError) {
	bi, err := createBinaryInfo()
	if err != nil {
		return nil, err
	}

	h, err := createHooker()
	defer func() {
		if err != nil {
			_ = hooker.Destroy()
		}
	}()
	if err != nil {
		return nil, err
	}

	module.Init()
	p := &ProcessManager{hooker: h, binaryInfo: bi, locations: locations, trampolineManager: newTrampolineManager()}

	ctxGCControl, cancelGCControl := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelGCControl()
		}
	}()
	p.cancelGCControl = cancelGCControl
	triggerChan := make(chan bool, 10000)
	gcController := newGCController()
	utils.CreateGoroutine(func() {
		gcController.start(ctxGCControl, triggerChan)
	})
	callback.SetTriggerChan(triggerChan)

	ctxBPMonitor, cancelBPMonitor := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelBPMonitor()
		}
	}()
	p.cancelBPMonitor = cancelBPMonitor
	utils.CreateGoroutine(func() {
		p.monitorBreakpoints(ctxBPMonitor, breakpointMonitorInterval)
	})

	err = prologue.Init(p.binaryInfo)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *ProcessManager) monitorBreakpoints(ctx context.Context, breakpointMonitorInterval time.Duration) {
	for {
		monitorTimeout := time.NewTimer(breakpointMonitorInterval)
		select {
		case <-ctx.Done():
			return

		case <-monitorTimeout.C:
			bpInstances := p.locations.GetBreakpointInstances()
			for _, bpInstance := range bpInstances {
				failedCounter := atomic.SwapUint64(bpInstance.FailedCounter, 0)
				if failedCounter > 0 {
					locations, ok := p.locations.FindLocationsByBreakpointName(bpInstance.Breakpoint.Name)
					if !ok {
						continue
					}

					for i := range locations {
						locations[i].SetError(rookoutErrors.NewFailedToExecuteBreakpoint(failedCounter))
					}
				}
			}
		}
	}
}

func createHooker() (hooker.Hooker, rookoutErrors.RookoutError) {
	h := hooker.New(unsafe.Pointer(reflect.ValueOf(callback.Callback).Pointer()))

	err := hooker.Init(funcForInit)
	if err != nil {
		logger.Logger().WithError(err).Errorf("Unable to start hooker")
		return nil, err
	}

	return h, nil
}

func createBinaryInfo() (*binary_info.BinaryInfo, rookoutErrors.RookoutError) {
	bi := binary_info.NewBinaryInfo()
	exec, err := os.Executable()
	if err != nil {
		return nil, rookoutErrors.NewFailedToGetExecutable(err)
	}
	err = bi.LoadBinaryInfo(exec, binary_info.GetEntrypoint(exec), nil)
	if err != nil {
		return nil, rookoutErrors.NewFailedToLoadBinaryInfo(err)
	}
	bi.Dwarf = bi.Images[0].Dwarf 

	callback.SetBinaryInfo(bi)
	return bi, nil
}

func (p *ProcessManager) WriteBreakpoint(filename string, lineno int) (*augs.Breakpoint, rookoutErrors.RookoutError) {
	addrs, _, _, err := p.binaryInfo.FindFileLocation(filename, lineno)
	if err != nil {
		return nil, err
	}
	if bp, ok := p.locations.FindBreakpointByAddrs(addrs); ok {
		return bp, nil
	}

	return writeBreakpoint(filename, lineno, p.getFunction, addrs, p.locations.BreakpointStorage, p.binaryInfo, p.hooker)
}

func (p *ProcessManager) getFunction(addr uint64) (*binary_info.Function, *augs.Function, rookoutErrors.RookoutError) {
	filename, lineno, biFunction := p.binaryInfo.PCToLine(addr)
	if function, ok := p.locations.FindFunctionByEntry(biFunction.Entry); ok {
		return biFunction, function, nil
	}

	function, err := NewFunction(biFunction, filename, lineno, p.binaryInfo, p.hooker, p.trampolineManager)
	if err != nil {
		return nil, nil, err
	}
	p.locations.AddFunction(function)
	return biFunction, function, nil
}

func (p *ProcessManager) EraseBreakpoint(bp *augs.Breakpoint) rookoutErrors.RookoutError {
	return eraseBreakpoint(bp, p.hooker, p.locations.BreakpointStorage)
}


func (_ *ProcessManager) Clean() rookoutErrors.RookoutError {
	err := hooker.Destroy()
	if err != nil {
		return rookoutErrors.NewFailedToDestroyNative(err)
	}
	return nil
}
