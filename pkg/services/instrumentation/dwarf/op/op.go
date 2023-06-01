// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package op

import (
	"bytes"
	"errors"
	"fmt"
)



type Opcode byte


type Piece struct {
	Size  int
	Kind  PieceKind
	Val   uint64
	Bytes []byte
}


type PieceKind uint8

const (
	AddrPiece PieceKind = iota 
	RegPiece                   
	ImmPiece                   
)




func ExecuteStackProgram(regs DwarfRegisters, instructions []byte, ptrSize int) (int64, []Piece, error) {
	dwarfLocator, err := NewDwarfLocator(instructions, ptrSize)
	if err != nil {
		return 0, nil, err
	}

	return dwarfLocator.Locate(regs)
}

type DwarfLocator struct {
	executors []OpcodeExecutor
	ptrSize   int
}

func NewDwarfLocator(instructions []byte, ptrSize int) (*DwarfLocator, error) {
	var executors []OpcodeExecutor
	ctx := &OpcodeExecutorCreatorContext{
		prog:        make([]byte, len(instructions)),
		pointerSize: ptrSize,
	}
	copy(ctx.prog, instructions)
	buf := bytes.NewBuffer(instructions)
	ctx.buf = buf

	for {
		opcodeByte, err := buf.ReadByte()
		if err != nil {
			break
		}
		opcode := Opcode(opcodeByte)
		if opcode == DW_OP_nop {
			continue
		}
		executorCreator, ok := OpcodeToExecutorCreator(opcode)
		if !ok {
			return nil, fmt.Errorf("invalid instruction %#v", opcode)
		}

		executor, err := executorCreator(opcode, ctx)
		if err != nil {
			return nil, err
		}
		executors = append(executors, executor)
	}

	return &DwarfLocator{executors: executors, ptrSize: ptrSize}, nil
}

func (d *DwarfLocator) newDwarfLocatorContext(regs DwarfRegisters) *OpcodeExecutorContext {
	return &OpcodeExecutorContext{
		Stack:          make([]int64, 0, 3),
		PtrSize:        d.ptrSize,
		DwarfRegisters: regs,
	}
}

func (d *DwarfLocator) Locate(regs DwarfRegisters) (int64, []Piece, error) {
	ctx := d.newDwarfLocatorContext(regs)

	for _, executor := range d.executors {
		if err := executor.Execute(ctx); err != nil {
			return 0, nil, err
		}
	}

	if ctx.Pieces != nil {
		if len(ctx.Pieces) == 1 && ctx.Pieces[0].Kind == RegPiece {
			return int64(regs.Uint64Val(ctx.Pieces[0].Val)), ctx.Pieces, nil
		}
		return 0, ctx.Pieces, nil
	}

	if len(ctx.Stack) == 0 {
		return 0, nil, errors.New("empty OP stack")
	}

	return ctx.Stack[len(ctx.Stack)-1], nil, nil
}
