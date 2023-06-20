package assembler

import (
	"math"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/arm64"
	"golang.org/x/arch/arm64/arm64asm"
)


const RegZero = arm64asm.Reg(math.MaxInt16 - 1)

var sysRegToDwarfReg = arm64.ARM64DWARFRegisters

var asmRegToSysReg = map[Reg]int16{
	arm64asm.X0:  arm64.REG_R0,
	arm64asm.X1:  arm64.REG_R1,
	arm64asm.X2:  arm64.REG_R2,
	arm64asm.X3:  arm64.REG_R3,
	arm64asm.X4:  arm64.REG_R4,
	arm64asm.X5:  arm64.REG_R5,
	arm64asm.X6:  arm64.REG_R6,
	arm64asm.X7:  arm64.REG_R7,
	arm64asm.X8:  arm64.REG_R8,
	arm64asm.X9:  arm64.REG_R9,
	arm64asm.X10: arm64.REG_R10,
	arm64asm.X11: arm64.REG_R11,
	arm64asm.X12: arm64.REG_R12,
	arm64asm.X13: arm64.REG_R13,
	arm64asm.X14: arm64.REG_R14,
	arm64asm.X15: arm64.REG_R15,
	arm64asm.X16: arm64.REG_R16,
	arm64asm.X17: arm64.REG_R17,
	arm64asm.X18: arm64.REG_R18,
	arm64asm.X19: arm64.REG_R19,
	arm64asm.X20: arm64.REG_R20,
	arm64asm.X21: arm64.REG_R21,
	arm64asm.X22: arm64.REG_R22,
	arm64asm.X23: arm64.REG_R23,
	arm64asm.X24: arm64.REG_R24,
	arm64asm.X25: arm64.REG_R25,
	arm64asm.X26: arm64.REG_R26,
	arm64asm.X27: arm64.REG_R27,
	arm64asm.X28: arm64.REG_R28,
	arm64asm.X29: arm64.REG_R29,
	arm64asm.X30: arm64.REG_R30,
	arm64asm.Q0:  arm64.REG_F0,
	arm64asm.Q1:  arm64.REG_F1,
	arm64asm.Q2:  arm64.REG_F2,
	arm64asm.Q3:  arm64.REG_F3,
	arm64asm.Q4:  arm64.REG_F4,
	arm64asm.Q5:  arm64.REG_F5,
	arm64asm.Q6:  arm64.REG_F6,
	arm64asm.Q7:  arm64.REG_F7,
	arm64asm.Q8:  arm64.REG_F8,
	arm64asm.Q9:  arm64.REG_F9,
	arm64asm.Q10: arm64.REG_F10,
	arm64asm.Q11: arm64.REG_F11,
	arm64asm.Q12: arm64.REG_F12,
	arm64asm.Q13: arm64.REG_F13,
	arm64asm.Q14: arm64.REG_F14,
	arm64asm.Q15: arm64.REG_F15,
	arm64asm.Q16: arm64.REG_F16,
	arm64asm.Q17: arm64.REG_F17,
	arm64asm.Q18: arm64.REG_F18,
	arm64asm.Q19: arm64.REG_F19,
	arm64asm.Q20: arm64.REG_F20,
	arm64asm.Q21: arm64.REG_F21,
	arm64asm.Q22: arm64.REG_F22,
	arm64asm.Q23: arm64.REG_F23,
	arm64asm.Q24: arm64.REG_F24,
	arm64asm.Q25: arm64.REG_F25,
	arm64asm.Q26: arm64.REG_F26,
	arm64asm.Q27: arm64.REG_F27,
	arm64asm.Q28: arm64.REG_F28,
	arm64asm.Q29: arm64.REG_F29,
	arm64asm.Q30: arm64.REG_F30,
	arm64asm.SP:  arm64.REG_RSP,
	RegZero:      arm64.REGZERO,
}
