//go:build arm64
// +build arm64

package module

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"strconv"
	"strings"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/disassembler"
	"golang.org/x/arch/arm64/arm64asm"
)

type baseInstruction = arm64asm.Inst

type regState struct {
	regToValue  map[arm64asm.Reg]int
	skipUntilPC uintptr
}

func newRegState() *regState {
	r := &regState{
		regToValue: make(map[arm64asm.Reg]int),
	}
	r.regToValue[arm64asm.SP] = 0
	return r
}

func (r *regState) getStackSize() int {
	sp, _ := r.getRegValue(arm64asm.SP)
	
	return -sp
}

func (r *regState) getRegValue(reg arm64asm.Reg) (int, bool) {
	val, ok := r.regToValue[reg]
	return val, ok
}

func (r *regState) setRegValue(reg arm64asm.Reg, value int) {
	r.regToValue[reg] = value
}


const extShiftLDL = 9


type RegExtshiftAmount struct {
	reg       arm64asm.Reg
	extShift  uint8
	amount    uint8
	show_zero bool
}


type MemImmediate struct {
	Base arm64asm.RegSP
	Mode uint8
	imm  int32
}


type ImmShift struct {
	imm   uint16
	shift uint8
}

func getImmValue(immString string) int64 {
	immString = immString[1:]
	base := 10
	if strings.HasPrefix(immString, "0x") {
		base = 16
		immString = immString[2:]
	}
	immNum, _ := strconv.ParseInt(immString, base, 64)
	return immNum
}

func tryGetReg(reg interface{}) (arm64asm.Reg, bool) {
	if r, ok := reg.(arm64asm.Reg); ok {
		return r, true
	}
	if r, ok := reg.(arm64asm.RegSP); ok {
		return (arm64asm.Reg(r)), true
	}
	return 0, false
}







func (r *regState) updateAddrIndexInstruction(i *disassembler.Instruction) {
	var destArg arm64asm.Arg
	if i.Op == arm64asm.STP || i.Op == arm64asm.LDP {
		destArg = i.Args[2]
	} else {
		destArg = i.Args[1]
	}

	memArg, ok := destArg.(arm64asm.MemImmediate)
	
	if !ok || (memArg.Mode != arm64asm.AddrPreIndex && memArg.Mode != arm64asm.AddrPostIndex) {
		return
	}

	
	val, ok := r.getRegValue(arm64asm.Reg(memArg.Base))
	if !ok {
		
		return
	}

	
	mem := (*MemImmediate)(unsafe.Pointer(&memArg))

	
	val = val + int(mem.imm)
	r.setRegValue(arm64asm.Reg(mem.Base), val)
}







func (r *regState) updateAddSubInstruction(i *disassembler.Instruction) {
	destReg, ok := tryGetReg(i.Args[0])
	if !ok {
		logger.Logger().Warningf("Got unexpected dest reg in add/sub: %v [%T], instruction = %v", i.Args[0], i.Args[0], i)
		return
	}

	sourceReg, ok := tryGetReg(i.Args[1])
	if !ok {
		logger.Logger().Warningf("Got unexpected source reg in add/sub: %v [%T], instruction = %v", i.Args[1], i.Args[1], i)
		return
	}

	
	if imm, ok := i.Args[2].(arm64asm.ImmShift); ok {
		immNum := getImmValue(imm.String())

		val, ok := r.getRegValue(sourceReg)
		if !ok {
			
			return
		}

		if i.Op == arm64asm.SUB {
			r.setRegValue(destReg, val-int(immNum))
		} else {
			r.setRegValue(destReg, val+int(immNum))
		}
	}

	
	if reg, ok := i.Args[2].(arm64asm.RegExtshiftAmount); ok {
		
		newReg := *(*RegExtshiftAmount)(unsafe.Pointer(&reg))

		
		sourceVal2, ok := r.getRegValue(newReg.reg)
		if !ok {
			
			return
		}

		
		if newReg.extShift == extShiftLDL {
			sourceVal2 = sourceVal2 << newReg.amount
		} else if newReg.extShift != 0 {
			return
		}

		sourceVal1, ok := r.getRegValue(sourceReg)
		if !ok {
			
			return
		}

		if i.Op == arm64asm.SUB {
			r.setRegValue(destReg, sourceVal1-sourceVal2)
		} else {
			r.setRegValue(destReg, sourceVal1+sourceVal2)
		}
	}
}





func (r *regState) updateMovInstruction(i *disassembler.Instruction) {
	destReg, ok := tryGetReg(i.Args[0])
	if !ok {
		logger.Logger().Warningf("Got unexpected dest reg in mov: %v [%T], instruction = %v", i.Args[0], i.Args[0], i)
		return
	}

	
	if sourceReg, ok := tryGetReg(i.Args[1]); ok {
		val, ok := r.getRegValue(sourceReg)
		if !ok {
			
			return
		}
		r.setRegValue(destReg, val)
	}

	
	if sourceImm, ok := i.Args[1].(arm64asm.Imm64); ok {
		sourceNum := getImmValue(sourceImm.String())
		r.setRegValue(destReg, int(sourceNum))
	}

	
	if sourceImm, ok := i.Args[1].(arm64asm.Imm); ok {
		sourceNum := getImmValue(sourceImm.String())
		r.setRegValue(destReg, int(sourceNum))
	}
}

func (r *regState) update(i *disassembler.Instruction) {
	
	switch i.Op {
	case arm64asm.STR, arm64asm.STP, arm64asm.LDR, arm64asm.LDP:
		r.updateAddrIndexInstruction(i)

	case arm64asm.ADD, arm64asm.SUB:
		r.updateAddSubInstruction(i)

	case arm64asm.MOV:
		r.updateMovInstruction(i)
	}
}

func read(startPC uintptr, endPC uintptr) ([]*disassembler.Instruction, rookoutErrors.RookoutError) {
	
	instructions, err := disassembler.Decode(startPC, endPC, true)
	if err != nil {
		return nil, err
	}

	var skipUntil uintptr
	var prevInst *disassembler.Instruction
	var notSkipped []*disassembler.Instruction
	for _, inst := range instructions {
		if skipUntil > inst.Offset {
			continue
		}

		
		if prevInst != nil && prevInst.Op == arm64asm.B && inst.Op == arm64asm.NOP {
			
			
			
			if pcrel, ok := prevInst.Args[0].(arm64asm.PCRel); ok {
				skipUntil = prevInst.Offset + uintptr(pcrel)
			}
		}

		notSkipped = append(notSkipped, inst)
		prevInst = inst
	}

	return notSkipped, nil
}

func instructionSizeBytes(pc uintptr) (uintptr, rookoutErrors.RookoutError) {
	return 4, nil
}
