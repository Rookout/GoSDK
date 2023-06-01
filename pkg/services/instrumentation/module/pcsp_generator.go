package module

import (
	"github.com/Rookout/GoSDK/pkg/services/disassembler"
)

func pcspFromInstructions(instructions []*disassembler.Instruction, lastOffset uintptr) []PCDataEntry {
	var pcDataEntries []PCDataEntry
	stackValue := 0
	state := newRegState()

	for _, inst := range instructions {
		state.update(inst)
		newStackValue := state.getStackSize()
		if newStackValue == stackValue {
			continue
		}

		pcDataEntries = append(pcDataEntries, PCDataEntry{
			Offset: inst.Offset + uintptr(inst.Len),
			Value:  int32(stackValue),
		})
		stackValue = newStackValue
	}

	
	pcDataEntries = append(pcDataEntries, PCDataEntry{
		Value:  int32(state.getStackSize()),
		Offset: lastOffset,
	})

	return pcDataEntries
}



func generatePCSP(startPC uintptr, endPC uintptr) ([]PCDataEntry, error) {
	instructions, err := read(startPC, endPC)
	if err != nil {
		return nil, err
	}

	return pcspFromInstructions(instructions, endPC-startPC), nil
}
