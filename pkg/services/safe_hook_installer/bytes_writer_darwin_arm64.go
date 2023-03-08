//go:build darwin && arm64
// +build darwin,arm64

package safe_hook_installer

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)


func (h *HookWriter) AddWritePermission() rookoutErrors.RookoutError {
	return nil
}


func (h *HookWriter) RestorePermissions() rookoutErrors.RookoutError {
	return nil
}

func (h *HookWriter) write() int {
	return h.write_darwin_arm64()
}
