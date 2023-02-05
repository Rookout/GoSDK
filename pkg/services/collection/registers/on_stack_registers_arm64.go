//go:build arm64
// +build arm64

package registers

import (
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)

type OnStackRegisters struct {
	x0  uintptr
	x1  uintptr
	x2  uintptr
	x3  uintptr
	x4  uintptr
	x5  uintptr
	x6  uintptr
	x7  uintptr
	x8  uintptr
	x9  uintptr
	x10 uintptr
	x11 uintptr
	x12 uintptr
	x13 uintptr
	x14 uintptr
	x15 uintptr
	x16 uintptr
	x17 uintptr
	x18 uintptr
	x19 uintptr
	x20 uintptr
	x21 uintptr
	x22 uintptr
	x23 uintptr
	x24 uintptr
	x25 uintptr
	x26 uintptr
	x27 uintptr
	x28 uintptr
	x29 uintptr 
	x30 uintptr 
	pc  uintptr
	sp  uintptr
}

func NewOnStackRegisters(context uintptr) *OnStackRegisters {
	
	return &OnStackRegisters{
		x0:  getRegAtOffset(context, 0x8),
		x1:  getRegAtOffset(context, 0x10),
		x2:  getRegAtOffset(context, 0x18),
		x3:  getRegAtOffset(context, 0x20),
		x4:  getRegAtOffset(context, 0x28),
		x5:  getRegAtOffset(context, 0x30),
		x6:  getRegAtOffset(context, 0x38),
		x7:  getRegAtOffset(context, 0x40),
		x8:  getRegAtOffset(context, 0x48),
		x9:  getRegAtOffset(context, 0x50),
		x10: getRegAtOffset(context, 0x58),
		x11: getRegAtOffset(context, 0x60),
		x12: getRegAtOffset(context, 0x68),
		x13: getRegAtOffset(context, 0x70),
		x14: getRegAtOffset(context, 0x78),
		x15: getRegAtOffset(context, 0x80),

		x16: getRegAtOffset(context, 0x1a0),
		x17: getRegAtOffset(context, 0x1a8),
		x18: getRegAtOffset(context, 0x1b0),
		x21: getRegAtOffset(context, 0x1b8),
		x22: getRegAtOffset(context, 0x1c0),
		x23: getRegAtOffset(context, 0x1c8),
		x24: getRegAtOffset(context, 0x1d0),
		x25: getRegAtOffset(context, 0x1d8),
		x26: getRegAtOffset(context, 0x1e0),
		x27: getRegAtOffset(context, 0x1e8),
		x28: getRegAtOffset(context, 0x1f0),

		x19: getRegAtOffset(context, 0x1f0+0x120),
		x20: getRegAtOffset(context, 0x1f0+0x128),
		x29: getRegAtOffset(context, 0x1f0+0x130), 
		x30: getRegAtOffset(context, 0x1f0+0x138), 
		sp:  getRegAtOffset(context, 0x1f0+0x140),
		pc:  getRegAtOffset(context, 0x1f0+0x148),
	}
}

func getRegAtOffset(context uintptr, offset uintptr) uintptr {
	addr := context + offset
	return *((*uintptr)(unsafe.Pointer(addr)))
}

func (o OnStackRegisters) PC() uint64 {
	return uint64(o.pc)
}

func (o OnStackRegisters) SP() uint64 {
	return uint64(o.sp)
}

func (o OnStackRegisters) BP() uint64 {
	return uint64(o.x29)
}
func (o OnStackRegisters) TLS() uint64 {
	return 0
}


func (o OnStackRegisters) GAddr() (uint64, bool) {
	return uint64(o.x28), true 
}
func (o OnStackRegisters) Get(int) (uint64, error) {
	panic("not implemented")
}
func (o OnStackRegisters) Slice(floatingPoint bool) ([]Register, error) {
	
	return []Register{
		{
			Name: "x0",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x0)),
		},
		{
			Name: "x1",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x1)),
		},
		{
			Name: "x2",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x2)),
		},
		{
			Name: "x3",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x3)),
		},
		{
			Name: "x4",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x4)),
		},
		{
			Name: "x5",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x5)),
		},
		{
			Name: "x6",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x6)),
		},
		{
			Name: "x7",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x7)),
		},
		{
			Name: "x8",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x8)),
		},
		{
			Name: "x9",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x9)),
		},
		{
			Name: "x10",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x10)),
		},
		{
			Name: "x11",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x11)),
		},
		{
			Name: "x12",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x12)),
		},
		{
			Name: "x13",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x13)),
		},
		{
			Name: "x14",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x14)),
		},
		{
			Name: "x15",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x15)),
		},
		{
			Name: "x16",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x16)),
		},
		{
			Name: "x17",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x17)),
		},
		{
			Name: "x18",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x18)),
		},
		{
			Name: "x19",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x19)),
		},
		{
			Name: "x20",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x20)),
		},
		{
			Name: "x21",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x21)),
		},
		{
			Name: "x22",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x22)),
		},
		{
			Name: "x23",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x23)),
		},
		{
			Name: "x24",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x24)),
		},
		{
			Name: "x25",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x25)),
		},
		{
			Name: "x26",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x26)),
		},
		{
			Name: "x27",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x27)),
		},
		{
			Name: "x28",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x28)),
		},
		{
			Name: "x29",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x29)),
		},
		{
			Name: "x30",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x30)),
		},
		{
			Name: "pc",
			Reg:  op.DwarfRegisterFromUint64(o.PC()),
		},
		{
			Name: "sp",
			Reg:  op.DwarfRegisterFromUint64(o.SP()),
		},
		{
			Name: "lr",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.x30)), 
		},
	}, nil
}



func (o OnStackRegisters) Copy() (Registers, error) {
	regCopy := o
	return regCopy, nil
}
