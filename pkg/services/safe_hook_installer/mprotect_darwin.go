//go:build darwin
// +build darwin

package safe_hook_installer

import (
	"fmt"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

/* #cgo CFLAGS:
#include <mach/mach.h>
#include <sys/mman.h>
#include <libkern/OSCacheControl.h>
#include <unistd.h>

#ifdef __aarch64__
void  __attribute__ ((noinline)) start_marker_function(){
    asm(".align 14");
}
#endif

int get_current_memory_protection(uint64_t addr, uint64_t size) {
    vm_address_t vm_addr = (vm_address_t) addr;
    vm_size_t vm_size = (vm_size_t) size;
    vm_region_basic_info_data_64_t info;
    mach_msg_type_number_t info_count = VM_REGION_BASIC_INFO_COUNT_64;
    memory_object_name_t object;

    kern_return_t kret = vm_region_64(mach_task_self(), &vm_addr, &vm_size, VM_REGION_BASIC_INFO_64,
                                        (vm_region_info_t) &info, &info_count, &object);

	if (kret != KERN_SUCCESS) {
		return -1;
	}

    return info.protection;
}

int mprotect_darwin(uint64_t addr, uint64_t size, int protection) {
    if (mprotect((void *)addr, (size_t)size, protection) != 0) {
        // When a caller finds that they cannot obtain write permission on a mapped entry, the following VM_PROT_COPY flag
        // can be used.
        // The entry will be made "needs copy" effectively copying the object (using COW), and write permission will be added
        // to the maximum protections for the associated entry.
#ifdef __aarch64__
        if ((protection & PROT_EXEC) == 0) {
            protection = protection | VM_PROT_COPY;
        }
#else
        protection = protection | VM_PROT_COPY;
#endif
        kern_return_t kret = vm_protect(mach_task_self(), addr, size, 0,
                                        protection);
        if (kret != KERN_SUCCESS) {
			return kret;
        }
    }

    // * https://community.arm.com/arm-community-blogs/b/architectures-and-processors-blog/posts/caches-and-self-modifying-code
    // * https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man3/sys_icache_invalidate.3.html
    // * In arm the instruction cache isn't invalidated when modifying the code. This could cause callers to see the old
    // * value before making any modification (or even worse, the old permissions).
    // * Here we use apple's functions for invalidating the cache. First we flush the dcache, and then we invalidate
    // * the icache. This ensures that any modifications to cache lines (such as modified data and permissions modification).
    // * We probably don't need to flush the dcache since invalidating the icache should be enough, but we do it
    // * for completeness and safety.
#ifdef __aarch64__
    sys_dcache_flush((void *) addr, size);
    sys_icache_invalidate((void *) addr, size);
#endif

	return 0;
}

#ifdef __aarch64__
// Parameters:
//     dest - Address to write bytes to
//     src - Address to take bytes to write from
//     size - Number of bytes to write
//     dest_start_page - Page aligned address of start of dest memory area
//     dest_end_page - Page aligned address of end of dest memory area
// Returns 0 on success, and on error:
// -1: Failed to get current memory protection
// -2: Failed to add write permission to memory
// -3: Failed to return to previous memory permissions
int write_bytes(unsigned long long dest, unsigned long long src, unsigned long long size, unsigned long long dest_start_page, unsigned long long dest_end_page) {
	int current_memory_protection = get_current_memory_protection(dest_start_page, size);
	if (-1 == current_memory_protection) {
		return -1;
	}

    int mprotect_res = mprotect_darwin(dest_start_page, dest_end_page-dest_start_page, PROT_READ|PROT_WRITE);
    if(0 != mprotect_res){
        return -2;
    }

    for(int i=0; i<size; i++){
        *(unsigned char*)(dest+i) = *(unsigned char*)(src+i);
    }

    mprotect_res = mprotect_darwin(dest_start_page, dest_end_page-dest_start_page, current_memory_protection);
	if (0 != mprotect_res) {
		return -3;
	}

	return 0;
}
#else
int write_bytes(unsigned long long dest, unsigned long long src, unsigned long long size, unsigned long long dest_start_page, unsigned long long dest_end_page) {
	return -4;
}
#endif

#ifdef __aarch64__
void  __attribute__ ((noinline)) end_marker_function(){
	asm(".align 14");
}
#endif
*/
import "C"

func getCurrentMemoryProtection(addr uint64, size uint64) (int, rookoutErrors.RookoutError) {
	prot := int(C.get_current_memory_protection(C.ulonglong(addr), C.ulonglong(size)))
	if prot == -1 {
		return 0, rookoutErrors.NewFailedToGetCurrentMemoryProtection(addr, size)
	}
	return prot, nil
}

func (h *HookWriter) changeHookPermissions(prot int) rookoutErrors.RookoutError {
	ret := C.mprotect_darwin(C.ulonglong(h.hookPageAlignedStart), C.ulonglong(h.hookPageAlignedEnd-h.hookPageAlignedStart), C.int(prot))
	if ret != 0 {
		rookoutErrors.NewMprotectFailed(h.hookPageAlignedStart, int(h.hookPageAlignedEnd-h.hookPageAlignedStart), prot, fmt.Sprintf("mprotect failed: %d", ret))
	}
	return nil
}

func (h *HookWriter) write_darwin_arm64() int {
	
	return int(C.write_bytes(C.ulonglong(h.HookAddr), C.ulonglong(uintptr(unsafe.Pointer(&(h.Hook[0])))), C.ulonglong(len(h.Hook)), C.ulonglong(h.hookPageAlignedStart), C.ulonglong(h.hookPageAlignedEnd)))
}
