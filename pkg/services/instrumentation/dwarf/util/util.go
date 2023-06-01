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

package util

import (
	"bytes"
	"debug/dwarf"
	"encoding/binary"
	"fmt"
	"io"
)



type ByteReaderWithLen interface {
	io.ByteReader
	io.Reader
	Len() int
}






func DecodeULEB128(buf ByteReaderWithLen) (uint64, uint32) {
	var (
		result uint64
		shift  uint64
		length uint32
	)

	if buf.Len() == 0 {
		return 0, 0
	}

	for {
		b, err := buf.ReadByte()
		if err != nil {
			panic("Could not parse ULEB128 value")
		}
		length++

		result |= uint64((uint(b) & 0x7f) << shift)

		
		if b&0x80 == 0 {
			break
		}

		shift += 7
	}

	return result, length
}



func DecodeSLEB128(buf ByteReaderWithLen) (int64, uint32) {
	var (
		b      byte
		err    error
		result int64
		shift  uint64
		length uint32
	)

	if buf.Len() == 0 {
		return 0, 0
	}

	for {
		b, err = buf.ReadByte()
		if err != nil {
			panic("Could not parse SLEB128 value")
		}
		length++

		result |= int64((int64(b) & 0x7f) << shift)
		shift += 7
		if b&0x80 == 0 {
			break
		}
	}

	if (shift < 8*uint64(length)) && (b&0x40 > 0) {
		result |= -(1 << shift)
	}

	return result, length
}



func EncodeULEB128(out io.ByteWriter, x uint64) {
	for {
		b := byte(x & 0x7f)
		x = x >> 7
		if x != 0 {
			b = b | 0x80
		}
		out.WriteByte(b)
		if x == 0 {
			break
		}
	}
}



func EncodeSLEB128(out io.ByteWriter, x int64) {
	for {
		b := byte(x & 0x7f)
		x >>= 7

		signb := b & 0x40

		last := false
		if (x == 0 && signb == 0) || (x == -1 && signb != 0) {
			last = true
		} else {
			b = b | 0x80
		}
		out.WriteByte(b)

		if last {
			break
		}
	}
}


func ParseString(data *bytes.Buffer) (string, error) {
	str, err := data.ReadString(0x0)
	if err != nil {
		return "", err
	}

	return str[:len(str)-1], nil
}


func ReadUintRaw(reader io.Reader, order binary.ByteOrder, ptrSize int) (uint64, error) {
	switch ptrSize {
	case 2:
		var n uint16
		if err := binary.Read(reader, order, &n); err != nil {
			return 0, err
		}
		return uint64(n), nil
	case 4:
		var n uint32
		if err := binary.Read(reader, order, &n); err != nil {
			return 0, err
		}
		return uint64(n), nil
	case 8:
		var n uint64
		if err := binary.Read(reader, order, &n); err != nil {
			return 0, err
		}
		return n, nil
	}
	return 0, fmt.Errorf("pointer size %d not supported", ptrSize)
}


func WriteUint(writer io.Writer, order binary.ByteOrder, ptrSize int, data uint64) error {
	switch ptrSize {
	case 4:
		return binary.Write(writer, order, uint32(data))
	case 8:
		return binary.Write(writer, order, data)
	}
	return fmt.Errorf("pointer size %d not supported", ptrSize)
}


func ReadDwarfLengthVersion(data []byte) (length uint64, dwarf64 bool, version uint8, byteOrder binary.ByteOrder) {
	if len(data) < 4 {
		return 0, false, 0, binary.LittleEndian
	}

	lengthfield := binary.LittleEndian.Uint32(data)
	voff := 4
	if lengthfield == ^uint32(0) {
		dwarf64 = true
		voff = 12
	}

	if voff+1 >= len(data) {
		return 0, false, 0, binary.LittleEndian
	}

	byteOrder = binary.LittleEndian
	x, y := data[voff], data[voff+1]
	switch {
	default:
		fallthrough
	case x == 0 && y == 0:
		version = 0
		byteOrder = binary.LittleEndian
	case x == 0:
		version = y
		byteOrder = binary.BigEndian
	case y == 0:
		version = x
		byteOrder = binary.LittleEndian
	}

	if dwarf64 {
		length = byteOrder.Uint64(data[4:])
	} else {
		length = uint64(byteOrder.Uint32(data))
	}

	return length, dwarf64, version, byteOrder
}

const (
	_DW_UT_compile = 0x1 + iota
	_DW_UT_type
	_DW_UT_partial
	_DW_UT_skeleton
	_DW_UT_split_compile
	_DW_UT_split_type
)


func ReadUnitVersions(data []byte) map[dwarf.Offset]uint8 {
	r := make(map[dwarf.Offset]uint8)
	off := dwarf.Offset(0)
	for len(data) > 0 {
		length, dwarf64, version, _ := ReadDwarfLengthVersion(data)

		data = data[4:]
		off += 4
		secoffsz := 4
		if dwarf64 {
			off += 8
			secoffsz = 8
			data = data[8:]
		}

		var headerSize int

		switch version {
		case 2, 3, 4:
			headerSize = 3 + secoffsz
		default: 
			unitType := data[2]

			switch unitType {
			case _DW_UT_compile, _DW_UT_partial:
				headerSize = 5 + secoffsz

			case _DW_UT_skeleton, _DW_UT_split_compile:
				headerSize = 4 + secoffsz + 8

			case _DW_UT_type, _DW_UT_split_type:
				headerSize = 4 + secoffsz + 8 + secoffsz
			}
		}

		r[off+dwarf.Offset(headerSize)] = version

		data = data[length:] 
		off += dwarf.Offset(length)
	}
	return r
}
