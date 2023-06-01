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

// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Buffered reading and decoding of DWARF data streams.

//lint:file-ignore ST1021 imported file

package util

import (
	"debug/dwarf"
	"fmt"
)


type buf struct {
	dwarf  *dwarf.Data
	format dataFormat
	name   string
	off    dwarf.Offset
	data   []byte
	Err    error
}



type dataFormat interface {
	
	version() int

	
	dwarf64() (dwarf64 bool, isKnown bool)

	
	addrsize() int
}


type UnknownFormat struct{}

func (u UnknownFormat) version() int {
	return 0
}

func (u UnknownFormat) dwarf64() (bool, bool) {
	return false, false
}

func (u UnknownFormat) addrsize() int {
	return 0
}

func MakeBuf(d *dwarf.Data, format dataFormat, name string, off dwarf.Offset, data []byte) buf {
	return buf{d, format, name, off, data, nil}
}

func (b *buf) Uint8() uint8 {
	if len(b.data) < 1 {
		b.error("underflow")
		return 0
	}
	val := b.data[0]
	b.data = b.data[1:]
	b.off++
	return val
}



func (b *buf) Varint() (c uint64, bits uint) {
	for i := 0; i < len(b.data); i++ {
		byte := b.data[i]
		c |= uint64(byte&0x7F) << bits
		bits += 7
		if byte&0x80 == 0 {
			b.off += dwarf.Offset(i + 1)
			b.data = b.data[i+1:]
			return c, bits
		}
	}
	return 0, 0
}


func (b *buf) Uint() uint64 {
	x, _ := b.Varint()
	return x
}


func (b *buf) Int() int64 {
	ux, bits := b.Varint()
	x := int64(ux)
	if x&(1<<(bits-1)) != 0 {
		x |= -1 << bits
	}
	return x
}


func (b *buf) AssertEmpty() {
	if len(b.data) == 0 {
		return
	}
	if len(b.data) > 5 {
		b.error(fmt.Sprintf("unexpected extra data: %x...", b.data[0:5]))
	}
	b.error(fmt.Sprintf("unexpected extra data: %x", b.data))
}

func (b *buf) error(s string) {
	if b.Err == nil {
		b.data = nil
		b.Err = dwarf.DecodeError{Name: b.name, Offset: b.off, Err: s}
	}
}
