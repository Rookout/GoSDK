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
	"encoding/binary"
)


type DwarfRegisters struct {
	staticBase uint64

	cfa       int64
	frameBase int64
	ObjBase   int64
	regs      []*DwarfRegister

	ByteOrder binary.ByteOrder
	PCRegNum  uint64
	SPRegNum  uint64
	BPRegNum  uint64
	LRRegNum  uint64

	FloatLoadError   error 
	loadMoreCallback func()
}

type DwarfRegister struct {
	Uint64Val uint64
	Bytes     []byte
}


func NewDwarfRegisters(staticBase uint64, regs []*DwarfRegister, byteOrder binary.ByteOrder, pcRegNum, spRegNum, bpRegNum, lrRegNum uint64) *DwarfRegisters {
	return &DwarfRegisters{
		staticBase: staticBase,
		regs:       regs,
		ByteOrder:  byteOrder,
		PCRegNum:   pcRegNum,
		SPRegNum:   spRegNum,
		BPRegNum:   bpRegNum,
		LRRegNum:   lrRegNum,
	}
}

func (regs *DwarfRegisters) CFA() int64                      { return regs.cfa }
func (regs *DwarfRegisters) StaticBase() uint64              { return regs.staticBase }
func (regs *DwarfRegisters) FrameBase() int64                { return regs.frameBase }
func (regs *DwarfRegisters) SetCFA(cfa int64)                { regs.cfa = cfa }
func (regs *DwarfRegisters) SetStaticBase(staticBase uint64) { regs.staticBase = staticBase }
func (regs *DwarfRegisters) SetFrameBase(frameBase int64)    { regs.frameBase = frameBase }



func (regs *DwarfRegisters) SetLoadMoreCallback(fn func()) {
	regs.loadMoreCallback = fn
}



func (regs *DwarfRegisters) CurrentSize() int {
	return len(regs.regs)
}


func (regs *DwarfRegisters) Uint64Val(idx uint64) uint64 {
	reg := regs.Reg(idx)
	if reg == nil {
		return 0
	}
	return regs.regs[idx].Uint64Val
}



func (regs *DwarfRegisters) Bytes(idx uint64) []byte {
	reg := regs.Reg(idx)
	if reg == nil {
		return nil
	}
	if reg.Bytes == nil {
		var buf bytes.Buffer
		binary.Write(&buf, regs.ByteOrder, reg.Uint64Val)
		reg.Bytes = buf.Bytes()
	}
	return reg.Bytes
}

func (regs *DwarfRegisters) loadMore() {
	if regs.loadMoreCallback == nil {
		return
	}
	regs.loadMoreCallback()
	regs.loadMoreCallback = nil
}


func (regs *DwarfRegisters) Reg(idx uint64) *DwarfRegister {
	if idx >= uint64(len(regs.regs)) {
		regs.loadMore()
		if idx >= uint64(len(regs.regs)) {
			return nil
		}
	}
	if regs.regs[idx] == nil {
		regs.loadMore()
	}
	return regs.regs[idx]
}

func (regs *DwarfRegisters) PC() uint64 {
	return regs.Uint64Val(regs.PCRegNum)
}

func (regs *DwarfRegisters) SP() uint64 {
	return regs.Uint64Val(regs.SPRegNum)
}

func (regs *DwarfRegisters) BP() uint64 {
	return regs.Uint64Val(regs.BPRegNum)
}


func (regs *DwarfRegisters) AddReg(idx uint64, reg *DwarfRegister) {
	if idx >= uint64(len(regs.regs)) {
		newRegs := make([]*DwarfRegister, idx+1)
		copy(newRegs, regs.regs)
		regs.regs = newRegs
	}
	regs.regs[idx] = reg
}


func (regs *DwarfRegisters) ClearRegisters() {
	regs.loadMoreCallback = nil
	for regnum := range regs.regs {
		regs.regs[regnum] = nil
	}
}

func DwarfRegisterFromUint64(v uint64) *DwarfRegister {
	return &DwarfRegister{Uint64Val: v}
}

func DwarfRegisterFromBytes(bytes []byte) *DwarfRegister {
	var v uint64
	switch len(bytes) {
	case 1:
		v = uint64(bytes[0])
	case 2:
		x := binary.LittleEndian.Uint16(bytes)
		v = uint64(x)
	case 4:
		x := binary.LittleEndian.Uint32(bytes)
		v = uint64(x)
	default:
		if len(bytes) >= 8 {
			v = binary.LittleEndian.Uint64(bytes[:8])
		}
	}
	return &DwarfRegister{Uint64Val: v, Bytes: bytes}
}
