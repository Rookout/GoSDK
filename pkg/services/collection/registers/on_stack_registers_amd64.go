//go:build amd64
// +build amd64

package registers

import (
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)

type OnStackRegisters struct {
	
	
	RDI    uintptr
	RDX    uintptr
	RBX    uintptr
	RAX    uintptr
	RCX    uintptr
	RSI    uintptr
	R8     uintptr
	R9     uintptr
	R10    uintptr
	R11    uintptr
	R12    uintptr
	R13    uintptr
	R14    uintptr
	R15    uintptr
	RBP    uintptr
	RIP    uintptr
	TLSVal uintptr
	RSP    uintptr
}

func NewOnStackRegisters(context uintptr) *OnStackRegisters {
	
	return &OnStackRegisters{
		RDI: getRegAtOffset(context, 0x8),
		RAX: getRegAtOffset(context, 0x10),
		RBX: getRegAtOffset(context, 0x18),
		RCX: getRegAtOffset(context, 0x20),
		RDX: getRegAtOffset(context, 0x28),
		RSI: getRegAtOffset(context, 0x30),
		R8:  getRegAtOffset(context, 0x38),
		R9:  getRegAtOffset(context, 0x40),
		R10: getRegAtOffset(context, 0x48),
		R11: getRegAtOffset(context, 0x50),
		R14: getRegAtOffset(context, 0x148),
		R15: getRegAtOffset(context, 0x160),
		RBP: getRegAtOffset(context, 0x168),
		RSP: getRegAtOffset(context, 0x170),
		R12: getRegAtOffset(context, 0x178),
		R13: getRegAtOffset(context, 0x180),
		RIP: getRegAtOffset(context, 0x188),
	}
}

func getRegAtOffset(context uintptr, offset uintptr) uintptr {
	addr := context + offset
	return *((*uintptr)(unsafe.Pointer(addr)))
}

func (o OnStackRegisters) PC() uint64 {
	return uint64(o.RIP)
}

func (o OnStackRegisters) SP() uint64 {
	return uint64(o.RSP)
}

func (o OnStackRegisters) BP() uint64 {
	return uint64(o.RBP)
}
func (o OnStackRegisters) TLS() uint64 {
	return uint64(o.TLSVal)
}


func (o OnStackRegisters) GAddr() (uint64, bool) {
	return 0, false
}
func (o OnStackRegisters) Get(int) (uint64, error) {
	panic("not implemented")
}
func (o OnStackRegisters) Slice(floatingPoint bool) ([]Register, error) {
	
	return []Register{
		{
			Name: "Rsp",
			Reg:  op.DwarfRegisterFromUint64(o.SP()),
		},
		{
			Name: "Rbp",
			Reg:  op.DwarfRegisterFromUint64(o.BP()),
		},
		{
			Name: "Rip",
			Reg:  op.DwarfRegisterFromUint64(o.PC()),
		},

		{
			Name: "Rax",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.RAX)),
		},
		{
			Name: "Rdx",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.RDX)),
		},
		{
			Name: "Rcx",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.RCX)),
		},
		{
			Name: "Rbx",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.RBX)),
		},
		{
			Name: "Rsi",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.RSI)),
		},
		{
			Name: "Rdi",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.RDI)),
		},

		{
			Name: "R8",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R8)),
		},
		{
			Name: "R9",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R9)),
		},
		{
			Name: "R10",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R10)),
		},
		{
			Name: "R11",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R11)),
		},
		{
			Name: "R12",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R12)),
		},
		{
			Name: "R13",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R13)),
		},
		{
			Name: "R14",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R14)),
		},
		{
			Name: "R15",
			Reg:  op.DwarfRegisterFromUint64(uint64(o.R15)),
		},
	}, nil
}



func (o OnStackRegisters) Copy() (Registers, error) {
	regCopy := o
	return regCopy, nil
}
