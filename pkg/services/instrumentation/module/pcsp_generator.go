package module

import (
	"reflect"
	"unsafe"
)

func pcspFromInstructions(instructions []*instruction) []PCDataEntry {
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

	
	lastInstruction := instructions[len(instructions)-1]
	pcDataEntries = append(pcDataEntries, PCDataEntry{
		Value:  int32(state.getStackSize()),
		Offset: lastInstruction.Offset + uintptr(lastInstruction.Len),
	})

	return pcDataEntries
}



func generatePCSP(startPC uintptr, endPC uintptr) ([]PCDataEntry, error) {
	instructions, err := read(startPC, endPC)
	if err != nil {
		return nil, err
	}

	return pcspFromInstructions(instructions), nil
}

func makeSliceFromPointer(p uintptr, length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: p,
		Len:  length,
		Cap:  length,
	}))
}

type instruction struct {
	baseInstruction
	Len    int
	PC     uintptr
	Offset uintptr
}
