//go:build windows
// +build windows

package safe_hook_installer

import "github.com/Rookout/GoSDK/pkg/rookoutErrors"

func (h *HookWriter) AddWritePermission() rookoutErrors.RookoutError {
	return rookoutErrors.NewUnsupportedPlatform()
}

func (h *HookWriter) RestorePermissions() rookoutErrors.RookoutError {
	return rookoutErrors.NewUnsupportedPlatform()
}

func (h *HookWriter) write() int {
	return 0
}
