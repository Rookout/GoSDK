//go:build arm64 && darwin
// +build arm64,darwin

package safe_hook_installer

import (
	"syscall"
	_ "unsafe"
)


import "C"


func setWritable(address uintptr, size int) int {
	
	return int(C.native_mprotect(C.ulonglong(address), C.ulonglong(size), C.int(syscall.PROT_READ|syscall.PROT_WRITE)))
}


func setExecutable(address uintptr, size int) int {
	
	return int(C.native_mprotect(C.ulonglong(address), C.ulonglong(size), C.int(syscall.PROT_READ|syscall.PROT_EXEC)))
}
