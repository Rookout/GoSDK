package disassembler

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"golang.org/x/arch/arm64/arm64asm"
)

type inst = arm64asm.Inst

func (i *Instruction) GetDestPC() (uintptr, rookoutErrors.RookoutError) {
	if !IsCall(i) && !IsJump(i) {
		return 0, rookoutErrors.NewUnexpectedInstructionOp(i)
	}

	relDest, ok := i.Args[0].(arm64asm.PCRel)
	if !ok {
		
		if _, ok := i.Args[0].(arm64asm.Cond); !ok {
			return 0, rookoutErrors.NewArgIsNotRel(i)
		}

		relDest, ok = i.Args[1].(arm64asm.PCRel)
		if !ok {
			return 0, rookoutErrors.NewArgIsNotRel(i)
		}
	}
	return uintptr(int64(relDest) + int64(i.PC)), nil
}

func IsCall(i *Instruction) bool {
	switch i.Op {
	case arm64asm.BL,
		arm64asm.BLR:
		return true
	}
	return false
}

func IsJump(i *Instruction) bool {
	switch i.Op {
	case arm64asm.B,
		arm64asm.BR,
		arm64asm.CBNZ,
		arm64asm.CBZ,
		arm64asm.TBNZ,
		arm64asm.TBZ:
		return true
	}
	return false
}

func decode(bytes []byte) (*Instruction, error) {
	inst, err := arm64asm.Decode(bytes)
	if err != nil {
		return nil, err
	}

	return &Instruction{
		inst: inst,
		Len:  4,
	}, nil
}

func Decode(startPC uintptr, endPC uintptr, skipUnknown bool) ([]*Instruction, rookoutErrors.RookoutError) {
	var instructions []*Instruction
	funcLen := endPC - startPC
	funcAsm := utils.MakeSliceFromPointer(startPC, int(funcLen))
	offset := uintptr(0)

	for offset < funcLen {
		inst, err := decode(funcAsm[offset:])
		if err != nil {
			
			if err.Error() == "unknown instruction" && skipUnknown {
				offset += 4
				continue
			} else {
				return nil, rookoutErrors.NewFailedToDecode(funcAsm[offset:], err)
			}
		}

		inst.PC = startPC + uintptr(offset)
		inst.Offset = offset

		instructions = append(instructions, inst)
		offset += uintptr(inst.Len)
	}

	return instructions, nil
}
