//go:build amd64
// +build amd64

package instrumentation

import (
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type trampolineManager struct {
}

func newTrampolineManager() *trampolineManager {
	return &trampolineManager{}
}

func (t *trampolineManager) getTrampolineAddress() (*uint64, unsafe.Pointer, rookoutErrors.RookoutError) {
	return nil, nil, nil
}
