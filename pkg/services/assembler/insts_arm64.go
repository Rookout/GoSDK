package assembler

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/arm64"
)

type args struct {
	Dst1   Arg
	Dst2   Arg
	Src1   Arg
	Src2   Arg
	SrcReg *Reg
	DstReg *Reg
}


func (b *Builder) instArgs(op obj.As, args args) *Instruction {
	inst := b.NewInstruction()
	inst.As = op
	if args.Dst1 != nil {
		inst.To = argToAddr(args.Dst1)
	}
	if args.Dst2 != nil {
		inst.RestArgs = append(inst.RestArgs, obj.AddrPos{Addr: argToAddr(args.Dst2), Pos: obj.Destination})
	}
	if args.DstReg != nil {
		inst.RegTo2 = AsmRegToSysReg(*args.DstReg)
	}
	if args.Src1 != nil {
		inst.From = argToAddr(args.Src1)
	}
	if args.Src2 != nil {
		inst.RestArgs = append(inst.RestArgs, obj.AddrPos{Addr: argToAddr(args.Src2), Pos: obj.Source})
	}
	if args.SrcReg != nil {
		inst.Reg = AsmRegToSysReg(*args.SrcReg)
	}
	return inst
}




func (b *Builder) Inst(op obj.As, dst, src Arg, cond ...uint8) *Instruction {
	inst := b.NewInstruction()
	inst.As = op
	inst.To = argToAddr(dst)
	inst.From = argToAddr(src)
	if len(cond) != 0 {
		inst.Scond = cond[0]
	}
	return inst
}




func (b *Builder) Cmp(arg1 Reg, arg2 Arg) *Instruction {
	return b.instArgs(ACMP, args{
		SrcReg: &arg1,
		Src1:   arg2,
	})
}





func (b *Builder) Sub3(dst Arg, src1 Reg, src2 Arg) *Instruction {
	return b.instArgs(arm64.ASUB, args{
		Dst1:   dst,
		SrcReg: &src1,
		Src1:   src2,
	})
}




func (b *Builder) Swpal(arg1 Arg, arg2 Reg, arg3 Arg) *Instruction {
	return b.instArgs(arm64.ASWPALD, args{
		Src1:   arg1,
		Dst1:   arg3,
		DstReg: &arg2,
	})
}




func (b *Builder) BranchToReg(op obj.As, dst Reg) *Instruction {
	var dstArg Arg = dst
	
	if op == AJMP {
		dstArg = Mem{Base: dst}
	}
	return b.instArgs(op, args{
		Dst1: dstArg,
	})
}
