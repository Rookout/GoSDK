package hooker

import (
	"sync/atomic"
)

func (c *breakpointFlowRunner) ApplyBreakpointsState() error {
	
	if c.IsUnhookedState() {
		return c.installHook()
	}

	trampoline, err := c.nativeAPI.GetStateEntryAddr(c.function.Entry, c.function.End, c.stateID)
	if err != nil {
		return err
	}
	atomic.StoreUint64(c.function.FinalTrampolinePointer, uint64(trampoline))

	if !c.function.Hooked {
		c.jumpDestination = uintptr(c.function.MiddleTrampolineAddress)
		return c.installHook()
	}

	return nil
}
