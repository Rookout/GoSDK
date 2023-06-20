// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

import (
	
	"unsafe"
)










type RegArgs struct {
	
	
	
	
	
	
	
	
	Ints   [IntArgRegs]uintptr  
	Floats [FloatArgRegs]uint64 

	

	
	
	
	Ptrs [IntArgRegs]unsafe.Pointer

	
	
	
	
	ReturnIsPtr IntArgRegBitmap
}

func (r *RegArgs) Dump() {
	print("Ints:")
	for _, x := range r.Ints {
		print(" ", x)
	}
	println()
	print("Floats:")
	for _, x := range r.Floats {
		print(" ", x)
	}
	println()
	print("Ptrs:")
	for _, x := range r.Ptrs {
		print(" ", x)
	}
	println()
}










func (r *RegArgs) IntRegArgAddr(reg int, argSize uintptr) unsafe.Pointer {
	if argSize > 8 || argSize == 0 || argSize&(argSize-1) != 0 {
		panic("invalid argSize")
	}
	offset := uintptr(0)
	if false {
		offset = 8 - argSize
	}
	return unsafe.Pointer(uintptr(unsafe.Pointer(&r.Ints[reg])) + offset)
}



type IntArgRegBitmap [(IntArgRegs + 7) / 8]uint8


func (b *IntArgRegBitmap) Set(i int) {
	b[i/8] |= uint8(1) << (i % 8)
}






//go:nosplit
func (b *IntArgRegBitmap) Get(i int) bool {
	return b[i/8]&(uint8(1)<<(i%8)) != 0
}
