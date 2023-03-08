package safe_hook_installer

import "github.com/Rookout/GoSDK/pkg/services/safe_hook_validator"


type hookManager struct {
	
	getFunctionType func(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error)
	hookAddr        uintptr
	hook            []byte
}

func NewHookManager(hookAddr uintptr, hook []byte, getFunctionType func(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error)) *hookManager {
	return &hookManager{
		hookAddr:        hookAddr,
		hook:            hook,
		getFunctionType: getFunctionType,
	}
}

func (h *hookManager) GetFunctionType(functionEntry uint64, functionEnd uint64) (safe_hook_validator.FunctionType, error) {
	return h.getFunctionType(functionEntry, functionEnd)
}

func (h *hookManager) GetDangerZoneStartAddress() uint64 {
	
	return uint64(h.hookAddr + 1)
}

func (h *hookManager) GetDangerZoneEndAddress() uint64 {
	return uint64(h.hookAddr + uintptr(len(h.hook)))
}
