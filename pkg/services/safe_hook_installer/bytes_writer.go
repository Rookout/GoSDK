//go:build !(darwin && arm64) && !windows
// +build !darwin !arm64
// +build !windows

package safe_hook_installer

import (
	"syscall"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/protector"
)

func (h *HookWriter) AddWritePermission() rookoutErrors.RookoutError {
	currentMemoryProtection, err := protector.GetMemoryProtection(uint64(h.hookPageAlignedStart), uint64(h.hookPageAlignedEnd-h.hookPageAlignedStart))
	if err != nil {
		return err
	}
	h.originalMemoryProtection = currentMemoryProtection

	return protector.ChangeMemoryProtection(h.hookPageAlignedStart, h.hookPageAlignedEnd, currentMemoryProtection|syscall.PROT_WRITE)
}

func (h *HookWriter) RestorePermissions() rookoutErrors.RookoutError {
	return protector.ChangeMemoryProtection(h.hookPageAlignedStart, h.hookPageAlignedEnd, h.originalMemoryProtection)
}

func (h *HookWriter) write() int {
	for i, b := range h.Hook {
		*(*uint8)(unsafe.Pointer(h.HookAddr + uintptr(i))) = b
	}
	return 0
}
