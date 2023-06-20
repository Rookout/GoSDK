package regbackup

import (
	"fmt"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/assembler/common"
	"golang.org/x/arch/x86/x86asm"
)

type XMM struct {
	A uintptr
	B uintptr
}

type Backup struct {
	Lock  uintptr
	RDI   uintptr
	RAX   uintptr
	RBX   uintptr
	RCX   uintptr
	RDX   uintptr
	RSI   uintptr
	R8    uintptr
	R9    uintptr
	R10   uintptr
	R11   uintptr
	XMM0  XMM
	XMM1  XMM
	XMM2  XMM
	XMM3  XMM
	XMM4  XMM
	XMM5  XMM
	XMM6  XMM
	XMM7  XMM
	XMM8  XMM
	XMM9  XMM
	XMM10 XMM
	XMM11 XMM
	XMM12 XMM
	XMM13 XMM
	XMM14 XMM
}

var smallRegToOffsetInBackup = map[assembler.Reg]uintptr{
	x86asm.RDI: unsafe.Offsetof(Backup{}.RDI),
	x86asm.RAX: unsafe.Offsetof(Backup{}.RAX),
	x86asm.RBX: unsafe.Offsetof(Backup{}.RBX),
	x86asm.RCX: unsafe.Offsetof(Backup{}.RCX),
	x86asm.RDX: unsafe.Offsetof(Backup{}.RDX),
	x86asm.RSI: unsafe.Offsetof(Backup{}.RSI),
	x86asm.R8:  unsafe.Offsetof(Backup{}.R8),
	x86asm.R9:  unsafe.Offsetof(Backup{}.R9),
	x86asm.R10: unsafe.Offsetof(Backup{}.R10),
	x86asm.R11: unsafe.Offsetof(Backup{}.R11),
}
var bigRegToOffsetInBackup = map[assembler.Reg]uintptr{
	x86asm.X0:  unsafe.Offsetof(Backup{}.XMM0),
	x86asm.X1:  unsafe.Offsetof(Backup{}.XMM1),
	x86asm.X2:  unsafe.Offsetof(Backup{}.XMM2),
	x86asm.X3:  unsafe.Offsetof(Backup{}.XMM3),
	x86asm.X4:  unsafe.Offsetof(Backup{}.XMM4),
	x86asm.X5:  unsafe.Offsetof(Backup{}.XMM5),
	x86asm.X6:  unsafe.Offsetof(Backup{}.XMM6),
	x86asm.X7:  unsafe.Offsetof(Backup{}.XMM7),
	x86asm.X8:  unsafe.Offsetof(Backup{}.XMM8),
	x86asm.X9:  unsafe.Offsetof(Backup{}.XMM9),
	x86asm.X10: unsafe.Offsetof(Backup{}.XMM10),
	x86asm.X11: unsafe.Offsetof(Backup{}.XMM11),
	x86asm.X12: unsafe.Offsetof(Backup{}.XMM12),
	x86asm.X13: unsafe.Offsetof(Backup{}.XMM13),
	x86asm.X14: unsafe.Offsetof(Backup{}.XMM14),
}

var prevStackHi = assembler.Mem{Base: x86asm.RSP, Disp: 0x10}
var prevStackLo = assembler.Mem{Base: x86asm.RSP, Disp: 0x8}
var r12Backup = assembler.Mem{Base: x86asm.RSP, Disp: 0x0}

const backupSlotAddrReg = x86asm.R12

type Generator struct {
	stackUsage   int
	backupBuffer []Backup
	onFailLabel  string
	regsToUpdate []assembler.Reg
}

func NewGenerator(backupBuffer []Backup, onFailLabel string, regsToUpdate []assembler.Reg) *Generator {
	stackUsage := 0x8 
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
	backupSlotIndexReg := x86asm.R13

	return b.AddInstructions(
		
		b.Inst(assembler.AMOVQ, backupSlotAddrReg, x86asm.Imm(backupBufferAddr)),
		
		b.Inst(assembler.AMOVQ, backupSlotIndexReg, x86asm.Imm(1)),
		b.Label("findFreeBackupLoop"),
		b.Cmp(backupSlotIndexReg, x86asm.Imm(len(g.backupBuffer))),
		
		b.BranchToLabel(assembler.AJGT, g.onFailLabel),
		
		b.Inst(assembler.AXCHGQ, backupSlotIndexReg, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(unsafe.Offsetof(Backup{}.Lock))}),
		
		b.Cmp(backupSlotIndexReg, x86asm.Imm(0)),
		b.BranchToLabel(assembler.AJEQ, "backupRegs"),
		
		b.Inst(assembler.AADDQ, backupSlotAddrReg, x86asm.Imm(unsafe.Sizeof(Backup{}))),
		
		b.Inst(assembler.AADDQ, backupSlotIndexReg, x86asm.Imm(1)),
		
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

	
	for arg, offset := range smallRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.AMOVQ, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}, arg),
		)
		if err != nil {
			return err
		}
	}
	for arg, offset := range bigRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.AMOVUPS, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}, arg),
		)
		if err != nil {
			return err
		}
	}

	return nil
}


func (g *Generator) generateBackupStackAddrs(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		common.MovGToR12(b),
		b.Inst(assembler.AMOVQ, x86asm.R13, assembler.Mem{Base: x86asm.R12, Disp: common.StackLoOffset}),
		b.Inst(assembler.AMOVQ, prevStackLo, x86asm.R13),
		b.Inst(assembler.AMOVQ, x86asm.R13, assembler.Mem{Base: x86asm.R12, Disp: common.StackHiOffset}),
		b.Inst(assembler.AMOVQ, prevStackHi, x86asm.R13),
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
	for arg, offset := range smallRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.AMOVQ, arg, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}),
		)
		if err != nil {
			return err
		}
	}
	for arg, offset := range bigRegToOffsetInBackup {
		err := b.AddInstructions(
			b.Inst(assembler.AMOVUPS, arg, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(offset)}),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateReleaseBackupSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.AXORQ, x86asm.R13, x86asm.R13),
		b.Inst(assembler.AXCHGQ, assembler.Mem{Base: backupSlotAddrReg, Disp: int64(unsafe.Offsetof(Backup{}.Lock))}, x86asm.R13),
	)
}

func (g *Generator) GenerateRegRestore(b *assembler.Builder) rookoutErrors.RookoutError {
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
	err = b.AddInstructions(
		b.Inst(assembler.AADDQ, x86asm.RSP, assembler.Imm(g.stackUsage)),
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateBackupSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.ASUBQ, x86asm.RSP, assembler.Imm(g.stackUsage)),
		b.Inst(assembler.AMOVQ, r12Backup, x86asm.R12),
	)
}

func (g *Generator) generateRestoreSlot(b *assembler.Builder) rookoutErrors.RookoutError {
	return b.AddInstructions(
		b.Inst(assembler.AMOVQ, x86asm.R12, r12Backup),
	)
}

func (g *Generator) generateRegUpdate(b *assembler.Builder) rookoutErrors.RookoutError {
	for _, reg := range g.regsToUpdate {
		startLabel := fmt.Sprintf("%sBackupStart", reg.String())
		endLabel := fmt.Sprintf("%sBackupEnd", reg.String())
		err := b.AddInstructions(
			b.Label(startLabel),
			b.Cmp(reg, prevStackLo),
			b.BranchToLabel(assembler.AJLT, endLabel),
			b.Cmp(reg, prevStackHi),
			b.BranchToLabel(assembler.AJGT, endLabel),
			b.Inst(assembler.ASUBQ, reg, prevStackHi),
			common.MovGToR12(b),
			b.Inst(assembler.AADDQ, reg, assembler.Mem{Base: x86asm.R12, Disp: common.StackHiOffset}),
			b.Label(endLabel),
			b.PsuedoNop(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
