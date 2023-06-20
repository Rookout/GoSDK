// cmd/7l/noop.c, cmd/7l/obj.c, cmd/ld/pass.c from Vita Nuova.
// https://code.google.com/p/ken-cc/source/browse/
//
// 	Copyright © 1994-1999 Lucent Technologies Inc. All rights reserved.
// 	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
// 	Portions Copyright © 1997-1999 Vita Nuova Limited
// 	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
// 	Portions Copyright © 2004,2006 Bruce Ellis
// 	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
// 	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
// 	Portions Copyright © 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package arm64

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/src"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/sys"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/abi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"
	"log"
	"math"
)



var zrReplace = map[obj.As]bool{
	AMOVD:  true,
	AMOVW:  true,
	AMOVWU: true,
	AMOVH:  true,
	AMOVHU: true,
	AMOVB:  true,
	AMOVBU: true,
	ASBC:   true,
	ASBCW:  true,
	ASBCS:  true,
	ASBCSW: true,
	AADC:   true,
	AADCW:  true,
	AADCS:  true,
	AADCSW: true,
	AFMOVD: true,
	AFMOVS: true,
	AMSR:   true,
}

func (c *ctxt7) stacksplit(p *obj.Prog, framesize int32) *obj.Prog {
	if c.ctxt.Flag_maymorestack != "" {
		p = c.cursym.Func().SpillRegisterArgs(p, c.newprog)

		
		
		const frameSize = 32
		p = obj.Appendp(p, c.newprog)
		p.As = AMOVD
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGLINK
		p.To.Type = obj.TYPE_MEM
		p.Scond = C_XPRE
		p.To.Offset = -frameSize
		p.To.Reg = REGSP
		p.Spadj = frameSize

		
		p = obj.Appendp(p, c.newprog)
		p.As = AMOVD
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGFP
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = REGSP
		p.To.Offset = -8

		p = obj.Appendp(p, c.newprog)
		p.As = ASUB
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = 8
		p.Reg = REGSP
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGFP

		
		
		p = obj.Appendp(p, c.newprog)
		p.As = AMOVD
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGCTXT
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = REGSP
		p.To.Offset = 8

		
		p = obj.Appendp(p, c.newprog)
		p.As = ABL
		p.To.Type = obj.TYPE_BRANCH
		
		p.To.Sym = c.ctxt.LookupABI(c.ctxt.Flag_maymorestack, c.cursym.ABI())

		
		p = obj.Appendp(p, c.newprog)
		p.As = AMOVD
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = REGSP
		p.From.Offset = 8
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGCTXT

		
		p = obj.Appendp(p, c.newprog)
		p.As = AMOVD
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = REGSP
		p.From.Offset = -8
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGFP

		
		p = obj.Appendp(p, c.newprog)
		p.As = AMOVD
		p.From.Type = obj.TYPE_MEM
		p.Scond = C_XPOST
		p.From.Offset = frameSize
		p.From.Reg = REGSP
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGLINK
		p.Spadj = -frameSize

		p = c.cursym.Func().UnspillRegisterArgs(p, c.newprog)
	}

	
	startPred := p

	
	p = obj.Appendp(p, c.newprog)

	p.As = AMOVD
	p.From.Type = obj.TYPE_MEM
	p.From.Reg = REGG
	p.From.Offset = 2 * int64(c.ctxt.Arch.PtrSize) 
	if c.cursym.CFunc() {
		p.From.Offset = 3 * int64(c.ctxt.Arch.PtrSize) 
	}
	p.To.Type = obj.TYPE_REG
	p.To.Reg = REGRT1

	
	
	
	
	p = c.ctxt.StartUnsafePoint(p, c.newprog)

	q := (*obj.Prog)(nil)
	if framesize <= abi.StackSmall {
		
		

		p = obj.Appendp(p, c.newprog)
		p.As = ACMP
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGRT1
		p.Reg = REGSP
	} else if framesize <= abi.StackBig {
		
		
		
		p = obj.Appendp(p, c.newprog)

		p.As = ASUB
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(framesize) - abi.StackSmall
		p.Reg = REGSP
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGRT2

		p = obj.Appendp(p, c.newprog)
		p.As = ACMP
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGRT1
		p.Reg = REGRT2
	} else {
		
		
		
		
		
		
		
		
		
		
		

		p = obj.Appendp(p, c.newprog)
		p.As = ASUBS
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(framesize) - abi.StackSmall
		p.Reg = REGSP
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGRT2

		p = obj.Appendp(p, c.newprog)
		q = p
		p.As = ABLO
		p.To.Type = obj.TYPE_BRANCH

		p = obj.Appendp(p, c.newprog)
		p.As = ACMP
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGRT1
		p.Reg = REGRT2
	}

	
	bls := obj.Appendp(p, c.newprog)
	bls.As = ABLS
	bls.To.Type = obj.TYPE_BRANCH

	end := c.ctxt.EndUnsafePoint(bls, c.newprog, -1)

	var last *obj.Prog
	for last = c.cursym.Func().Text; last.Link != nil; last = last.Link {
	}

	
	
	
	spfix := obj.Appendp(last, c.newprog)
	spfix.As = obj.ANOP
	spfix.Spadj = -framesize

	pcdata := c.ctxt.EmitEntryStackMap(c.cursym, spfix, c.newprog)
	pcdata = c.ctxt.StartUnsafePoint(pcdata, c.newprog)

	if q != nil {
		q.To.SetTarget(pcdata)
	}
	bls.To.SetTarget(pcdata)

	spill := c.cursym.Func().SpillRegisterArgs(pcdata, c.newprog)

	
	movlr := obj.Appendp(spill, c.newprog)
	movlr.As = AMOVD
	movlr.From.Type = obj.TYPE_REG
	movlr.From.Reg = REGLINK
	movlr.To.Type = obj.TYPE_REG
	movlr.To.Reg = REG_R3

	debug := movlr
	if false {
		debug = obj.Appendp(debug, c.newprog)
		debug.As = AMOVD
		debug.From.Type = obj.TYPE_CONST
		debug.From.Offset = int64(framesize)
		debug.To.Type = obj.TYPE_REG
		debug.To.Reg = REGTMP
	}

	
	call := obj.Appendp(debug, c.newprog)
	call.As = ABL
	call.To.Type = obj.TYPE_BRANCH
	morestack := "runtime.morestack"
	switch {
	case c.cursym.CFunc():
		morestack = "runtime.morestackc"
	case !c.cursym.Func().Text.From.Sym.NeedCtxt():
		morestack = "runtime.morestack_noctxt"
	}
	call.To.Sym = c.ctxt.Lookup(morestack)

	unspill := c.cursym.Func().UnspillRegisterArgs(call, c.newprog)
	pcdata = c.ctxt.EndUnsafePoint(unspill, c.newprog, -1)

	
	jmp := obj.Appendp(pcdata, c.newprog)
	jmp.As = AB
	jmp.To.Type = obj.TYPE_BRANCH
	jmp.To.SetTarget(startPred.Link)
	jmp.Spadj = +framesize

	return end
}

func progedit(ctxt *obj.Link, p *obj.Prog, newprog obj.ProgAlloc) {
	c := ctxt7{ctxt: ctxt, newprog: newprog}

	p.From.Class = 0
	p.To.Class = 0

	
	
	
	if p.From.Type == obj.TYPE_CONST && p.From.Offset == 0 && zrReplace[p.As] {
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REGZERO
	}

	
	switch p.As {
	case AB,
		ABL,
		obj.ARET,
		obj.ADUFFZERO,
		obj.ADUFFCOPY:
		if p.To.Sym != nil {
			p.To.Type = obj.TYPE_BRANCH
		}
		break
	}

	
	switch p.As {
	case AFMOVS:
		if p.From.Type == obj.TYPE_FCONST {
			f64 := p.From.Val.(float64)
			f32 := float32(f64)
			if c.chipfloat7(f64) > 0 {
				break
			}
			if math.Float32bits(f32) == 0 {
				p.From.Type = obj.TYPE_REG
				p.From.Reg = REGZERO
				break
			}
			p.From.Type = obj.TYPE_MEM
			p.From.Sym = c.ctxt.Float32Sym(f32)
			p.From.Name = obj.NAME_EXTERN
			p.From.Offset = 0
		}

	case AFMOVD:
		if p.From.Type == obj.TYPE_FCONST {
			f64 := p.From.Val.(float64)
			if c.chipfloat7(f64) > 0 {
				break
			}
			if math.Float64bits(f64) == 0 {
				p.From.Type = obj.TYPE_REG
				p.From.Reg = REGZERO
				break
			}
			p.From.Type = obj.TYPE_MEM
			p.From.Sym = c.ctxt.Float64Sym(f64)
			p.From.Name = obj.NAME_EXTERN
			p.From.Offset = 0
		}

		break
	}

	if c.ctxt.Flag_dynlink {
		c.rewriteToUseGot(p)
	}
}


func (c *ctxt7) rewriteToUseGot(p *obj.Prog) {
	if p.As == obj.ADUFFCOPY || p.As == obj.ADUFFZERO {
		
		
		
		
		
		var sym *obj.LSym
		if p.As == obj.ADUFFZERO {
			sym = c.ctxt.LookupABI("runtime.duffzero", obj.ABIInternal)
		} else {
			sym = c.ctxt.LookupABI("runtime.duffcopy", obj.ABIInternal)
		}
		offset := p.To.Offset
		p.As = AMOVD
		p.From.Type = obj.TYPE_MEM
		p.From.Name = obj.NAME_GOTREF
		p.From.Sym = sym
		p.To.Type = obj.TYPE_REG
		p.To.Reg = REGTMP
		p.To.Name = obj.NAME_NONE
		p.To.Offset = 0
		p.To.Sym = nil
		p1 := obj.Appendp(p, c.newprog)
		p1.As = AADD
		p1.From.Type = obj.TYPE_CONST
		p1.From.Offset = offset
		p1.To.Type = obj.TYPE_REG
		p1.To.Reg = REGTMP
		p2 := obj.Appendp(p1, c.newprog)
		p2.As = obj.ACALL
		p2.To.Type = obj.TYPE_REG
		p2.To.Reg = REGTMP
	}

	
	
	
	if p.From.Type == obj.TYPE_ADDR && p.From.Name == obj.NAME_EXTERN && !p.From.Sym.Local() {
		
		
		if p.As != AMOVD {
			c.ctxt.Diag("do not know how to handle TYPE_ADDR in %v with -dynlink", p)
		}
		if p.To.Type != obj.TYPE_REG {
			c.ctxt.Diag("do not know how to handle LEAQ-type insn to non-register in %v with -dynlink", p)
		}
		p.From.Type = obj.TYPE_MEM
		p.From.Name = obj.NAME_GOTREF
		if p.From.Offset != 0 {
			q := obj.Appendp(p, c.newprog)
			q.As = AADD
			q.From.Type = obj.TYPE_CONST
			q.From.Offset = p.From.Offset
			q.To = p.To
			p.From.Offset = 0
		}
	}
	if p.GetFrom3() != nil && p.GetFrom3().Name == obj.NAME_EXTERN {
		c.ctxt.Diag("don't know how to handle %v with -dynlink", p)
	}
	var source *obj.Addr
	
	
	
	if p.From.Name == obj.NAME_EXTERN && !p.From.Sym.Local() {
		if p.To.Name == obj.NAME_EXTERN && !p.To.Sym.Local() {
			c.ctxt.Diag("cannot handle NAME_EXTERN on both sides in %v with -dynlink", p)
		}
		source = &p.From
	} else if p.To.Name == obj.NAME_EXTERN && !p.To.Sym.Local() {
		source = &p.To
	} else {
		return
	}
	if p.As == obj.ATEXT || p.As == obj.AFUNCDATA || p.As == obj.ACALL || p.As == obj.ARET || p.As == obj.AJMP {
		return
	}
	if source.Sym.Type == objabi.STLSBSS {
		return
	}
	if source.Type != obj.TYPE_MEM {
		c.ctxt.Diag("don't know how to handle %v with -dynlink", p)
	}
	p1 := obj.Appendp(p, c.newprog)
	p2 := obj.Appendp(p1, c.newprog)
	p1.As = AMOVD
	p1.From.Type = obj.TYPE_MEM
	p1.From.Sym = source.Sym
	p1.From.Name = obj.NAME_GOTREF
	p1.To.Type = obj.TYPE_REG
	p1.To.Reg = REGTMP

	p2.As = p.As
	p2.From = p.From
	p2.To = p.To
	if p.From.Name == obj.NAME_EXTERN {
		p2.From.Reg = REGTMP
		p2.From.Name = obj.NAME_NONE
		p2.From.Sym = nil
	} else if p.To.Name == obj.NAME_EXTERN {
		p2.To.Reg = REGTMP
		p2.To.Name = obj.NAME_NONE
		p2.To.Sym = nil
	} else {
		return
	}
	obj.Nopout(p)
}

func preprocess(ctxt *obj.Link, cursym *obj.LSym, newprog obj.ProgAlloc) {
	if cursym.Func().Text == nil || cursym.Func().Text.Link == nil {
		return
	}

	c := ctxt7{ctxt: ctxt, newprog: newprog, cursym: cursym}

	p := c.cursym.Func().Text
	textstksiz := p.To.Offset
	if textstksiz == -8 {
		
		p.From.Sym.Set(obj.AttrNoFrame, true)
		textstksiz = 0
	}
	if textstksiz < 0 {
		c.ctxt.Diag("negative frame size %d - did you mean NOFRAME?", textstksiz)
	}
	if p.From.Sym.NoFrame() {
		if textstksiz != 0 {
			c.ctxt.Diag("NOFRAME functions must have a frame size of 0, not %d", textstksiz)
		}
	}

	c.cursym.Func().Args = p.To.Val.(int32)
	c.cursym.Func().Locals = int32(textstksiz)

	
	for p := c.cursym.Func().Text; p != nil; p = p.Link {
		switch p.As {
		case obj.ATEXT:
			p.Mark |= LEAF

		case ABL,
			obj.ADUFFZERO,
			obj.ADUFFCOPY:
			c.cursym.Func().Text.Mark &^= LEAF
		}
	}

	var q *obj.Prog
	var q1 *obj.Prog
	var retjmp *obj.LSym
	for p := c.cursym.Func().Text; p != nil; p = p.Link {
		o := p.As
		switch o {
		case obj.ATEXT:
			c.cursym.Func().Text = p
			c.autosize = int32(textstksiz)

			if p.Mark&LEAF != 0 && c.autosize == 0 {
				
				p.From.Sym.Set(obj.AttrNoFrame, true)
			}

			if !p.From.Sym.NoFrame() {
				
				
				c.autosize += 8
			}

			if c.autosize != 0 {
				extrasize := int32(0)
				if c.autosize%16 == 8 {
					
					extrasize = 8
				} else if c.autosize&(16-1) == 0 {
					
					extrasize = 16
				} else {
					c.ctxt.Diag("%v: unaligned frame size %d - must be 16 aligned", p, c.autosize-8)
				}
				c.autosize += extrasize
				c.cursym.Func().Locals += extrasize

				
				
				p.To.Offset = int64(c.autosize) | int64(extrasize)<<32
			} else {
				
				p.To.Offset = 0
			}

			if c.autosize == 0 && c.cursym.Func().Text.Mark&LEAF == 0 {
				if c.ctxt.Debugvlog {
					c.ctxt.Logf("save suppressed in: %s\n", c.cursym.Func().Text.From.Sym.Name)
				}
				c.cursym.Func().Text.Mark |= LEAF
			}

			if cursym.Func().Text.Mark&LEAF != 0 {
				cursym.Set(obj.AttrLeaf, true)
				if p.From.Sym.NoFrame() {
					break
				}
			}

			if p.Mark&LEAF != 0 && c.autosize < abi.StackSmall {
				
				
				p.From.Sym.Set(obj.AttrNoSplit, true)
			}

			if !p.From.Sym.NoSplit() {
				p = c.stacksplit(p, c.autosize) 
			}

			var prologueEnd *obj.Prog

			aoffset := c.autosize
			if aoffset > 0xf0 {
				
				
				aoffset = 0xf0
			}

			
			
			q = p
			if c.autosize > aoffset {
				
				
				
				

				
				q1 = obj.Appendp(q, c.newprog)
				q1.Pos = p.Pos
				q1.As = ASUB
				q1.From.Type = obj.TYPE_CONST
				q1.From.Offset = int64(c.autosize)
				q1.Reg = REGSP
				q1.To.Type = obj.TYPE_REG
				q1.To.Reg = REG_R20

				prologueEnd = q1

				
				q1 = obj.Appendp(q1, c.newprog)
				q1.Pos = p.Pos
				q1.As = ASTP
				q1.From.Type = obj.TYPE_REGREG
				q1.From.Reg = REGFP
				q1.From.Offset = REGLINK
				q1.To.Type = obj.TYPE_MEM
				q1.To.Reg = REG_R20
				q1.To.Offset = -8

				
				
				q1 = c.ctxt.StartUnsafePoint(q1, c.newprog)

				
				q1 = obj.Appendp(q1, c.newprog)
				q1.Pos = p.Pos
				q1.As = AMOVD
				q1.From.Type = obj.TYPE_REG
				q1.From.Reg = REG_R20
				q1.To.Type = obj.TYPE_REG
				q1.To.Reg = REGSP
				q1.Spadj = c.autosize

				q1 = c.ctxt.EndUnsafePoint(q1, c.newprog, -1)

				if buildcfg.GOOS == "ios" {
					
					
					
					
					q1 = obj.Appendp(q1, c.newprog)
					q1.Pos = p.Pos
					q1.As = ASTP
					q1.From.Type = obj.TYPE_REGREG
					q1.From.Reg = REGFP
					q1.From.Offset = REGLINK
					q1.To.Type = obj.TYPE_MEM
					q1.To.Reg = REGSP
					q1.To.Offset = -8
				}
			} else {
				
				
				
				
				
				
				
				
				
				q1 = obj.Appendp(q, c.newprog)
				q1.As = AMOVD
				q1.Pos = p.Pos
				q1.From.Type = obj.TYPE_REG
				q1.From.Reg = REGLINK
				q1.To.Type = obj.TYPE_MEM
				q1.Scond = C_XPRE
				q1.To.Offset = int64(-aoffset)
				q1.To.Reg = REGSP
				q1.Spadj = aoffset

				prologueEnd = q1

				
				q1 = obj.Appendp(q1, c.newprog)
				q1.Pos = p.Pos
				q1.As = AMOVD
				q1.From.Type = obj.TYPE_REG
				q1.From.Reg = REGFP
				q1.To.Type = obj.TYPE_MEM
				q1.To.Reg = REGSP
				q1.To.Offset = -8
			}

			prologueEnd.Pos = prologueEnd.Pos.WithXlogue(src.PosPrologueEnd)

			q1 = obj.Appendp(q1, c.newprog)
			q1.Pos = p.Pos
			q1.As = ASUB
			q1.From.Type = obj.TYPE_CONST
			q1.From.Offset = 8
			q1.Reg = REGSP
			q1.To.Type = obj.TYPE_REG
			q1.To.Reg = REGFP

			if c.cursym.Func().Text.From.Sym.Wrapper() {
				
				
				
				
				
				
				
				
				
				
				
				
				
				
				
				
				
				
				q = q1

				
				q = obj.Appendp(q, c.newprog)
				q.As = AMOVD
				q.From.Type = obj.TYPE_MEM
				q.From.Reg = REGG
				q.From.Offset = 4 * int64(c.ctxt.Arch.PtrSize) 
				q.To.Type = obj.TYPE_REG
				q.To.Reg = REGRT1

				
				cbnz := obj.Appendp(q, c.newprog)
				cbnz.As = ACBNZ
				cbnz.From.Type = obj.TYPE_REG
				cbnz.From.Reg = REGRT1
				cbnz.To.Type = obj.TYPE_BRANCH

				
				end := obj.Appendp(cbnz, c.newprog)
				end.As = obj.ANOP

				
				var last *obj.Prog
				for last = end; last.Link != nil; last = last.Link {
				}

				
				mov := obj.Appendp(last, c.newprog)
				mov.As = AMOVD
				mov.From.Type = obj.TYPE_MEM
				mov.From.Reg = REGRT1
				mov.From.Offset = 0 
				mov.To.Type = obj.TYPE_REG
				mov.To.Reg = REGRT2

				
				cbnz.To.SetTarget(mov)

				
				q = obj.Appendp(mov, c.newprog)
				q.As = AADD
				q.From.Type = obj.TYPE_CONST
				q.From.Offset = int64(c.autosize) + 8
				q.Reg = REGSP
				q.To.Type = obj.TYPE_REG
				q.To.Reg = REG_R20

				
				q = obj.Appendp(q, c.newprog)
				q.As = ACMP
				q.From.Type = obj.TYPE_REG
				q.From.Reg = REGRT2
				q.Reg = REG_R20

				
				q = obj.Appendp(q, c.newprog)
				q.As = ABNE
				q.To.Type = obj.TYPE_BRANCH
				q.To.SetTarget(end)

				
				q = obj.Appendp(q, c.newprog)
				q.As = AADD
				q.From.Type = obj.TYPE_CONST
				q.From.Offset = 8
				q.Reg = REGSP
				q.To.Type = obj.TYPE_REG
				q.To.Reg = REG_R20

				
				q = obj.Appendp(q, c.newprog)
				q.As = AMOVD
				q.From.Type = obj.TYPE_REG
				q.From.Reg = REG_R20
				q.To.Type = obj.TYPE_MEM
				q.To.Reg = REGRT1
				q.To.Offset = 0 

				
				q = obj.Appendp(q, c.newprog)
				q.As = AB
				q.To.Type = obj.TYPE_BRANCH
				q.To.SetTarget(end)
			}

		case obj.ARET:
			nocache(p)
			if p.From.Type == obj.TYPE_CONST {
				c.ctxt.Diag("using BECOME (%v) is not supported!", p)
				break
			}

			retjmp = p.To.Sym
			p.To = obj.Addr{}
			if c.cursym.Func().Text.Mark&LEAF != 0 {
				if c.autosize != 0 {
					p.As = AADD
					p.From.Type = obj.TYPE_CONST
					p.From.Offset = int64(c.autosize)
					p.To.Type = obj.TYPE_REG
					p.To.Reg = REGSP
					p.Spadj = -c.autosize

					
					p = obj.Appendp(p, c.newprog)
					p.As = ASUB
					p.From.Type = obj.TYPE_CONST
					p.From.Offset = 8
					p.Reg = REGSP
					p.To.Type = obj.TYPE_REG
					p.To.Reg = REGFP
				}
			} else {
				aoffset := c.autosize
				
				p.As = ALDP
				p.From.Type = obj.TYPE_MEM
				p.From.Offset = -8
				p.From.Reg = REGSP
				p.To.Type = obj.TYPE_REGREG
				p.To.Reg = REGFP
				p.To.Offset = REGLINK

				
				q = newprog()
				q.As = AADD
				q.From.Type = obj.TYPE_CONST
				q.From.Offset = int64(aoffset)
				q.To.Type = obj.TYPE_REG
				q.To.Reg = REGSP
				q.Spadj = -aoffset
				q.Pos = p.Pos
				q.Link = p.Link
				p.Link = q
				p = q
			}

			
			
			
			
			
			const debugRETZERO = false
			if debugRETZERO {
				if p.As != obj.ARET {
					q = newprog()
					q.Pos = p.Pos
					q.Link = p.Link
					p.Link = q
					p = q
				}
				p.As = AADR
				p.From.Type = obj.TYPE_BRANCH
				p.From.Offset = 0
				p.To.Type = obj.TYPE_REG
				p.To.Reg = REGTMP

			}

			if p.As != obj.ARET {
				q = newprog()
				q.Pos = p.Pos
				q.Link = p.Link
				p.Link = q
				p = q
			}

			if retjmp != nil { 
				p.As = AB
				p.To.Type = obj.TYPE_BRANCH
				p.To.Sym = retjmp
				p.Spadj = +c.autosize
				break
			}

			p.As = obj.ARET
			p.To.Type = obj.TYPE_MEM
			p.To.Offset = 0
			p.To.Reg = REGLINK
			p.Spadj = +c.autosize

		case AADD, ASUB:
			if p.To.Type == obj.TYPE_REG && p.To.Reg == REGSP && p.From.Type == obj.TYPE_CONST {
				if p.As == AADD {
					p.Spadj = int32(-p.From.Offset)
				} else {
					p.Spadj = int32(+p.From.Offset)
				}
			}

		case obj.AGETCALLERPC:
			if cursym.Leaf() {
				
				p.As = AMOVD
				p.From.Type = obj.TYPE_REG
				p.From.Reg = REGLINK
			} else {
				
				p.As = AMOVD
				p.From.Type = obj.TYPE_MEM
				p.From.Reg = REGSP
			}

		case obj.ADUFFCOPY:
			
			
			
			
			
			

			q1 := p
			
			q4 := obj.Appendp(p, c.newprog)
			q4.Pos = p.Pos
			q4.As = obj.ADUFFCOPY
			q4.To = p.To

			q1.As = AADR
			q1.From.Type = obj.TYPE_BRANCH
			q1.To.Type = obj.TYPE_REG
			q1.To.Reg = REG_R27

			q2 := obj.Appendp(q1, c.newprog)
			q2.Pos = p.Pos
			q2.As = ASTP
			q2.From.Type = obj.TYPE_REGREG
			q2.From.Reg = REGFP
			q2.From.Offset = int64(REG_R27)
			q2.To.Type = obj.TYPE_MEM
			q2.To.Reg = REGSP
			q2.To.Offset = -24

			
			q3 := obj.Appendp(q2, c.newprog)
			q3.Pos = p.Pos
			q3.As = ASUB
			q3.From.Type = obj.TYPE_CONST
			q3.From.Offset = 24
			q3.Reg = REGSP
			q3.To.Type = obj.TYPE_REG
			q3.To.Reg = REGFP

			q5 := obj.Appendp(q4, c.newprog)
			q5.Pos = p.Pos
			q5.As = ASUB
			q5.From.Type = obj.TYPE_CONST
			q5.From.Offset = 8
			q5.Reg = REGSP
			q5.To.Type = obj.TYPE_REG
			q5.To.Reg = REGFP
			q1.From.SetTarget(q5)
			p = q5

		case obj.ADUFFZERO:
			
			
			
			
			
			

			q1 := p
			
			q4 := obj.Appendp(p, c.newprog)
			q4.Pos = p.Pos
			q4.As = obj.ADUFFZERO
			q4.To = p.To

			q1.As = AADR
			q1.From.Type = obj.TYPE_BRANCH
			q1.To.Type = obj.TYPE_REG
			q1.To.Reg = REG_R27

			q2 := obj.Appendp(q1, c.newprog)
			q2.Pos = p.Pos
			q2.As = ASTP
			q2.From.Type = obj.TYPE_REGREG
			q2.From.Reg = REGFP
			q2.From.Offset = int64(REG_R27)
			q2.To.Type = obj.TYPE_MEM
			q2.To.Reg = REGSP
			q2.To.Offset = -24

			
			q3 := obj.Appendp(q2, c.newprog)
			q3.Pos = p.Pos
			q3.As = ASUB
			q3.From.Type = obj.TYPE_CONST
			q3.From.Offset = 24
			q3.Reg = REGSP
			q3.To.Type = obj.TYPE_REG
			q3.To.Reg = REGFP

			q5 := obj.Appendp(q4, c.newprog)
			q5.Pos = p.Pos
			q5.As = ASUB
			q5.From.Type = obj.TYPE_CONST
			q5.From.Offset = 8
			q5.Reg = REGSP
			q5.To.Type = obj.TYPE_REG
			q5.To.Reg = REGFP
			q1.From.SetTarget(q5)
			p = q5
		}

		if p.To.Type == obj.TYPE_REG && p.To.Reg == REGSP && p.Spadj == 0 {
			f := c.cursym.Func()
			if f.FuncFlag&abi.FuncFlagSPWrite == 0 {
				c.cursym.Func().FuncFlag |= abi.FuncFlagSPWrite
				if ctxt.Debugvlog || !ctxt.IsAsm {
					ctxt.Logf("auto-SPWRITE: %s %v\n", c.cursym.Name, p)
					if !ctxt.IsAsm {
						ctxt.Diag("invalid auto-SPWRITE in non-assembly")
						ctxt.DiagFlush()
						log.Fatalf("bad SPWRITE")
					}
				}
			}
		}
		if p.From.Type == obj.TYPE_SHIFT && (p.To.Reg == REG_RSP || p.Reg == REG_RSP) {
			offset := p.From.Offset
			op := offset & (3 << 22)
			if op != SHIFT_LL {
				ctxt.Diag("illegal combination: %v", p)
			}
			r := (offset >> 16) & 31
			shift := (offset >> 10) & 63
			if shift > 4 {
				
				
				
				shift = 7
			}
			p.From.Type = obj.TYPE_REG
			p.From.Reg = int16(REG_LSL + r + (shift&7)<<5)
			p.From.Offset = 0
		}
	}
}

func nocache(p *obj.Prog) {
	p.Optab = 0
	p.From.Class = 0
	p.To.Class = 0
}

var unaryDst = map[obj.As]bool{
	AWORD:  true,
	ADWORD: true,
	ABL:    true,
	AB:     true,
	ACLREX: true,
}

var Linkarm64 = obj.LinkArch{
	Arch:           sys.ArchARM64,
	Init:           buildop,
	Preprocess:     preprocess,
	Assemble:       span7,
	Progedit:       progedit,
	UnaryDst:       unaryDst,
	DWARFRegisters: ARM64DWARFRegisters,
}
