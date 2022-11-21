//go:build arm64 && darwin
// +build arm64,darwin

package binary_info

import "unsafe"




import "C"




func GetEntrypoint(imagePath string) uint64 {
	cPath := C.CString(imagePath)
	defer C.free(unsafe.Pointer(cPath))
	return uint64(C.DynamicBaseAddress(cPath))
}
