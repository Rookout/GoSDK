package assembler

var dwarfRegToAsmReg = func() map[uint64]Reg {
	mapping := make(map[uint64]Reg, len(asmRegToSysReg))
	for asmReg, sysReg := range asmRegToSysReg {
		if dwarfReg, ok := sysRegToDwarfReg[sysReg]; ok {
			mapping[uint64(dwarfReg)] = asmReg
		}
	}
	return mapping
}()

func AsmRegToSysReg(reg Reg) int16 {
	return asmRegToSysReg[reg]
}

func DwarfRegToAsmReg(reg uint64) (Reg, bool) {
	res, ok := dwarfRegToAsmReg[reg]
	return res, ok
}
