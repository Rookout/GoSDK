//go:build arm64
// +build arm64

package prologue

import (
	_ "unsafe"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/utils"
	"golang.org/x/arch/arm64/arm64asm"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/assembler/common"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker/regbackup"
)

var stackUsageBuffer = 0x40 + 0x400 

var regsBackupBuffer = make([]regbackup.Backup, 1000)


func (g *Generator) generateCheckStackUsage(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Label(startLabel),
		common.MovGToX20(b),
		b.Inst(assembler.AMOVD, arm64asm.X20, assembler.Mem{Base: arm64asm.X20, Disp: common.StackguardOffset}),
		b.Sub3(arm64asm.X19, arm64asm.SP, assembler.Imm(uint64(g.stackUsage+stackUsageBuffer))),
		b.Cmp(arm64asm.X19, arm64asm.X20),
		b.BranchToLabel(assembler.ABGT, endLabel),
	)
}

func (g *Generator) getOriginalRegBackup() (regBackup []byte, regRestore []byte) {
	for _, inst := range g.epilogueInstructions {
		instBytes := utils.MakeSliceFromPointer(inst.PC, inst.Len)
		switch inst.Op {
		case arm64asm.STR, arm64asm.STP:
			regBackup = append(regBackup, instBytes...)
		case arm64asm.LDR, arm64asm.LDP:
			regRestore = append(regRestore, instBytes...)
		case arm64asm.MOV, arm64asm.BL, arm64asm.NOP:
		default:
			logger.Logger().Warningf("Found unexpected instruction in epilogue: %v", inst)
		}
	}
	return regBackup, regRestore
}

func (g *Generator) generateCallMorestack(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.AMOVD, arm64asm.X3, arm64asm.X30),
		b.Inst(assembler.AMOVD, arm64asm.X20, assembler.Imm(uint64(g.morestackAddr))),
		b.BranchToReg(assembler.ACALL, arm64asm.X20),
	)
}

func (g *Generator) generateJumpToStart(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.BranchToLabel(assembler.AJMP, startLabel),
		b.Label(endLabel),
		b.PsuedoNop(),
	)
}



func (g *Generator) generateCallFallback(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.BranchToLabel(assembler.AJMP, "afterFallback"),
		b.Label(fallbackLabel),
		b.Inst(assembler.AMOVD, arm64asm.X20, assembler.Imm(uint64(g.fallbackAddr))),
		b.BranchToReg(assembler.AJMP, arm64asm.X20),
		b.Label("afterFallback"),
	)
}
