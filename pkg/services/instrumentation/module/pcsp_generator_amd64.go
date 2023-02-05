//go:build amd64
// +build amd64

package module

import (
	"github.com/Rookout/GoSDK/pkg/logger"
	"golang.org/x/arch/x86/x86asm"
)

type baseInstruction = x86asm.Inst

type regState struct {
	stackSize int
}

func newRegState() *regState {
	return &regState{}
}

func (r *regState) getStackSize() int {
	return r.stackSize
}

func (r *regState) setStackSize(stackSize int) {
	r.stackSize = stackSize
}

func (r *regState) update(i *instruction) {
	switch i.Op {
	case x86asm.ADD, x86asm.SUB:
		if i.Args[0] != x86asm.RSP {
			return
		}

		imm, ok := i.Args[1].(x86asm.Imm)
		if !ok {
			logger.Logger().Warningf("Got unexpected source reg in add/sub: %v [%T], instruction = %v", i.Args[1], i.Args[1], i)
			return
		}

		
		if i.Op == x86asm.ADD {
			r.setStackSize(r.getStackSize() - int(imm))
		} else {
			r.setStackSize(r.getStackSize() + int(imm))
		}

	
	case x86asm.PUSH, x86asm.PUSHFQ:
		r.setStackSize(r.getStackSize() + 8)

	
	case x86asm.POP, x86asm.POPFQ:
		r.setStackSize(r.getStackSize() - 8)
	}
}

func decode(bytes []byte) (*instruction, error) {
	inst, err := x86asm.Decode(bytes, 64)
	if err != nil {
		return nil, err
	}

	return &instruction{
		baseInstruction: inst,
		Len:             inst.Len,
	}, nil
}

func read(startPC uintptr, endPC uintptr) ([]*instruction, error) {
	var instructions []*instruction
	funcLen := endPC - startPC
	funcAsm := makeSliceFromPointer(startPC, int(funcLen))
	offset := uintptr(0)

	for offset < funcLen {
		inst, err := decode(funcAsm[offset:])
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
