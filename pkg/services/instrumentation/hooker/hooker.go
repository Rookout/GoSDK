package hooker

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/safe_hook_installer"
	"github.com/Rookout/GoSDK/pkg/types"
	"sort"
	"unsafe"
)

const stackBig = 4096 

type Hooker interface {
	StartWritingBreakpoint(addr uint64, functionEntry uint64, functionEnd uint64) (hookContext BreakpointFlowRunner, err error)
	StartErasingBreakpoint(addr uint64) (hookContext BreakpointFlowRunner, err error)
	GetStackUsageMap() (map[uint64][]map[string]int64, rookoutErrors.RookoutError)
}

type hookerImpl struct {
	bpCallback         unsafe.Pointer
	prologueCallback   unsafe.Pointer
	shouldRunPrologue  unsafe.Pointer
	api                types.NativeHookerAPI
	functions          map[uint64]*hookedFunction 
	breakpoints        map[uint64]*hookedFunction 
	findStackFrameSize func(uint64) int32
}

type hookerManipulator interface {
	addBreakpoint(bpAddress types.Address, functionEntry, functionEnd types.Address)
	removeBreakpoint(bpAddress types.Address, functionEntry types.Address)
	getActiveBreakpointsWithNew(functionEntry, newBreakpoint types.Address) []types.Address
	getActiveBreakpointsWithoutOld(functionEntry, oldBreakpoint types.Address) []types.Address
	getNativeAPI() types.NativeHookerAPI
}

type hookedFunction struct {
	Entry          types.Address
	End            types.Address
	StackFrameSize int32
	Breakpoints    map[types.Address]struct{}
}

func New(bpCallback unsafe.Pointer, prologueCallback unsafe.Pointer, shouldRunPrologue unsafe.Pointer, findStackFrameSize func(uint64) int32) Hooker {
	return &hookerImpl{
		bpCallback:         bpCallback,
		prologueCallback:   prologueCallback,
		shouldRunPrologue:  shouldRunPrologue,
		api:                NewNativeAPI(),
		functions:          map[uint64]*hookedFunction{},
		breakpoints:        map[uint64]*hookedFunction{},
		findStackFrameSize: findStackFrameSize,
	}
}

func sortSlice(slice []uint64) []uint64 {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})

	return slice
}

func (h *hookerImpl) addBreakpoint(bpAddress types.Address, functionEntry, functionEnd types.Address) {
	if !h.hasFunction(functionEntry) {
		h.functions[functionEntry] = &hookedFunction{Entry: functionEntry, End: functionEnd, StackFrameSize: h.findStackFrameSize(bpAddress), Breakpoints: map[types.Address]struct{}{}}
	}

	h.functions[functionEntry].Breakpoints[bpAddress] = struct{}{}
	h.breakpoints[bpAddress] = h.functions[functionEntry]
}

func (h *hookerImpl) removeBreakpoint(bpAddress types.Address, functionEntry types.Address) {
	delete(h.functions[functionEntry].Breakpoints, bpAddress)
	delete(h.breakpoints, bpAddress)
}

func (h *hookerImpl) StartWritingBreakpoint(bpAddress types.Address, functionEntry types.Address, functionEnd types.Address) (BreakpointFlowRunner, error) {
	if h.hasBreakpoint(functionEntry, bpAddress) {
		return nil, fmt.Errorf("A breakpoint with this Address (%d) already exists\n", bpAddress)
	}
	initInfo := h.getHookingContextInitInfo(functionEntry, functionEnd, bpAddress)
	return startAddingBreakpoint(h, initInfo, safe_hook_installer.NewSafeHookInstaller)
}

func (h *hookerImpl) StartErasingBreakpoint(bpAddress types.Address) (BreakpointFlowRunner, error) {
	functionEntry, functionEnd, exists := h.getFunctionAddressesByBreakpoint(bpAddress)
	if !exists {
		return nil, fmt.Errorf("No breakpoint exists in the Address %d\n", bpAddress)
	}
	initInfo := h.getHookingContextInitInfo(functionEntry, functionEnd, bpAddress)
	return startRemovingBreakpoint(h, initInfo, safe_hook_installer.NewSafeHookInstaller)
}

func (h *hookerImpl) GetStackUsageMap() (map[uint64][]map[string]int64, rookoutErrors.RookoutError) {
	return h.api.GetStackUsageMap()
}

func (h *hookerImpl) getFunctionAddressesByBreakpoint(bpAddress types.Address) (uint64, uint64, bool) {
	if function, ok := h.breakpoints[bpAddress]; ok {
		return function.Entry, function.End, true
	}

	return 0, 0, false
}

func (h *hookerImpl) getNativeAPI() types.NativeHookerAPI {
	return h.api
}

func (h *hookerImpl) getActiveBreakpointsWithNew(functionEntry, newBreakpoint types.Address) []types.Address {
	bps := []types.Address{newBreakpoint}
	if !h.hasFunction(functionEntry) {
		return bps
	}
	for addr := range h.functions[functionEntry].Breakpoints {
		bps = append(bps, addr)
	}
	return sortSlice(bps)
}

func (h *hookerImpl) getActiveBreakpointsWithoutOld(functionEntry, oldBreakpoint types.Address) []types.Address {
	var bps []types.Address
	for addr := range h.functions[functionEntry].Breakpoints {
		if addr != oldBreakpoint {
			bps = append(bps, addr)
		}
	}
	return sortSlice(bps)
}

func (h *hookerImpl) hasFunction(functionEntry types.Address) bool {
	_, ok := h.functions[functionEntry]
	return ok

}

func (h *hookerImpl) hasBreakpoint(functionEntry, bpAddress types.Address) bool {
	if !h.hasFunction(functionEntry) {
		return false
	}
	if _, ok := h.functions[functionEntry].Breakpoints[bpAddress]; !ok {
		return false
	}
	return true
}

func (h *hookerImpl) getStackFrameSize(functionEntry types.Address) int32 {
	if h.hasFunction(functionEntry) {
		return h.functions[functionEntry].StackFrameSize
	} else {
		return h.findStackFrameSize(functionEntry)
	}
}

func (h *hookerImpl) getHookingContextInitInfo(functionEntry, functionEnd, bpAddress types.Address) BreakpointFlowRunnerInitializationInfo {
	initInfo := BreakpointFlowRunnerInitializationInfo{
		functionEntry:     functionEntry,
		functionEnd:       functionEnd,
		breakpointAddress: bpAddress,
		bpCallback:        uintptr(h.bpCallback),
	}

	if functionStackFrameSize := h.getStackFrameSize(functionEntry); functionStackFrameSize < stackBig {
		initInfo.prologueCallback = uintptr(h.prologueCallback)
		initInfo.shouldRunPrologueCallback = uintptr(h.shouldRunPrologue)
		initInfo.functionStackFrameSize = functionStackFrameSize
	} else {
		initInfo.prologueCallback = uintptr(unsafe.Pointer(nil))
		initInfo.shouldRunPrologueCallback = uintptr(unsafe.Pointer(nil))
		initInfo.functionStackFrameSize = -1
	}

	return initInfo
}
