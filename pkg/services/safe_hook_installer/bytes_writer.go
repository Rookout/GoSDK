//go:build !arm64 || !darwin
// +build !arm64 !darwin

package safe_hook_installer

import "unsafe"

func writeBytes(dest, src uintptr, length int) int {
	for i := 0; i < length; i++ {
		*(*uint8)(unsafe.Pointer(dest + uintptr(i))) = *(*uint8)(unsafe.Pointer(src + uintptr(i)))
	}
	return 0
}
