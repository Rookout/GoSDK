//go:build amd64
// +build amd64

package registers

import (
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
	RBP    uintptr
	RIP    uintptr
	TLSVal uintptr
	RSP    uintptr
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
	}, nil
}



func (o OnStackRegisters) Copy() (Registers, error) {
	regCopy := o
	return regCopy, nil
}
