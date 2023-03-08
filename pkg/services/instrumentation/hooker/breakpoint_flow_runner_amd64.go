package hooker

import (
	"encoding/binary"
	"math"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

const jmp = "\xe9"

func (c *breakpointFlowRunner) ApplyBreakpointsState() (err error) {
	
	if !c.IsDefaultState() {
		c.jumpDestination, err = c.nativeAPI.GetStateEntryAddr(c.function.Entry, c.function.End, c.stateID)
		if err != nil {
			return err
		}
	}
	return c.installHook()
}

func (c *breakpointFlowRunner) buildJMP(hookAddr uintptr) ([]byte, rookoutErrors.RookoutError) {
	relativeAddr := int64(c.jumpDestination - (hookAddr + uintptr(len(jmp)) + 4))
	if relativeAddr > math.MaxInt32 || relativeAddr < math.MinInt32 {
		return nil, rookoutErrors.NewBranchDestTooFar(hookAddr, c.jumpDestination, c.stateID)
	}

	jmpBytes := make([]byte, len(jmp)+4)
	copy(jmpBytes, []byte(jmp))
	binary.LittleEndian.PutUint32(jmpBytes[len(jmp):], uint32(relativeAddr))
	return jmpBytes, nil
}
