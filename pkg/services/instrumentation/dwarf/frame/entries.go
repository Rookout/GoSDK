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

package frame

import (
	"encoding/binary"
	"fmt"
	"sort"
)



type CommonInformationEntry struct {
	Length                uint32
	CIE_id                uint32
	Version               uint8
	Augmentation          string
	CodeAlignmentFactor   uint64
	DataAlignmentFactor   int64
	ReturnAddressRegister uint64
	InitialInstructions   []byte
	staticBase            uint64

	
	ptrEncAddr ptrEnc
}



type FrameDescriptionEntry struct {
	Length       uint32
	CIE          *CommonInformationEntry
	Instructions []byte
	begin, size  uint64
	order        binary.ByteOrder
}



func (fde *FrameDescriptionEntry) Cover(addr uint64) bool {
	return (addr - fde.begin) < fde.size
}


func (fde *FrameDescriptionEntry) Begin() uint64 {
	return fde.begin
}


func (fde *FrameDescriptionEntry) End() uint64 {
	return fde.begin + fde.size
}


func (fde *FrameDescriptionEntry) Translate(delta uint64) {
	fde.begin += delta
}


func (fde *FrameDescriptionEntry) EstablishFrame(pc uint64) *FrameContext {
	return executeDwarfProgramUntilPC(fde, pc)
}

type FrameDescriptionEntries []*FrameDescriptionEntry

func newFrameIndex() FrameDescriptionEntries {
	return make(FrameDescriptionEntries, 0, 1000)
}


type ErrNoFDEForPC struct {
	PC uint64
}

func (err *ErrNoFDEForPC) Error() string {
	return fmt.Sprintf("could not find FDE for PC %#v", err.PC)
}


func (fdes FrameDescriptionEntries) FDEForPC(pc uint64) (*FrameDescriptionEntry, error) {
	idx := sort.Search(len(fdes), func(i int) bool {
		return fdes[i].Cover(pc) || fdes[i].Begin() >= pc
	})
	if idx == len(fdes) || !fdes[idx].Cover(pc) {
		return nil, &ErrNoFDEForPC{pc}
	}
	return fdes[idx], nil
}


func (fdes FrameDescriptionEntries) Append(otherFDEs FrameDescriptionEntries) FrameDescriptionEntries {
	r := append(fdes, otherFDEs...)
	sort.SliceStable(r, func(i, j int) bool {
		return r[i].Begin() < r[j].Begin()
	})
	
	uniqFDEs := fdes[:0]
	for _, fde := range fdes {
		if len(uniqFDEs) > 0 {
			last := uniqFDEs[len(uniqFDEs)-1]
			if last.Begin() == fde.Begin() && last.End() == fde.End() {
				continue
			}
		}
		uniqFDEs = append(uniqFDEs, fde)
	}
	return r
}







type ptrEnc uint8

const (
	ptrEncAbs    ptrEnc = 0x00 
	ptrEncOmit   ptrEnc = 0xff 
	ptrEncUleb   ptrEnc = 0x01 
	ptrEncUdata2 ptrEnc = 0x02 
	ptrEncUdata4 ptrEnc = 0x03 
	ptrEncUdata8 ptrEnc = 0x04 
	ptrEncSigned ptrEnc = 0x08 
	ptrEncSleb   ptrEnc = 0x09 
	ptrEncSdata2 ptrEnc = 0x0a 
	ptrEncSdata4 ptrEnc = 0x0b 
	ptrEncSdata8 ptrEnc = 0x0c 

	ptrEncPCRel    ptrEnc = 0x10 
	ptrEncTextRel  ptrEnc = 0x20 
	ptrEncDataRel  ptrEnc = 0x30 
	ptrEncFuncRel  ptrEnc = 0x40 
	ptrEncAligned  ptrEnc = 0x50 
	ptrEncIndirect ptrEnc = 0x80 
)


func (ptrEnc ptrEnc) Supported() bool {
	if ptrEnc != ptrEncOmit {
		szenc := ptrEnc & 0x0f
		if ((szenc > ptrEncUdata8) && (szenc < ptrEncSigned)) || (szenc > ptrEncSdata8) {
			
			return false
		}
		if ptrEnc&0xf0 != ptrEncPCRel {
			
			return false
		}
	}
	return true
}
