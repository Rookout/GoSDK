//go:build amd64
// +build amd64

package prologue

import (
	"strings"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/assembler/common"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker/regbackup"
	"github.com/Rookout/GoSDK/pkg/utils"
	"golang.org/x/arch/x86/x86asm"
)

var stackUsageBuffer = 0x28 + 256 
var regsBackupBuffer = make([]regbackup.Backup, 1000)



func (g *Generator) generateCallFallback(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.BranchToLabel(assembler.AJMP, "afterFallback"),
		b.Label(fallbackLabel),
		b.Inst(assembler.AMOVQ, x86asm.R13, assembler.Imm(g.fallbackAddr)),
		b.BranchToReg(assembler.AJMP, x86asm.R13),
		b.Label("afterFallback"),
	)
}

func (g *Generator) getOriginalRegBackup() (regBackup []byte, regRestore []byte) {
	for _, inst := range g.epilogueInstructions {
		if !strings.HasPrefix(inst.Op.String(), "MOV") {
			continue
		}

		instBytes := utils.MakeSliceFromPointer(inst.PC, inst.Len)
		_, restore := inst.Args[0].(assembler.Reg)
		if restore {
			regRestore = append(regRestore, instBytes...)
		} else {
			regBackup = append(regBackup, instBytes...)
		}
	}
	return regBackup, regRestore
}


func (g *Generator) generateCheckStackUsage(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Label(startLabel),
		common.MovGToR12(b),
		b.Inst(assembler.AMOVQ, x86asm.R12, assembler.Mem{Base: x86asm.R12, Disp: common.StackguardOffset}),
		b.Inst(assembler.ALEAQ, x86asm.R13, assembler.Mem{Base: x86asm.RSP, Disp: -int64(g.stackUsage + stackUsageBuffer)}),
		b.Cmp(x86asm.R13, x86asm.R12),
		b.BranchToLabel(assembler.AJHI, endLabel),
	)
}

func (g *Generator) generateCallMorestack(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.AMOVQ, x86asm.R13, assembler.Imm(g.morestackAddr)),
		b.BranchToReg(assembler.ACALL, x86asm.R13),
	)
}

func (g *Generator) generateJumpToStart(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.BranchToLabel(assembler.AJMP, startLabel),
		b.Label(endLabel),
		b.PsuedoNop(),
	)
}
