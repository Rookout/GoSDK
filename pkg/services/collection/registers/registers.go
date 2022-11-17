package registers

import "github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"





type Registers interface {
	PC() uint64
	SP() uint64
	BP() uint64
	TLS() uint64
	
	GAddr() (uint64, bool)
	Get(int) (uint64, error)
	Slice(floatingPoint bool) ([]Register, error)
	
	
	Copy() (Registers, error)
}


type Register struct {
	Name string
	Reg  *op.DwarfRegister
}



func AppendUint64Register(regs []Register, name string, value uint64) []Register {
	return append(regs, Register{name, op.DwarfRegisterFromUint64(value)})
}



func AppendBytesRegister(regs []Register, name string, value []byte) []Register {
	return append(regs, Register{name, op.DwarfRegisterFromBytes(value)})
}
