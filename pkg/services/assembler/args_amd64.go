package assembler

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"golang.org/x/arch/x86/x86asm"
)

const NoArg = x86asm.Reg(0)

type Arg = x86asm.Arg
type Reg = x86asm.Reg

type Imm = x86asm.Imm


const placeholderOp = APUSHFQ
const placeholderLen = 1

func argToAddr(arg Arg) (addr obj.Addr) {
	if arg == NoArg {
		return addr
	}

	switch t := arg.(type) {
	case Imm:
		addr.Type = obj.TYPE_CONST
		addr.Offset = int64(t)
	case Reg:
		addr.Type = obj.TYPE_REG
		addr.Reg = AsmRegToSysReg(t)
	case Mem:
		addr.Type = obj.TYPE_MEM
		addr.Reg = AsmRegToSysReg(t.Base)
		addr.Offset = t.Disp
	}
	return addr
}
