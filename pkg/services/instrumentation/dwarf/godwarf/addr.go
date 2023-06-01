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

package godwarf

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
)



type DebugAddrSection struct {
	byteOrder binary.ByteOrder
	ptrSz     int
	data      []byte
}


func ParseAddr(data []byte) *DebugAddrSection {
	if len(data) == 0 {
		return nil
	}
	r := &DebugAddrSection{data: data}
	_, dwarf64, _, byteOrder := util.ReadDwarfLengthVersion(data)
	r.byteOrder = byteOrder
	data = data[6:]
	if dwarf64 {
		data = data[8:]
	}

	addrSz := data[0]
	segSelSz := data[1]
	r.ptrSz = int(addrSz + segSelSz)

	return r
}


func (addr *DebugAddrSection) GetSubsection(addrBase uint64) *DebugAddr {
	if addr == nil {
		return nil
	}
	return &DebugAddr{DebugAddrSection: addr, addrBase: addrBase}
}


type DebugAddr struct {
	*DebugAddrSection
	addrBase uint64
}


func (addr *DebugAddr) Get(idx uint64) (uint64, error) {
	if addr == nil || addr.DebugAddrSection == nil {
		return 0, errors.New("debug_addr section not present")
	}
	off := idx*uint64(addr.ptrSz) + addr.addrBase
	return util.ReadUintRaw(bytes.NewReader(addr.data[off:]), addr.byteOrder, addr.ptrSz)
}
