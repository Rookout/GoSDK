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

package memory

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)








type CompositeMemory struct {
	realmem MemoryReader
	regs    op.DwarfRegisters
	pieces  []op.Piece
	data    []byte
}

func NewCompositeMemory(mem MemoryReader, regs op.DwarfRegisters, pieces []op.Piece, pointerSize int) (*CompositeMemory, error) {
	cmem := &CompositeMemory{realmem: mem, regs: regs, pieces: pieces, data: []byte{}}
	for i := range pieces {
		piece := &pieces[i]
		switch piece.Kind {
		case op.RegPiece:
			reg := regs.Bytes(piece.Val)
			if piece.Size == 0 && i == len(pieces)-1 {
				piece.Size = len(reg)
			}
			if piece.Size > len(reg) {
				if regs.FloatLoadError != nil {
					return nil, fmt.Errorf("could not read %d bytes from register %d (size: %d), also error loading floating point registers: %v", piece.Size, piece.Val, len(reg), regs.FloatLoadError)
				}
				return nil, fmt.Errorf("could not read %d bytes from register %d (size: %d)", piece.Size, piece.Val, len(reg))
			}
			cmem.data = append(cmem.data, reg[:piece.Size]...)
		case op.AddrPiece:
			buf := make([]byte, piece.Size)
			_, err := mem.ReadMemory(buf, piece.Val)
			if err != nil {
				return nil, err
			}
			cmem.data = append(cmem.data, buf...)
		case op.ImmPiece:
			buf := piece.Bytes
			if buf == nil {
				sz := 8
				if piece.Size > sz {
					sz = piece.Size
				}
				if piece.Size == 0 && i == len(pieces)-1 {
					piece.Size = pointerSize 
				}
				buf = make([]byte, sz)
				binary.LittleEndian.PutUint64(buf, piece.Val)
			}
			cmem.data = append(cmem.data, buf[:piece.Size]...)
		default:
			panic("unsupported piece kind")
		}
	}
	return cmem, nil
}

func (m *CompositeMemory) ReadMemory(data []byte, addr uint64) (int, error) {
	addr -= FakeAddress
	if addr >= uint64(len(m.data)) || addr+uint64(len(data)) > uint64(len(m.data)) {
		return 0, errors.New("read out of bounds")
	}
	copy(data, m.data[addr:addr+uint64(len(data))])
	return len(data), nil
}







func DereferenceMemory(m MemoryReader) MemoryReader {
	if cmem, ok := m.(*CompositeMemory); ok {
		return cmem.realmem
	}

	return m
}

func (m *CompositeMemory) ID() string {
	return string(m.data)
}
