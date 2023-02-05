//go:build !go1.15 || go1.20
// +build !go1.15 go1.20

package module

import "unsafe"

func FindFuncMaxSPDelta(_ uint64) int32 {
	panic("Unsupported go version!")
}

func PatchModuleData(function interface{}, rawAddressMapping unsafe.Pointer, pcspNativeInfo interface{}, stateId int) error {
	panic("Unsupported go version!")
}
