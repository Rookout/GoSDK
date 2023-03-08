//go:build arm64
// +build arm64

package instrumentation

import (
	"sync"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type trampolineManager struct {
	trampolineAddressesInUse     map[int]bool
	trampolineAddressesInUseLock sync.Mutex
}

func newTrampolineManager() *trampolineManager {
	t := &trampolineManager{}
	t.trampolineAddressesInUse = make(map[int]bool, trampolineCount)
	for i := range finalTrampolineAddresses {
		t.trampolineAddressesInUse[i] = false
	}
	return t
}

func (t *trampolineManager) getTrampolineAddress() (*uint64, unsafe.Pointer, rookoutErrors.RookoutError) {
	t.trampolineAddressesInUseLock.Lock()
	defer t.trampolineAddressesInUseLock.Unlock()

	for i, inUse := range t.trampolineAddressesInUse {
		if inUse {
			continue
		}
		t.trampolineAddressesInUse[i] = true
		return &(finalTrampolineAddresses[i]), getMiddleTrampolineAddress(i), nil
	}

	return nil, nil, rookoutErrors.NewAllTrampolineAddressesInUse()
}
