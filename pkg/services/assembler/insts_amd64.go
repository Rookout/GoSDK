package assembler

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
)



func (b *Builder) Inst(op obj.As, dst, src Arg) *Instruction {
	inst := b.NewInstruction()
	inst.As = op
	inst.To = argToAddr(dst)
	inst.From = argToAddr(src)
	return inst
}



func (b *Builder) Cmp(dst, src Arg) *Instruction {
	inst := b.NewInstruction()
	inst.As = ACMPQ
	
	inst.To = argToAddr(src)
	inst.From = argToAddr(dst)
	return inst
}




func (b *Builder) BranchToReg(op obj.As, dst Arg) *Instruction {
	return b.Inst(op, dst, NoArg)
}
