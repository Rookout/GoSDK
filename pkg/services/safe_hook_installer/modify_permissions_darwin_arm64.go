//go:build arm64 && darwin
// +build arm64,darwin

package safe_hook_installer

import (
	"syscall"
	_ "unsafe"
)

/* #cgo CFLAGS:
#include <mach/mach.h>
#include <sys/mman.h>
#include <stdio.h>
#include <libkern/OSCacheControl.h>

int native_mprotect(unsigned long long addr, unsigned long long size, int protection) {
    if (mprotect((void*)addr, size, protection) != 0) {
        // When a caller finds that he cannot obtain write permission on a mapped entry, the following VM_PROT_COPY flag
        // can be used.
        // The entry will be made "needs copy" effectively copying the object (using COW), and write permission will be added
        // to the maximum protections for the associated entry.
        if ((protection & PROT_EXEC) == 0) {
            protection = protection | VM_PROT_COPY;
        }
        kern_return_t kret = vm_protect(mach_task_self(), addr, size, 0, protection);
        if (kret != KERN_SUCCESS) {
			return kret;
        }
    }
	sys_dcache_flush((void *) addr, size);
    sys_icache_invalidate((void *) addr, size);
	return 0;
}
*/
import "C"


func setWritable(address uintptr, size int) int {
	
	return int(C.native_mprotect(C.ulonglong(address), C.ulonglong(size), C.int(syscall.PROT_READ|syscall.PROT_WRITE)))
}


func setExecutable(address uintptr, size int) int {
	
	return int(C.native_mprotect(C.ulonglong(address), C.ulonglong(size), C.int(syscall.PROT_READ|syscall.PROT_EXEC)))
}
