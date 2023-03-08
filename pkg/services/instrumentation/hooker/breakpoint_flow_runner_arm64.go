package hooker

import (
	"encoding/binary"
	"sync/atomic"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

const b = 0x14000000

func (c *breakpointFlowRunner) ApplyBreakpointsState() error {
	
	if c.IsDefaultState() {
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

func abs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

func (c *breakpointFlowRunner) buildJMP(hookAddr uintptr) ([]byte, rookoutErrors.RookoutError) {
	
	offset := int64(c.jumpDestination-hookAddr) / int64(4)

	
	
	
	if offset%4 != 0 {
		return nil, rookoutErrors.NewInvalidBranchDest(hookAddr, c.jumpDestination, c.stateID)
	} else if abs(offset)&0b1111111111111111111111111 != abs(offset) {
		return nil, rookoutErrors.NewBranchDestTooFar(hookAddr, c.jumpDestination, c.stateID)
	}

	
	encodedOffset := uint32(offset) & 0b11111111111111111111111111
	if offset < 0 {
		encodedOffset |= 0x04000000
	}

	
	encodedInst := uint32(b) | encodedOffset

	
	encodedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(encodedBytes, encodedInst)

	return encodedBytes, nil
}
