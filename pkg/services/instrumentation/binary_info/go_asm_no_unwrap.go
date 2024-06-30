//go:build (amd64 && go1.16 && !go1.17) || (arm64 && go1.16 && !go1.18)
// +build amd64,go1.16,!go1.17 arm64,go1.16,!go1.18

package binary_info

import (
	"reflect"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)


func (b *BinaryInfo) GetUnwrappedFuncPointer(f func()) (uintptr, rookoutErrors.RookoutError) {
	return reflect.ValueOf(f).Pointer(), nil
}
