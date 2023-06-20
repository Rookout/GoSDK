package assembler

import (
	"math"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"golang.org/x/arch/arm64/arm64asm"
)

const NoArg = arm64asm.Reg(math.MaxInt16)

const placeholderOp = ANOOP
const placeholderLen = 4

type Arg = arm64asm.Arg
type Reg = arm64asm.Reg

func Imm(imm uint64) arm64asm.Imm64 {
	return arm64asm.Imm64{
		Imm: imm,
	}
}

type regReg struct {
	arm64asm.Arg
	reg1 arm64asm.Reg
	reg2 arm64asm.Reg
}

func RegReg(reg1 arm64asm.Reg, reg2 arm64asm.Reg) arm64asm.Arg {
	return regReg{reg1: reg1, reg2: reg2}
}

func argToAddr(arg arm64asm.Arg) (addr obj.Addr) {
	if arg == NoArg {
		return addr
	}

	switch t := arg.(type) {
	case arm64asm.Imm64:
		addr.Type = obj.TYPE_CONST
		addr.Offset = int64(t.Imm)
	case Reg:
		addr.Type = obj.TYPE_REG
		addr.Reg = AsmRegToSysReg(t)
	case Mem:
		addr.Type = obj.TYPE_MEM
		addr.Reg = AsmRegToSysReg(t.Base)
		addr.Offset = t.Disp
	case regReg:
		addr.Type = obj.TYPE_REGREG
		addr.Reg = AsmRegToSysReg(t.reg1)
		addr.Offset = int64(AsmRegToSysReg(t.reg2))
	}
	return addr
}
