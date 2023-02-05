//go:build arm64 && darwin
// +build arm64,darwin

package binary_info

import "unsafe"



/* #cgo CFLAGS:
#include <mach-o/getsect.h>
#include <stdlib.h>
#include <mach-o/dyld.h>
#include <string.h>

bool str_ends_with(const char* str, const char* pattern, int pattern_size){
	int str_size = strlen(str);
	int delta = str_size-pattern_size;
	if (delta<0){
		return false;
	}
	return strcmp(str+delta,pattern) == 0;
}

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
		// Iterate over all loaded images and check if there is a match to the imagePath
		if (strcmp(_dyld_get_image_name(i), imagePath) == 0) {
			return _dyld_get_image_vmaddr_slide(i);
		}
	}

	// If we didn't find a match, it's possible that the image path we search is a suffix of one of the loaded images, so it's best effort to use that.
	// I saw that it happens when running the tests with our "make test". In this case the loaded image path is "/private/<imagePath>"
	// instead of <imagePath>
	int imagePathLen = strlen(imagePath);
	for (uint32_t i = 0; i < _dyld_image_count(); i++) {
		if (str_ends_with(_dyld_get_image_name(i), imagePath, imagePathLen)) {
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
