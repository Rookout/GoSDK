// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package abi









func CommonSize(ptrSize int) int { return 4*ptrSize + 8 + 8 }


func StructFieldSize(ptrSize int) int { return 3 * ptrSize }



func UncommonSize() uint64 { return 4 + 2 + 2 + 4 + 4 }


func IMethodSize(ptrSize int) int { return 4 + 4 }


func KindOff(ptrSize int) int { return 2*ptrSize + 7 }


func SizeOff(ptrSize int) int { return 0 }


func PtrBytesOff(ptrSize int) int { return ptrSize }


func TFlagOff(ptrSize int) int { return 2*ptrSize + 4 }






type Offset struct {
	off        uint64 
	align      uint8  
	ptrSize    uint8  
	sliceAlign uint8  
}


func NewOffset(ptrSize uint8, twoWordAlignSlices bool) Offset {
	if twoWordAlignSlices {
		return Offset{off: 0, align: 1, ptrSize: ptrSize, sliceAlign: 2 * ptrSize}
	}
	return Offset{off: 0, align: 1, ptrSize: ptrSize, sliceAlign: ptrSize}
}

func assertIsAPowerOfTwo(x uint8) {
	if x == 0 {
		panic("Zero is not a power of two")
	}
	if x&-x == x {
		return
	}
	panic("Not a power of two")
}


func InitializedOffset(off int, align uint8, ptrSize uint8, twoWordAlignSlices bool) Offset {
	assertIsAPowerOfTwo(align)
	o0 := NewOffset(ptrSize, twoWordAlignSlices)
	o0.off = uint64(off)
	o0.align = align
	return o0
}

func (o Offset) align_(a uint8) Offset {
	o.off = (o.off + uint64(a) - 1) & ^(uint64(a) - 1)
	if o.align < a {
		o.align = a
	}
	return o
}



func (o Offset) Align(a uint8) Offset {
	assertIsAPowerOfTwo(a)
	return o.align_(a)
}


func (o Offset) plus(x uint64) Offset {
	o = o.align_(uint8(x))
	o.off += x
	return o
}


func (o Offset) D8() Offset {
	return o.plus(1)
}


func (o Offset) D16() Offset {
	return o.plus(2)
}


func (o Offset) D32() Offset {
	return o.plus(4)
}


func (o Offset) D64() Offset {
	return o.plus(8)
}


func (o Offset) P() Offset {
	if o.ptrSize == 0 {
		panic("This offset has no defined pointer size")
	}
	return o.plus(uint64(o.ptrSize))
}


func (o Offset) Slice() Offset {
	o = o.align_(o.sliceAlign)
	o.off += 3 * uint64(o.ptrSize)
	
	
	
	
	return o.Align(o.sliceAlign)
}


func (o Offset) String() Offset {
	o = o.align_(o.sliceAlign)
	o.off += 2 * uint64(o.ptrSize)
	return o 
}


func (o Offset) Interface() Offset {
	o = o.align_(o.sliceAlign)
	o.off += 2 * uint64(o.ptrSize)
	return o 
}



func (o Offset) Offset() uint64 {
	return o.Align(o.align).off
}

func (o Offset) PlusUncommon() Offset {
	o.off += UncommonSize()
	return o
}


func CommonOffset(ptrSize int, twoWordAlignSlices bool) Offset {
	return InitializedOffset(CommonSize(ptrSize), uint8(ptrSize), uint8(ptrSize), twoWordAlignSlices)
}
