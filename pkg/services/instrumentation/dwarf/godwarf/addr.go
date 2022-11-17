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
