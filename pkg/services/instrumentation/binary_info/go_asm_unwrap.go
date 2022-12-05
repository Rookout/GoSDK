//go:build amd64 && go1.17
// +build amd64,go1.17

package binary_info

import (
	"reflect"
	"runtime"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)



func (b *BinaryInfo) GetUnwrappedFuncPointer(f func()) (uintptr, rookoutErrors.RookoutError) {
	wrappedFuncPointer := reflect.ValueOf(f).Pointer()
	funcName := runtime.FuncForPC(wrappedFuncPointer).Name()

	for i := range b.Functions {
		if b.Functions[i].Name == funcName {
			
			if wrappedFuncPointer != uintptr(b.Functions[i].Entry) {
				logger.Logger().Infof("Found unwrapped func address for %s: 0x%x (replaces 0x%x)", funcName, b.Functions[i].Entry, wrappedFuncPointer)
				return uintptr(b.Functions[i].Entry), nil
			}
		}
	}

	logger.Logger().Infof("No unwrapped function found for: %s", funcName)
	return 0, rookoutErrors.NewUnwrappedFuncNotFound(funcName)
}
