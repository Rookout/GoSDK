package disassembler

type Instruction struct {
	inst
	Len    int
	PC     uintptr
	Offset uintptr
}

func GetFirstInstruction(instructions []*Instruction, filter func(i *Instruction) bool) (*Instruction, int, bool) {
	for i, inst := range instructions {
		if filter(inst) {
			return inst, i, true
		}
	}
	return nil, 0, false
}
