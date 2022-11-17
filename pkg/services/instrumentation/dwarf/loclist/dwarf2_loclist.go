package loclist

import (
	"encoding/binary"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)


type Reader interface {
	Find(off int, staticBase, base, pc uint64, debugAddr *godwarf.DebugAddr) (*Entry, error)
	Empty() bool
}


type Dwarf2Reader struct {
	data  []byte
	cur   int
	ptrSz int
}


func NewDwarf2Reader(data []byte, ptrSz int) *Dwarf2Reader {
	return &Dwarf2Reader{data: data, ptrSz: ptrSz}
}


func (rdr *Dwarf2Reader) Empty() bool {
	return rdr.data == nil
}


func (rdr *Dwarf2Reader) Seek(off int) {
	rdr.cur = off
}



func (rdr *Dwarf2Reader) Next(e *Entry) bool {
	e.LowPC = rdr.oneAddr()
	e.HighPC = rdr.oneAddr()

	if e.LowPC == 0 && e.HighPC == 0 {
		return false
	}

	if e.BaseAddressSelection() {
		e.Instr = nil
		return true
	}

	instrlen := binary.LittleEndian.Uint16(rdr.read(2))
	e.Instr = rdr.read(int(instrlen))
	return true
}




func (rdr *Dwarf2Reader) Find(off int, staticBase, base, pc uint64, debugAddr *godwarf.DebugAddr) (*Entry, error) {
	rdr.Seek(off)
	var e Entry
	for rdr.Next(&e) {
		if e.BaseAddressSelection() {
			base = e.HighPC + staticBase
			continue
		}
		if pc >= e.LowPC+base && pc < e.HighPC+base {
			return &e, nil
		}
	}
	return nil, nil
}

func (rdr *Dwarf2Reader) read(sz int) []byte {
	r := rdr.data[rdr.cur : rdr.cur+sz]
	rdr.cur += sz
	return r
}

func (rdr *Dwarf2Reader) oneAddr() uint64 {
	switch rdr.ptrSz {
	case 4:
		addr := binary.LittleEndian.Uint32(rdr.read(rdr.ptrSz))
		if addr == ^uint32(0) {
			return ^uint64(0)
		}
		return uint64(addr)
	case 8:
		addr := uint64(binary.LittleEndian.Uint64(rdr.read(rdr.ptrSz)))
		return addr
	default:
		panic("bad address size")
	}
}


type Entry struct {
	LowPC, HighPC uint64
	Instr         []byte
}



func (e *Entry) BaseAddressSelection() bool {
	return e.LowPC == ^uint64(0)
}
