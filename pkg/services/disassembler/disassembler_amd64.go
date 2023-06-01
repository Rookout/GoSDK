package disassembler

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"golang.org/x/arch/x86/x86asm"
)

var maxInstLen = 15

type inst = x86asm.Inst

func (i *Instruction) GetDestPC() (uintptr, rookoutErrors.RookoutError) {
	if !IsCall(i) && !IsJump(i) {
		return 0, rookoutErrors.NewUnexpectedInstructionOp(i)
	}

	relDest, ok := i.Args[0].(x86asm.Rel)
	if !ok {
		return 0, rookoutErrors.NewArgIsNotRel(i)
	}
	return uintptr(int64(relDest) + int64(i.PC) + int64(i.Len)), nil
}

func IsCall(i *Instruction) bool {
	return i.Op == x86asm.CALL
}

func IsJump(i *Instruction) bool {
	switch i.Op {
	case x86asm.JA,
		x86asm.JAE,
		x86asm.JB,
		x86asm.JBE,
		x86asm.JCXZ,
		x86asm.JE,
		x86asm.JECXZ,
		x86asm.JG,
		x86asm.JGE,
		x86asm.JL,
		x86asm.JLE,
		x86asm.JMP,
		x86asm.JNE,
		x86asm.JNO,
		x86asm.JNP,
		x86asm.JNS,
		x86asm.JO,
		x86asm.JP,
		x86asm.JRCXZ,
		x86asm.JS:
		return true
	}
	return false
}

func decodeOne(bytes []byte) (*Instruction, rookoutErrors.RookoutError) {
	inst, err := x86asm.Decode(bytes, 64)
	if err != nil {
		return nil, rookoutErrors.NewFailedToDecode(bytes, err)
	}

	return &Instruction{
		inst: inst,
		Len:  inst.Len,
	}, nil
}


func DecodeOne(startPC uintptr) (*Instruction, rookoutErrors.RookoutError) {
	funcAsm := utils.MakeSliceFromPointer(startPC, maxInstLen)
	inst, err := decodeOne(funcAsm)
	if err != nil {
		return nil, rookoutErrors.NewFailedToDecode(funcAsm, err)
	}

	inst.PC = startPC
	return inst, nil
}


func Decode(startPC uintptr, endPC uintptr, _ bool) ([]*Instruction, rookoutErrors.RookoutError) {
	var instructions []*Instruction
	funcLen := endPC - startPC
	funcAsm := utils.MakeSliceFromPointer(startPC, int(funcLen))
	offset := uintptr(0)

	for offset < funcLen {
		inst, err := decodeOne(funcAsm[offset:])
		if err != nil {
			return nil, err
		}

		inst.PC = startPC + uintptr(offset)
		inst.Offset = offset

		instructions = append(instructions, inst)
		offset += uintptr(inst.Len)
	}

	return instructions, nil
}
