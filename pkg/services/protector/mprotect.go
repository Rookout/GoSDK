package protector

import "syscall"


func GetPageStart(addr uintptr) uintptr {
	return addr & (^uintptr(syscall.Getpagesize() - 1))
}


func GetPageEnd(addr uintptr) uintptr {
	return GetPageStart(addr) + uintptr(syscall.Getpagesize())
}
