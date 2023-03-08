package locations_set

import (
	"sync"

	"github.com/Rookout/GoSDK/pkg/augs"
)

type BreakpointStorage struct {
	functions   map[*augs.Function][]*augs.BreakpointInstance 
	storageLock *sync.Mutex
}

func newBreakpointStorage() *BreakpointStorage {
	return &BreakpointStorage{
		functions:   make(map[*augs.Function][]*augs.BreakpointInstance),
		storageLock: &sync.Mutex{},
	}
}

func (b *BreakpointStorage) GetBreakpointInstances() []*augs.BreakpointInstance {
	b.storageLock.Lock()
	defer b.storageLock.Unlock()

	var bpInstances []*augs.BreakpointInstance
	for _, instances := range b.functions {
		bpInstances = append(bpInstances, instances...)
	}

	return bpInstances
}

func (b *BreakpointStorage) AddBreakpointInstance(bpInstance *augs.BreakpointInstance) {
	b.storageLock.Lock()
	defer b.storageLock.Unlock()

	var bpInstances []*augs.BreakpointInstance
	if prevInstances, ok := b.functions[bpInstance.Function]; ok {
		bpInstances = prevInstances
	} else {
		b.addFunction(bpInstance.Function)
	}

	bpInstances = append(bpInstances, bpInstance)
	b.functions[bpInstance.Function] = bpInstances
}

func (b *BreakpointStorage) AddFunction(function *augs.Function) {
	b.storageLock.Lock()
	defer b.storageLock.Unlock()

	b.addFunction(function)
}

func (b *BreakpointStorage) addFunction(function *augs.Function) {
	function.GetBreakpointInstances = func() []*augs.BreakpointInstance {
		return b.FindBreakpointsByFunctionEntry(function.Entry)
	}
	b.functions[function] = nil
}

func (b *BreakpointStorage) RemoveBreakpointInstance(bpInstance *augs.BreakpointInstance) {
	b.storageLock.Lock()
	defer b.storageLock.Unlock()

	if _, ok := b.functions[bpInstance.Function]; !ok {
		return
	}

	var bpInstances []*augs.BreakpointInstance
	for _, instance := range b.functions[bpInstance.Function] {
		if instance.Addr == bpInstance.Addr {
			continue
		}
		bpInstances = append(bpInstances, instance)
	}
	b.functions[bpInstance.Function] = bpInstances
}

func (b *BreakpointStorage) FindBreakpointsByFunctionEntry(entry uint64) []*augs.BreakpointInstance {
	b.storageLock.Lock()
	defer b.storageLock.Unlock()

	for f, bps := range b.functions {
		if f.Entry == entry {
			return bps
		}
	}

	return nil
}

func (b *BreakpointStorage) FindFunctionByEntry(entry uint64) (function *augs.Function, exists bool) {
	b.storageLock.Lock()
	defer b.storageLock.Unlock()

	for f := range b.functions {
		if f.Entry == entry {
			return f, true
		}
	}
	return nil, false
}

func (b *BreakpointStorage) FindBreakpointByAddr(addr uint64) (breakpoint *augs.BreakpointInstance, exists bool) {
	for _, bps := range b.functions {
		for _, bp := range bps {
			if bp.Addr == addr {
				return bp, true
			}
		}
	}

	return nil, false
}
