







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
