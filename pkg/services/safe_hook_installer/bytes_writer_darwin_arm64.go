//go:build darwin && arm64
// +build darwin,arm64

package safe_hook_installer

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/protector"
)


func (h *HookWriter) AddWritePermission() rookoutErrors.RookoutError {
	return nil
}


func (h *HookWriter) RestorePermissions() rookoutErrors.RookoutError {
	return nil
}

func (h *HookWriter) write() int {
	
	return protector.Write(h.HookAddr, h.Hook, h.hookPageAlignedStart, h.hookPageAlignedEnd)
}
