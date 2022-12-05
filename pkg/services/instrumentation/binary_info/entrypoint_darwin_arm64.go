//go:build arm64 && darwin
// +build arm64,darwin

package binary_info

import "unsafe"



/* #cgo CFLAGS:
#include <mach-o/getsect.h>
#include <stdlib.h>
#include <mach-o/dyld.h>
#include <string.h>

uint64_t StaticBaseAddress(void) {
	uint64_t addr = 0;
	const struct segment_command_64* command = getsegbyname("__TEXT");
	if (command) {
		addr = command->vmaddr;
	}
	return addr;
}

intptr_t ImageSlide(char* imagePath) {
	for (uint32_t i = 0; i < _dyld_image_count(); i++) {
		if (strcmp(_dyld_get_image_name(i), imagePath) == 0) {
			return _dyld_get_image_vmaddr_slide(i);
		}
	}
	return 0;
}

uint64_t DynamicBaseAddress(char* imagePath) {
	return StaticBaseAddress() + ImageSlide(imagePath);
}
*/
import "C"




func GetEntrypoint(imagePath string) uint64 {
	cPath := C.CString(imagePath)
	defer C.free(unsafe.Pointer(cPath))
	return uint64(C.DynamicBaseAddress(cPath))
}
