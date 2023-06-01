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

// Package frame contains data structures and
// related functions for parsing and searching
// through Dwarf .debug_frame data.
package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
)

type parsefunc func(*parseContext) parsefunc

type parseContext struct {
	staticBase uint64

	buf         *bytes.Buffer
	totalLen    int
	entries     FrameDescriptionEntries
	ciemap      map[int]*CommonInformationEntry
	common      *CommonInformationEntry
	frame       *FrameDescriptionEntry
	length      uint32
	ptrSize     int
	ehFrameAddr uint64
	err         error
}






func Parse(data []byte, order binary.ByteOrder, staticBase uint64, ptrSize int, ehFrameAddr uint64) (FrameDescriptionEntries, error) {
	var (
		buf  = bytes.NewBuffer(data)
		pctx = &parseContext{buf: buf, totalLen: len(data), entries: newFrameIndex(), staticBase: staticBase, ptrSize: ptrSize, ehFrameAddr: ehFrameAddr, ciemap: map[int]*CommonInformationEntry{}}
	)

	for fn := parselength; buf.Len() != 0; {
		fn = fn(pctx)
		if pctx.err != nil {
			return nil, pctx.err
		}
	}

	for i := range pctx.entries {
		pctx.entries[i].order = order
	}

	return pctx.entries, nil
}

func (ctx *parseContext) parsingEHFrame() bool {
	return ctx.ehFrameAddr > 0
}

func (ctx *parseContext) cieEntry(cieid uint32) bool {
	if ctx.parsingEHFrame() {
		return cieid == 0x00
	}
	return cieid == 0xffffffff
}

func (ctx *parseContext) offset() int {
	return ctx.totalLen - ctx.buf.Len()
}

func parselength(ctx *parseContext) parsefunc {
	start := ctx.offset()
	binary.Read(ctx.buf, binary.LittleEndian, &ctx.length) 

	if ctx.length == 0 {
		
		return parselength
	}

	var cieid uint32
	binary.Read(ctx.buf, binary.LittleEndian, &cieid)

	ctx.length -= 4 

	if ctx.cieEntry(cieid) {
		ctx.common = &CommonInformationEntry{Length: ctx.length, staticBase: ctx.staticBase, CIE_id: cieid}
		ctx.ciemap[start] = ctx.common
		return parseCIE
	}

	if ctx.ehFrameAddr > 0 {
		cieid = uint32(start - int(cieid) + 4)
	}

	common := ctx.ciemap[int(cieid)]

	if common == nil {
		ctx.err = fmt.Errorf("unknown CIE_id %#x at %#x", cieid, start)
	}

	ctx.frame = &FrameDescriptionEntry{Length: ctx.length, CIE: common}
	return parseFDE
}

func parseFDE(ctx *parseContext) parsefunc {
	startOff := ctx.offset()
	r := ctx.buf.Next(int(ctx.length))

	reader := bytes.NewReader(r)
	num := ctx.readEncodedPtr(addrSum(ctx.ehFrameAddr+uint64(startOff), reader), reader, ctx.frame.CIE.ptrEncAddr)
	ctx.frame.begin = num + ctx.staticBase

	
	
	
	
	sizePtrEnc := ctx.frame.CIE.ptrEncAddr & 0x0f
	ctx.frame.size = ctx.readEncodedPtr(0, reader, sizePtrEnc)

	
	
	ctx.entries = append(ctx.entries, ctx.frame)

	if ctx.parsingEHFrame() && len(ctx.frame.CIE.Augmentation) > 0 {
		
		
		
		n, _ := util.DecodeULEB128(reader)
		reader.Seek(int64(n), io.SeekCurrent)
	}

	
	
	

	off, _ := reader.Seek(0, io.SeekCurrent)
	ctx.frame.Instructions = r[off:]
	ctx.length = 0

	return parselength
}

func addrSum(base uint64, buf *bytes.Reader) uint64 {
	n, _ := buf.Seek(0, io.SeekCurrent)
	return base + uint64(n)
}

func parseCIE(ctx *parseContext) parsefunc {
	data := ctx.buf.Next(int(ctx.length))
	buf := bytes.NewBuffer(data)
	
	ctx.common.Version, _ = buf.ReadByte()

	
	ctx.common.Augmentation, _ = util.ParseString(buf)

	if ctx.parsingEHFrame() {
		if ctx.common.Augmentation == "eh" {
			ctx.err = fmt.Errorf("unsupported 'eh' augmentation at %#x", ctx.offset())
		}
		if len(ctx.common.Augmentation) > 0 && ctx.common.Augmentation[0] != 'z' {
			ctx.err = fmt.Errorf("unsupported augmentation at %#x (does not start with 'z')", ctx.offset())
		}
	}

	
	ctx.common.CodeAlignmentFactor, _ = util.DecodeULEB128(buf)

	
	ctx.common.DataAlignmentFactor, _ = util.DecodeSLEB128(buf)

	
	if ctx.parsingEHFrame() && ctx.common.Version == 1 {
		b, _ := buf.ReadByte()
		ctx.common.ReturnAddressRegister = uint64(b)
	} else {
		ctx.common.ReturnAddressRegister, _ = util.DecodeULEB128(buf)
	}

	ctx.common.ptrEncAddr = ptrEncAbs

	if ctx.parsingEHFrame() && len(ctx.common.Augmentation) > 0 {
		_, _ = util.DecodeULEB128(buf) 
		for i := 1; i < len(ctx.common.Augmentation); i++ {
			switch ctx.common.Augmentation[i] {
			case 'L':
				_, _ = buf.ReadByte() 
			case 'R':
				
				b, _ := buf.ReadByte()
				ctx.common.ptrEncAddr = ptrEnc(b)
				if !ctx.common.ptrEncAddr.Supported() {
					ctx.err = fmt.Errorf("pointer encoding not supported %#x at %#x", ctx.common.ptrEncAddr, ctx.offset())
					return nil
				}
			case 'S':
				
			case 'P':
				
				
				
				
				e, _ := buf.ReadByte()
				if !ptrEnc(e).Supported() {
					ctx.err = fmt.Errorf("pointer encoding not supported %#x at %#x", e, ctx.offset())
					return nil
				}
				ctx.readEncodedPtr(0, buf, ptrEnc(e))
			default:
				ctx.err = fmt.Errorf("unsupported augmentation character %c at %#x", ctx.common.Augmentation[i], ctx.offset())
				return nil
			}
		}
	}

	
	
	
	
	ctx.common.InitialInstructions = buf.Bytes() 
	ctx.length = 0

	return parselength
}







func (ctx *parseContext) readEncodedPtr(addr uint64, buf util.ByteReaderWithLen, ptrEnc ptrEnc) uint64 {
	if ptrEnc == ptrEncOmit {
		return 0
	}

	var ptr uint64

	switch ptrEnc & 0xf {
	case ptrEncAbs, ptrEncSigned:
		ptr, _ = util.ReadUintRaw(buf, binary.LittleEndian, ctx.ptrSize)
	case ptrEncUleb:
		ptr, _ = util.DecodeULEB128(buf)
	case ptrEncUdata2:
		ptr, _ = util.ReadUintRaw(buf, binary.LittleEndian, 2)
	case ptrEncSdata2:
		ptr, _ = util.ReadUintRaw(buf, binary.LittleEndian, 2)
		ptr = uint64(int16(ptr))
	case ptrEncUdata4:
		ptr, _ = util.ReadUintRaw(buf, binary.LittleEndian, 4)
	case ptrEncSdata4:
		ptr, _ = util.ReadUintRaw(buf, binary.LittleEndian, 4)
		ptr = uint64(int32(ptr))
	case ptrEncUdata8, ptrEncSdata8:
		ptr, _ = util.ReadUintRaw(buf, binary.LittleEndian, 8)
	case ptrEncSleb:
		n, _ := util.DecodeSLEB128(buf)
		ptr = uint64(n)
	}

	if ptrEnc&0xf0 == ptrEncPCRel {
		ptr += addr
	}

	return ptr
}



func DwarfEndian(infoSec []byte) binary.ByteOrder {
	if len(infoSec) < 6 {
		return binary.BigEndian
	}
	x, y := infoSec[4], infoSec[5]
	switch {
	case x == 0 && y == 0:
		return binary.BigEndian
	case x == 0:
		return binary.BigEndian
	case y == 0:
		return binary.LittleEndian
	default:
		return binary.BigEndian
	}
}
