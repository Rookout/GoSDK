package assembler

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/x86"
	"golang.org/x/arch/x86/x86asm"
)

var sysRegToDwarfReg = x86.AMD64DWARFRegisters

var asmRegToSysReg = map[Reg]int16{
	x86asm.RAX: x86.REG_AX,
	x86asm.RCX: x86.REG_CX,
	x86asm.RDX: x86.REG_DX,
	x86asm.RBX: x86.REG_BX,
	x86asm.RSP: x86.REG_SP,
	x86asm.RBP: x86.REG_BP,
	x86asm.RSI: x86.REG_SI,
	x86asm.RDI: x86.REG_DI,
	x86asm.R8:  x86.REG_R8,
	x86asm.R9:  x86.REG_R9,
	x86asm.R10: x86.REG_R10,
	x86asm.R11: x86.REG_R11,
	x86asm.R12: x86.REG_R12,
	x86asm.R13: x86.REG_R13,
	x86asm.R14: x86.REG_R14,
	x86asm.R15: x86.REG_R15,
	x86asm.X0:  x86.REG_X0,
	x86asm.X1:  x86.REG_X1,
	x86asm.X2:  x86.REG_X2,
	x86asm.X3:  x86.REG_X3,
	x86asm.X4:  x86.REG_X4,
	x86asm.X5:  x86.REG_X5,
	x86asm.X6:  x86.REG_X6,
	x86asm.X7:  x86.REG_X7,
	x86asm.X8:  x86.REG_X8,
	x86asm.X9:  x86.REG_X9,
	x86asm.X10: x86.REG_X10,
	x86asm.X11: x86.REG_X11,
	x86asm.X12: x86.REG_X12,
	x86asm.X13: x86.REG_X13,
	x86asm.X14: x86.REG_X14,
	x86asm.X15: x86.REG_X15,
	x86asm.FS:  x86.REG_FS,
}
