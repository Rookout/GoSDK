package assembler

import "github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"


const (
	ACALL = obj.ACALL
	AJMP  = obj.AJMP
	ANOP  = obj.ANOP
	ARET  = obj.ARET
)



func (b *Builder) BranchToLabel(op obj.As, dst string, args ...Arg) *Instruction {
	inst := b.NewInstruction()
	inst.As = op
	inst.To.Type = obj.TYPE_BRANCH
	inst.jumpDestLabel = dst
	if args != nil {
		inst.From = argToAddr(args[0])
	}
	return inst
}

func (b *Builder) Label(label string) *Instruction {
	inst := b.PsuedoNop()
	inst.label = label
	return inst
}

func (b *Builder) Bytes(bytes []byte) *Instruction {
	if len(bytes) == 0 {
		return b.PsuedoNop()
	}

	inst := b.NewInstruction()
	inst.placeholderFor = bytes
	inst.As = placeholderOp
	return inst
}



func (b *Builder) PsuedoNop() *Instruction {
	return b.Inst(ANOP, NoArg, NoArg)
}
