//go:build arm64 && darwin
// +build arm64,darwin

package safe_hook_installer

import (
	_ "unsafe"
)

/* #cgo CFLAGS:
#include <mach/mach.h>
#include <sys/mman.h>
#include <libkern/OSCacheControl.h>
#include <unistd.h>

void  __attribute__ ((noinline)) start_marker_function(){
    asm(".align 14");
}

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

int write_bytes(unsigned long long dest, unsigned long long src, int length){
    const unsigned long long page_size = getpagesize();
    const unsigned long long page_mask = ~(page_size-1);
    unsigned long long start_page = page_mask&dest;
    unsigned long long end_page = (dest+length+page_size)&page_mask;
    unsigned long long mprotect_size = end_page-start_page;
    int mprotect_res = native_mprotect(start_page, mprotect_size, PROT_READ|PROT_WRITE);
    if(mprotect_res){
        return mprotect_res;
    }
    for(int i=0; i<length; i++){
        *(unsigned char*)(dest+i) = *(unsigned char*)(src+i);
    }
    return native_mprotect(start_page, mprotect_size, PROT_READ|PROT_EXEC);
}

void  __attribute__ ((noinline)) end_marker_function(){
	asm(".align 14");
}

*/
import "C"

func writeBytes(dest, src uintptr, length int) int {
	return int(C.write_bytes(C.ulonglong(dest), C.ulonglong(src), C.int(length)))
}
