// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

import (
	"bytes"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/abi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"
	"io"
	"strings"
)

const REG_NONE = 0


func (p *Prog) Line() string {
	return p.Ctxt.OutermostPos(p.Pos).Format(false, true)
}
func (p *Prog) InnermostLine(w io.Writer) {
	p.Ctxt.InnermostPos(p.Pos).WriteTo(w, false, true)
}



func (p *Prog) InnermostLineNumber() string {
	return p.Ctxt.InnermostPos(p.Pos).LineNumber()
}



func (p *Prog) InnermostLineNumberHTML() string {
	return p.Ctxt.InnermostPos(p.Pos).LineNumberHTML()
}



func (p *Prog) InnermostFilename() string {
	
	
	pos := p.Ctxt.InnermostPos(p.Pos)
	if !pos.IsKnown() {
		return "<unknown file name>"
	}
	return pos.Filename()
}

var armCondCode = []string{
	".EQ",
	".NE",
	".CS",
	".CC",
	".MI",
	".PL",
	".VS",
	".VC",
	".HI",
	".LS",
	".GE",
	".LT",
	".GT",
	".LE",
	"",
	".NV",
}


const (
	C_SCOND     = (1 << 4) - 1
	C_SBIT      = 1 << 4
	C_PBIT      = 1 << 5
	C_WBIT      = 1 << 6
	C_FBIT      = 1 << 7
	C_UBIT      = 1 << 7
	C_SCOND_XOR = 14
)


func CConv(s uint8) string {
	if s == 0 {
		return ""
	}
	for i := range opSuffixSpace {
		sset := &opSuffixSpace[i]
		if sset.arch == buildcfg.GOARCH {
			return sset.cconv(s)
		}
	}
	return fmt.Sprintf("SC???%d", s)
}


func CConvARM(s uint8) string {
	
	
	

	sc := armCondCode[(s&C_SCOND)^C_SCOND_XOR]
	if s&C_SBIT != 0 {
		sc += ".S"
	}
	if s&C_PBIT != 0 {
		sc += ".P"
	}
	if s&C_WBIT != 0 {
		sc += ".W"
	}
	if s&C_UBIT != 0 { 
		sc += ".U"
	}
	return sc
}

func (p *Prog) String() string {
	if p == nil {
		return "<nil Prog>"
	}
	if p.Ctxt == nil {
		return "<Prog without ctxt>"
	}
	return fmt.Sprintf("%.5d (%v)\t%s", p.Pc, p.Line(), p.InstructionString())
}

func (p *Prog) InnermostString(w io.Writer) {
	if p == nil {
		io.WriteString(w, "<nil Prog>")
		return
	}
	if p.Ctxt == nil {
		io.WriteString(w, "<Prog without ctxt>")
		return
	}
	fmt.Fprintf(w, "%.5d (", p.Pc)
	p.InnermostLine(w)
	io.WriteString(w, ")\t")
	p.WriteInstructionString(w)
}



func (p *Prog) InstructionString() string {
	buf := new(bytes.Buffer)
	p.WriteInstructionString(buf)
	return buf.String()
}



func (p *Prog) WriteInstructionString(w io.Writer) {
	if p == nil {
		io.WriteString(w, "<nil Prog>")
		return
	}

	if p.Ctxt == nil {
		io.WriteString(w, "<Prog without ctxt>")
		return
	}

	sc := CConv(p.Scond)

	io.WriteString(w, p.As.String())
	io.WriteString(w, sc)
	sep := "\t"

	if p.From.Type != TYPE_NONE {
		io.WriteString(w, sep)
		WriteDconv(w, p, &p.From)
		sep = ", "
	}
	if p.Reg != REG_NONE {
		
		fmt.Fprintf(w, "%s%v", sep, Rconv(int(p.Reg)))
		sep = ", "
	}
	for i := range p.RestArgs {
		if p.RestArgs[i].Pos == Source {
			io.WriteString(w, sep)
			WriteDconv(w, p, &p.RestArgs[i].Addr)
			sep = ", "
		}
	}

	if p.As == ATEXT {
		
		
		
		
		s := p.From.Sym.TextAttrString()
		if s != "" {
			fmt.Fprintf(w, "%s%s", sep, s)
			sep = ", "
		}
	}
	if p.To.Type != TYPE_NONE {
		io.WriteString(w, sep)
		WriteDconv(w, p, &p.To)
		sep = ", "
	}
	if p.RegTo2 != REG_NONE {
		fmt.Fprintf(w, "%s%v", sep, Rconv(int(p.RegTo2)))
	}
	for i := range p.RestArgs {
		if p.RestArgs[i].Pos == Destination {
			io.WriteString(w, sep)
			WriteDconv(w, p, &p.RestArgs[i].Addr)
			sep = ", "
		}
	}
}

func (ctxt *Link) NewProg() *Prog {
	p := new(Prog)
	p.Ctxt = ctxt
	return p
}

func (ctxt *Link) CanReuseProgs() bool {
	return ctxt.Debugasm == 0
}



func Dconv(p *Prog, a *Addr) string {
	buf := new(bytes.Buffer)
	writeDconv(buf, p, a, false)
	return buf.String()
}




func DconvWithABIDetail(p *Prog, a *Addr) string {
	buf := new(bytes.Buffer)
	writeDconv(buf, p, a, true)
	return buf.String()
}



func WriteDconv(w io.Writer, p *Prog, a *Addr) {
	writeDconv(w, p, a, false)
}

func writeDconv(w io.Writer, p *Prog, a *Addr, abiDetail bool) {
	switch a.Type {
	default:
		fmt.Fprintf(w, "type=%d", a.Type)

	case TYPE_NONE:
		if a.Name != NAME_NONE || a.Reg != 0 || a.Sym != nil {
			a.WriteNameTo(w)
			fmt.Fprintf(w, "(%v)(NONE)", Rconv(int(a.Reg)))
		}

	case TYPE_REG:
		
		
		
		
		if a.Offset != 0 && (a.Reg < RBaseARM64 || a.Reg >= RBaseMIPS) {
			fmt.Fprintf(w, "$%d,%v", a.Offset, Rconv(int(a.Reg)))
			return
		}

		if a.Name != NAME_NONE || a.Sym != nil {
			a.WriteNameTo(w)
			fmt.Fprintf(w, "(%v)(REG)", Rconv(int(a.Reg)))
		} else {
			io.WriteString(w, Rconv(int(a.Reg)))
		}
		if (RBaseARM64+1<<10+1<<9)  <= a.Reg &&
			a.Reg < (RBaseARM64+1<<11)  {
			fmt.Fprintf(w, "[%d]", a.Index)
		}

	case TYPE_BRANCH:
		if a.Sym != nil {
			fmt.Fprintf(w, "%s%s(SB)", a.Sym.Name, abiDecorate(a, abiDetail))
		} else if a.Target() != nil {
			fmt.Fprint(w, a.Target().Pc)
		} else {
			fmt.Fprintf(w, "%d(PC)", a.Offset)
		}

	case TYPE_INDIR:
		io.WriteString(w, "*")
		a.writeNameTo(w, abiDetail)

	case TYPE_MEM:
		a.WriteNameTo(w)
		if a.Index != REG_NONE {
			if a.Scale == 0 {
				
				fmt.Fprintf(w, "(%v)", Rconv(int(a.Index)))
			} else {
				fmt.Fprintf(w, "(%v*%d)", Rconv(int(a.Index)), int(a.Scale))
			}
		}

	case TYPE_CONST:
		io.WriteString(w, "$")
		a.WriteNameTo(w)
		if a.Reg != 0 {
			fmt.Fprintf(w, "(%v)", Rconv(int(a.Reg)))
		}

	case TYPE_TEXTSIZE:
		if a.Val.(int32) == abi.ArgsSizeUnknown {
			fmt.Fprintf(w, "$%d", a.Offset)
		} else {
			fmt.Fprintf(w, "$%d-%d", a.Offset, a.Val.(int32))
		}

	case TYPE_FCONST:
		str := fmt.Sprintf("%.17g", a.Val.(float64))
		
		if !strings.ContainsAny(str, ".e") {
			str += ".0"
		}
		fmt.Fprintf(w, "$(%s)", str)

	case TYPE_SCONST:
		fmt.Fprintf(w, "$%q", a.Val.(string))

	case TYPE_ADDR:
		io.WriteString(w, "$")
		a.writeNameTo(w, abiDetail)

	case TYPE_SHIFT:
		v := int(a.Offset)
		ops := "<<>>->@>"
		switch buildcfg.GOARCH {
		case "arm":
			op := ops[((v>>5)&3)<<1:]
			if v&(1<<4) != 0 {
				fmt.Fprintf(w, "R%d%c%cR%d", v&15, op[0], op[1], (v>>8)&15)
			} else {
				fmt.Fprintf(w, "R%d%c%c%d", v&15, op[0], op[1], (v>>7)&31)
			}
			if a.Reg != 0 {
				fmt.Fprintf(w, "(%v)", Rconv(int(a.Reg)))
			}
		case "arm64":
			op := ops[((v>>22)&3)<<1:]
			r := (v >> 16) & 31
			fmt.Fprintf(w, "%s%c%c%d", Rconv(r+RBaseARM64), op[0], op[1], (v>>10)&63)
		default:
			panic("TYPE_SHIFT is not supported on " + buildcfg.GOARCH)
		}

	case TYPE_REGREG:
		fmt.Fprintf(w, "(%v, %v)", Rconv(int(a.Reg)), Rconv(int(a.Offset)))

	case TYPE_REGREG2:
		fmt.Fprintf(w, "%v, %v", Rconv(int(a.Offset)), Rconv(int(a.Reg)))

	case TYPE_REGLIST:
		io.WriteString(w, RLconv(a.Offset))

	case TYPE_SPECIAL:
		io.WriteString(w, SPCconv(a.Offset))
	}
}

func (a *Addr) WriteNameTo(w io.Writer) {
	a.writeNameTo(w, false)
}

func (a *Addr) writeNameTo(w io.Writer, abiDetail bool) {

	switch a.Name {
	default:
		fmt.Fprintf(w, "name=%d", a.Name)

	case NAME_NONE:
		switch {
		case a.Reg == REG_NONE:
			fmt.Fprint(w, a.Offset)
		case a.Offset == 0:
			fmt.Fprintf(w, "(%v)", Rconv(int(a.Reg)))
		case a.Offset != 0:
			fmt.Fprintf(w, "%d(%v)", a.Offset, Rconv(int(a.Reg)))
		}

		
	case NAME_EXTERN:
		reg := "SB"
		if a.Reg != REG_NONE {
			reg = Rconv(int(a.Reg))
		}
		if a.Sym != nil {
			fmt.Fprintf(w, "%s%s%s(%s)", a.Sym.Name, abiDecorate(a, abiDetail), offConv(a.Offset), reg)
		} else {
			fmt.Fprintf(w, "%s(%s)", offConv(a.Offset), reg)
		}

	case NAME_GOTREF:
		reg := "SB"
		if a.Reg != REG_NONE {
			reg = Rconv(int(a.Reg))
		}
		if a.Sym != nil {
			fmt.Fprintf(w, "%s%s@GOT(%s)", a.Sym.Name, offConv(a.Offset), reg)
		} else {
			fmt.Fprintf(w, "%s@GOT(%s)", offConv(a.Offset), reg)
		}

	case NAME_STATIC:
		reg := "SB"
		if a.Reg != REG_NONE {
			reg = Rconv(int(a.Reg))
		}
		if a.Sym != nil {
			fmt.Fprintf(w, "%s<>%s(%s)", a.Sym.Name, offConv(a.Offset), reg)
		} else {
			fmt.Fprintf(w, "<>%s(%s)", offConv(a.Offset), reg)
		}

	case NAME_AUTO:
		reg := "SP"
		if a.Reg != REG_NONE {
			reg = Rconv(int(a.Reg))
		}
		if a.Sym != nil {
			fmt.Fprintf(w, "%s%s(%s)", a.Sym.Name, offConv(a.Offset), reg)
		} else {
			fmt.Fprintf(w, "%s(%s)", offConv(a.Offset), reg)
		}

	case NAME_PARAM:
		reg := "FP"
		if a.Reg != REG_NONE {
			reg = Rconv(int(a.Reg))
		}
		if a.Sym != nil {
			fmt.Fprintf(w, "%s%s(%s)", a.Sym.Name, offConv(a.Offset), reg)
		} else {
			fmt.Fprintf(w, "%s(%s)", offConv(a.Offset), reg)
		}
	case NAME_TOCREF:
		reg := "SB"
		if a.Reg != REG_NONE {
			reg = Rconv(int(a.Reg))
		}
		if a.Sym != nil {
			fmt.Fprintf(w, "%s%s(%s)", a.Sym.Name, offConv(a.Offset), reg)
		} else {
			fmt.Fprintf(w, "%s(%s)", offConv(a.Offset), reg)
		}
	}
}

func offConv(off int64) string {
	if off == 0 {
		return ""
	}
	return fmt.Sprintf("%+d", off)
}







type opSuffixSet struct {
	arch  string
	cconv func(suffix uint8) string
}

var opSuffixSpace []opSuffixSet





func RegisterOpSuffix(arch string, cconv func(uint8) string) {
	opSuffixSpace = append(opSuffixSpace, opSuffixSet{
		arch:  arch,
		cconv: cconv,
	})
}

type regSet struct {
	lo    int
	hi    int
	Rconv func(int) string
}



var regSpace []regSet



const (
	
	
	RBase386     = 1 * 1024
	RBaseAMD64   = 2 * 1024
	RBaseARM     = 3 * 1024
	RBasePPC64   = 4 * 1024  
	RBaseARM64   = 8 * 1024  
	RBaseMIPS    = 13 * 1024 
	RBaseS390X   = 14 * 1024 
	RBaseRISCV   = 15 * 1024 
	RBaseWasm    = 16 * 1024
	RBaseLOONG64 = 17 * 1024
)




func RegisterRegister(lo, hi int, Rconv func(int) string) {
	regSpace = append(regSpace, regSet{lo, hi, Rconv})
}

func Rconv(reg int) string {
	if reg == REG_NONE {
		return "NONE"
	}
	for i := range regSpace {
		rs := &regSpace[i]
		if rs.lo <= reg && reg < rs.hi {
			return rs.Rconv(reg)
		}
	}
	return fmt.Sprintf("R???%d", reg)
}

type regListSet struct {
	lo     int64
	hi     int64
	RLconv func(int64) string
}

var regListSpace []regListSet



const (
	RegListARMLo = 0
	RegListARMHi = 1 << 16

	
	RegListARM64Lo = 1 << 60
	RegListARM64Hi = 1<<61 - 1

	
	RegListX86Lo = 1 << 61
	RegListX86Hi = 1<<62 - 1
)




func RegisterRegisterList(lo, hi int64, rlconv func(int64) string) {
	regListSpace = append(regListSpace, regListSet{lo, hi, rlconv})
}

func RLconv(list int64) string {
	for i := range regListSpace {
		rls := &regListSpace[i]
		if rls.lo <= list && list < rls.hi {
			return rls.RLconv(list)
		}
	}
	return fmt.Sprintf("RL???%d", list)
}


type spcSet struct {
	lo      int64
	hi      int64
	SPCconv func(int64) string
}

var spcSpace []spcSet




func RegisterSpecialOperands(lo, hi int64, rlconv func(int64) string) {
	spcSpace = append(spcSpace, spcSet{lo, hi, rlconv})
}


func SPCconv(spc int64) string {
	for i := range spcSpace {
		spcs := &spcSpace[i]
		if spcs.lo <= spc && spc < spcs.hi {
			return spcs.SPCconv(spc)
		}
	}
	return fmt.Sprintf("SPC???%d", spc)
}

type opSet struct {
	lo    As
	names []string
}


var aSpace []opSet



func RegisterOpcode(lo As, Anames []string) {
	if len(Anames) > AllowedOpCodes {
		panic(fmt.Sprintf("too many instructions, have %d max %d", len(Anames), AllowedOpCodes))
	}
	aSpace = append(aSpace, opSet{lo, Anames})
}

func (a As) String() string {
	if 0 <= a && int(a) < len(Anames) {
		return Anames[a]
	}
	for i := range aSpace {
		as := &aSpace[i]
		if as.lo <= a && int(a-as.lo) < len(as.names) {
			return as.names[a-as.lo]
		}
	}
	return fmt.Sprintf("A???%d", a)
}

var Anames = []string{
	"XXX",
	"CALL",
	"DUFFCOPY",
	"DUFFZERO",
	"END",
	"FUNCDATA",
	"JMP",
	"NOP",
	"PCALIGN",
	"PCDATA",
	"RET",
	"GETCALLERPC",
	"TEXT",
	"UNDEF",
}

func Bool2int(b bool) int {
	
	
	var i int
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

func abiDecorate(a *Addr, abiDetail bool) string {
	if !abiDetail || a.Sym == nil {
		return ""
	}
	return fmt.Sprintf("<%s>", a.Sym.ABI())
}
