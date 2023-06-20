package assembler

import (
	"fmt"
	"runtime"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/asm/arch"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
)

var a, ctxt = func() (*arch.Arch, *obj.Link) {
	a := arch.Set(runtime.GOARCH, false)
	ctxt := obj.Linknew(a.LinkArch)
	ctxt.DiagFunc = func(in string, args ...interface{}) {
		panic(fmt.Sprintf(in, args...))
	}
	a.Init(ctxt)
	return a, ctxt
}()


type Builder struct {
	arch *arch.Arch

	first *Instruction
	last  *Instruction

	
	labels map[string]*Instruction
}

type Instruction struct {
	*obj.Prog
	label          string
	jumpDestLabel  string
	link           *Instruction
	placeholderFor []byte
}


func (b *Builder) NewInstruction() *Instruction {
	return &Instruction{Prog: ctxt.NewProg()}
}

func (b *Builder) insertInstruction(inst *Instruction) {
	if b.first == nil {
		b.first = inst
		b.last = inst
	} else {
		b.last.link = inst
		b.last = inst
	}
}



func (b *Builder) AddInstructions(p *Instruction, instructions ...*Instruction) rookoutErrors.RookoutError {
	insts := []*Instruction{p}
	insts = append(insts, instructions...)

	for _, inst := range insts {
		if inst.label != "" {
			if _, ok := b.labels[inst.label]; ok {
				return rookoutErrors.NewLabelAlreadyExists(inst.label)
			}

			b.labels[inst.label] = inst
		}

		b.insertInstruction(inst)
	}

	return nil
}

func (b *Builder) fixup() rookoutErrors.RookoutError {
	for inst := b.first; inst != nil; inst = inst.link {
		if inst.link != nil {
			inst.Prog.Link = inst.link.Prog
		}

		if inst.placeholderFor != nil {
			if b.arch.Arch.Name == "arm64" && len(inst.placeholderFor)%4 != 0 {
				return rookoutErrors.NewInvalidBytes(inst.placeholderFor)
			}

			progAfterPlaceholder := inst.Prog.Link
			curProg := inst.Prog
			
			for i := 0; i < len(inst.placeholderFor)-placeholderLen; i += placeholderLen {
				curProg.Link = b.placeholderProg()
				curProg = curProg.Link
			}
			curProg.Link = progAfterPlaceholder
			continue
		}

		if inst.jumpDestLabel != "" {
			dest, ok := b.labels[inst.jumpDestLabel]
			if !ok {
				return rookoutErrors.NewInvalidJumpDest(inst.jumpDestLabel)
			}

			inst.To.Val = dest.Prog
		}
	}

	return nil
}

func (b *Builder) replacePlaceholders(assembled []byte) {
	for inst := b.first; inst != nil; inst = inst.link {
		if inst.placeholderFor != nil {
			copy(assembled[inst.Pc:], inst.placeholderFor)
		}
	}
}


func (b *Builder) Assemble() ([]byte, rookoutErrors.RookoutError) {
	err := b.fixup()
	if err != nil {
		return nil, err
	}

	var funcInfo interface{} = &obj.FuncInfo{
		Text: b.first.Prog,
	}
	s := &obj.LSym{
		Extra: &funcInfo,
	}

	err = func() (err rookoutErrors.RookoutError) {
		defer func() {
			if r := recover(); r != nil {
				err = rookoutErrors.NewFailedToAssemble(r)
			}
		}()

		b.arch.Assemble(ctxt, s, ctxt.NewProg)
		return nil
	}()
	if err != nil {
		return nil, err
	}

	assembled := s.P

	
	if b.arch.Arch.Name == "arm64" {
		for string(assembled[len(assembled)-4:]) == "\x00\x00\x00\x00" {
			assembled = assembled[:len(assembled)-4]
		}
	}

	b.replacePlaceholders(assembled)
	return assembled, nil
}


func NewBuilder() *Builder {
	builder := &Builder{
		arch:   a,
		labels: make(map[string]*Instruction),
	}

	
	if runtime.GOARCH == "arm64" {
		builder.AddInstructions(
			builder.PsuedoNop(),
		)
	}
	return builder
}

func (b *Builder) placeholderProg() *obj.Prog {
	return b.Inst(placeholderOp, NoArg, NoArg).Prog
}

func (i *Instruction) String() string {
	if i.label != "" {
		return i.label + ":"
	}
	s := i.Prog.String()
	if i.jumpDestLabel != "" {
		return s + " " + i.jumpDestLabel
	}
	return s
}
