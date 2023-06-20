package regbackup

import (
	"fmt"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/assembler/common"
	"golang.org/x/arch/arm64/arm64asm"
)

type Q struct {
	A uintptr
	B uintptr
}

type Backup struct {
	Lock uintptr
	X0   uintptr
	X1   uintptr
	X2   uintptr
	X3   uintptr
	X4   uintptr
	X5   uintptr
	X6   uintptr
	X7   uintptr
	X8   uintptr
	X9   uintptr
	X10  uintptr
	X11  uintptr
	X12  uintptr
	X13  uintptr
	X14  uintptr
	X15  uintptr
	X29  uintptr
	X30  uintptr
	_    uintptr 
	Q0   Q
	Q1   Q
	Q2   Q
	Q3   Q
	Q4   Q
	Q5   Q
	Q6   Q
	Q7   Q
	Q8   Q
	Q9   Q
	Q10  Q
	Q11  Q
	Q12  Q
	Q13  Q
	Q14  Q
	Q15  Q
}

var prevStackLoStorage = assembler.Mem{Base: arm64asm.SP, Disp: 0x10}
var prevStackHiStorage = assembler.Mem{Base: arm64asm.SP, Disp: 0x18}

const prevStackHiReg = arm64asm.X16
const prevStackLoReg = arm64asm.X17
const backupSlotAddrReg = arm64asm.X19

var smallRegRegToOffsetInBackup = map[assembler.Arg]uintptr{
	assembler.RegReg(arm64asm.X0, arm64asm.X1):   unsafe.Offsetof(Backup{}.X0),
	assembler.RegReg(arm64asm.X2, arm64asm.X3):   unsafe.Offsetof(Backup{}.X2),
	assembler.RegReg(arm64asm.X4, arm64asm.X5):   unsafe.Offsetof(Backup{}.X4),
	assembler.RegReg(arm64asm.X6, arm64asm.X7):   unsafe.Offsetof(Backup{}.X6),
	assembler.RegReg(arm64asm.X8, arm64asm.X9):   unsafe.Offsetof(Backup{}.X8),
	assembler.RegReg(arm64asm.X10, arm64asm.X11): unsafe.Offsetof(Backup{}.X10),
	assembler.RegReg(arm64asm.X12, arm64asm.X13): unsafe.Offsetof(Backup{}.X12),
	assembler.RegReg(arm64asm.X14, arm64asm.X15): unsafe.Offsetof(Backup{}.X14),
	assembler.RegReg(arm64asm.X29, arm64asm.X30): unsafe.Offsetof(Backup{}.X29),
}
var bigRegRegToOffsetInBackup = map[assembler.Arg]uintptr{
	assembler.RegReg(arm64asm.Q0, arm64asm.Q1):   unsafe.Offsetof(Backup{}.Q0),
	assembler.RegReg(arm64asm.Q2, arm64asm.Q3):   unsafe.Offsetof(Backup{}.Q2),
	assembler.RegReg(arm64asm.Q4, arm64asm.Q5):   unsafe.Offsetof(Backup{}.Q4),
	assembler.RegReg(arm64asm.Q6, arm64asm.Q7):   unsafe.Offsetof(Backup{}.Q6),
	assembler.RegReg(arm64asm.Q8, arm64asm.Q9):   unsafe.Offsetof(Backup{}.Q8),
	assembler.RegReg(arm64asm.Q10, arm64asm.Q11): unsafe.Offsetof(Backup{}.Q10),
	assembler.RegReg(arm64asm.Q12, arm64asm.Q13): unsafe.Offsetof(Backup{}.Q12),
	assembler.RegReg(arm64asm.Q14, arm64asm.Q15): unsafe.Offsetof(Backup{}.Q14),
}

type Generator struct {
	stackUsage   int
	backupBuffer []Backup
	onFailLabel  string
	regsToUpdate []assembler.Reg
}

func NewGenerator(backupBuffer []Backup, onFailLabel string, regsToUpdate []assembler.Reg) *Generator {
	stackUsage := 0x20
	if len(regsToUpdate) > 0 {
		stackUsage += 0x10 
	}

	return &Generator{
		backupBuffer: backupBuffer,
		stackUsage:   stackUsage,
		onFailLabel:  onFailLabel,
		regsToUpdate: regsToUpdate,
	}
}

func (g *Generator) generateFindFreeBackupSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	backupBufferAddr := uintptr(unsafe.Pointer(&g.backupBuffer[0]))
	backupSlotIndexReg := arm64asm.X20

	return b.AddInstructions(
		
		b.Inst(assembler.AMOVD, backupSlotAddrReg, assembler.Imm(uint64(backupBufferAddr))),
		
		b.Inst(assembler.AMOVD, backupSlotIndexReg, assembler.Imm(1)),

		b.Label("findFreeBackupLoop"),
		b.Cmp(backupSlotIndexReg, assembler.Imm(uint64(len(g.backupBuffer)))),
		
		b.BranchToLabel(assembler.ABGT, g.onFailLabel), 
		
		b.Swpal(backupSlotIndexReg, backupSlotIndexReg, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(unsafe.Offsetof(Backup{}.Lock))}),
		
		b.BranchToLabel(assembler.ACBZ, "backupRegs", backupSlotIndexReg),

		
		b.Inst(assembler.AADD, backupSlotAddrReg, assembler.Imm(uint64(unsafe.Sizeof(g.backupBuffer[0])))),
		
		b.Inst(assembler.AADD, backupSlotIndexReg, assembler.Imm(1)),
		
		b.BranchToLabel(assembler.AJMP, "findFreeBackupLoop"),
	)
}

func (g *Generator) generateBackup(b *assembler.Builder) rookoutErrors.RookoutError {
	err := b.AddInstructions(
		b.Label("backupRegs"),
	)
	if err != nil {
		return err
	}

	for arg, offset := range smallRegRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.ASTP, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}, arg),
		)
		if err != nil {
			return err
		}
	}
	for arg, offset := range bigRegRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.AFSTPQ, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}, arg),
		)
		if err != nil {
			return err
		}
	}

	return nil
}


func (g *Generator) generateBackupStackAddrs(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		common.MovGToX20(b),
		b.Inst(assembler.AMOVD, arm64asm.X19, assembler.Mem{Base: arm64asm.X20, Disp: common.StackLoOffset}),
		b.Inst(assembler.AMOVD, prevStackLoStorage, arm64asm.X19),
		b.Inst(assembler.AMOVD, arm64asm.X19, assembler.Mem{Base: arm64asm.X20, Disp: common.StackHiOffset}),
		b.Inst(assembler.AMOVD, prevStackHiStorage, arm64asm.X19),
	)
}

func (g *Generator) GenerateRegBackup(b *assembler.Builder) rookoutErrors.RookoutError {
	err := g.generateFindFreeBackupSlot(b)
	if err != nil {
		return err
	}
	err = g.generateBackup(b)
	if err != nil {
		return err
	}

	err = g.generateBackupSlot(b)
	if err != nil {
		return err
	}
	if len(g.regsToUpdate) > 0 {
		err = g.generateBackupStackAddrs(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateRestore(b *assembler.Builder) rookoutErrors.RookoutError {
	for arg, offset := range smallRegRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.ALDP, arg, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}),
		)
		if err != nil {
			return err
		}
	}
	for arg, offset := range bigRegRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.AFLDPQ, arg, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateReleaseBackupSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Swpal(assembler.RegZero, assembler.RegZero, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(unsafe.Offsetof(Backup{}.Lock))}),
	)
}

func (g *Generator) GenerateRegRestore(b *assembler.Builder) rookoutErrors.RookoutError {
	if len(g.regsToUpdate) > 0 {
		err := b.AddInstructions(
			b.Inst(assembler.AMOVD, prevStackLoReg, prevStackLoStorage),
			b.Inst(assembler.AMOVD, prevStackHiReg, prevStackHiStorage),
		)
		if err != nil {
			return err
		}
	}
	err := g.generateRestoreSlot(b)
	if err != nil {
		return err
	}
	err = g.generateRestore(b)
	if err != nil {
		return err
	}
	err = g.generateReleaseBackupSlot(b)
	if err != nil {
		return err
	}
	err = g.generateRegUpdate(b)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateBackupSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.ASTP, assembler.Mem{Base: arm64asm.SP, Disp: -int64(g.stackUsage)}, assembler.RegReg(arm64asm.X30, backupSlotAddrReg), assembler.C_XPRE),
		b.Inst(assembler.AMOVD, assembler.Mem{Base: arm64asm.SP, Disp: -0x8}, arm64asm.X29),
		b.Sub3(arm64asm.X29, arm64asm.SP, assembler.Imm(0x8)),
	)
}

func (g *Generator) generateRestoreSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.ALDP, assembler.RegReg(arm64asm.X30, backupSlotAddrReg), assembler.Mem{Base: arm64asm.SP, Disp: int64(g.stackUsage)}, assembler.C_XPOST),
	)
}

func (g *Generator) generateRegUpdate(b *assembler.Builder) rookoutErrors.RookoutError {
	for _, reg := range g.regsToUpdate {
		startLabel := fmt.Sprintf("%sBackupStart", reg.String())
		endLabel := fmt.Sprintf("%sBackupEnd", reg.String())
		err := b.AddInstructions(
			b.Label(startLabel),
			b.Cmp(reg, prevStackLoReg),
			b.BranchToLabel(assembler.ABLT, endLabel),
			b.Cmp(reg, prevStackHiReg),
			b.BranchToLabel(assembler.ABGT, endLabel),
			b.Inst(assembler.ASUB, reg, prevStackHiReg),
			common.MovGToX20(b),
			b.Inst(assembler.AMOVD, arm64asm.X20, assembler.Mem{Base: arm64asm.X20, Disp: common.StackHiOffset}),
			b.Inst(assembler.AADD, reg, arm64asm.X20),
			b.Label(endLabel),
			b.PsuedoNop(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
