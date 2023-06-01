//go:build arm64 && darwin
// +build arm64,darwin

package protector

/* #cgo CFLAGS:
#include <mach/mach.h>
#include <sys/mman.h>
#include <libkern/OSCacheControl.h>
#include <unistd.h>

void  __attribute__ ((noinline)) align_writer_darwin_arm64_start() {
    asm(".align 14");
}

int get_current_memory_protection(uint64_t addr, uint64_t size);

int mprotect_darwin(uint64_t addr, uint64_t size, int protection);

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

void  __attribute__ ((noinline)) align_writer_darwin_arm64_end() {
	asm(".align 14");
}
*/
import "C"
import "unsafe"


func Write(addr uintptr, bytes []byte, pageStart uintptr, pageEnd uintptr) int {
	return int(C.write_bytes(C.ulonglong(addr), C.ulonglong(uintptr(unsafe.Pointer(&(bytes[0])))), C.ulonglong(len(bytes)), C.ulonglong(pageStart), C.ulonglong(pageEnd)))
}
