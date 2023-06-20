package assembler

type Mem struct {
	Arg
	Base Reg
	Disp int64
}
