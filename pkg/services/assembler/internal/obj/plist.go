// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package obj

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/src"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/abi"
	"strings"
)

type Plist struct {
	Firstpc *Prog
	Curfn   interface{} 
}



type ProgAlloc func() *Prog

func Flushplist(ctxt *Link, plist *Plist, newprog ProgAlloc, myimportpath string) {
	
	var curtext *LSym
	var etext *Prog
	var text []*LSym

	var plink *Prog
	for p := plist.Firstpc; p != nil; p = plink {
		if ctxt.Debugasm > 0 && ctxt.Debugvlog {
			fmt.Printf("obj: %v\n", p)
		}
		plink = p.Link
		p.Link = nil

		switch p.As {
		case AEND:
			continue

		case ATEXT:
			s := p.From.Sym
			if s == nil {
				
				curtext = nil
				continue
			}
			text = append(text, s)
			etext = p
			curtext = s
			continue

		case AFUNCDATA:
			
			if curtext == nil { 
				continue
			}
			switch p.To.Sym.Name {
			case "go_args_stackmap":
				if p.From.Type != TYPE_CONST || p.From.Offset != abi.FUNCDATA_ArgsPointerMaps {
					ctxt.Diag("%s: FUNCDATA use of go_args_stackmap(SB) without FUNCDATA_ArgsPointerMaps", p.Pos)
				}
				p.To.Sym = ctxt.LookupDerived(curtext, curtext.Name+".args_stackmap")
			case "no_pointers_stackmap":
				if p.From.Type != TYPE_CONST || p.From.Offset != abi.FUNCDATA_LocalsPointerMaps {
					ctxt.Diag("%s: FUNCDATA use of no_pointers_stackmap(SB) without FUNCDATA_LocalsPointerMaps", p.Pos)
				}
				
				
				
				
				
				b := make([]byte, 8)
				ctxt.Arch.ByteOrder.PutUint32(b, 2)
				s := ctxt.GCLocalsSym(b)
				if !s.OnList() {
					ctxt.Globl(s, int64(len(s.P)), int(RODATA|DUPOK))
				}
				p.To.Sym = s
			}

		}

		if curtext == nil {
			etext = nil
			continue
		}
		etext.Link = p
		etext = p
	}

	if newprog == nil {
		newprog = ctxt.NewProg
	}

	
	if ctxt.IsAsm {
		for _, s := range text {
			if !strings.HasPrefix(s.Name, "\"\".") {
				continue
			}
			
			
			
			
			if s.ABI() != ABI0 {
				continue
			}
			foundArgMap, foundArgInfo := false, false
			for p := s.Func().Text; p != nil; p = p.Link {
				if p.As == AFUNCDATA && p.From.Type == TYPE_CONST {
					if p.From.Offset == abi.FUNCDATA_ArgsPointerMaps {
						foundArgMap = true
					}
					if p.From.Offset == abi.FUNCDATA_ArgInfo {
						foundArgInfo = true
					}
					if foundArgMap && foundArgInfo {
						break
					}
				}
			}
			if !foundArgMap {
				p := Appendp(s.Func().Text, newprog)
				p.As = AFUNCDATA
				p.From.Type = TYPE_CONST
				p.From.Offset = abi.FUNCDATA_ArgsPointerMaps
				p.To.Type = TYPE_MEM
				p.To.Name = NAME_EXTERN
				p.To.Sym = ctxt.LookupDerived(s, s.Name+".args_stackmap")
			}
			if !foundArgInfo {
				p := Appendp(s.Func().Text, newprog)
				p.As = AFUNCDATA
				p.From.Type = TYPE_CONST
				p.From.Offset = abi.FUNCDATA_ArgInfo
				p.To.Type = TYPE_MEM
				p.To.Name = NAME_EXTERN
				p.To.Sym = ctxt.LookupDerived(s, fmt.Sprintf("%s.arginfo%d", s.Name, s.ABI()))
			}
		}
	}

	
	for _, s := range text {
		mkfwd(s)
		if ctxt.Arch.ErrorCheck != nil {
			ctxt.Arch.ErrorCheck(ctxt, s)
		}
		linkpatch(ctxt, s, newprog)
		ctxt.Arch.Preprocess(ctxt, s, newprog)
		ctxt.Arch.Assemble(ctxt, s, newprog)
		if ctxt.Errors > 0 {
			continue
		}
		linkpcln(ctxt, s)
		if myimportpath != "" {
			ctxt.populateDWARF(plist.Curfn, s, myimportpath)
		}
		if ctxt.Headtype == objabi.Hwindows && ctxt.Arch.SEH != nil {
			s.Func().sehUnwindInfoSym = ctxt.Arch.SEH(ctxt, s)
		}
	}
}

func (ctxt *Link) InitTextSym(s *LSym, flag int, start src.XPos) {
	if s == nil {
		
		return
	}
	if s.Func() != nil {
		ctxt.Diag("%s: symbol %s redeclared\n\t%s: other declaration of symbol %s", ctxt.PosTable.Pos(start), s.Name, ctxt.PosTable.Pos(s.Func().Text.Pos), s.Name)
		return
	}
	s.NewFuncInfo()
	if s.OnList() {
		ctxt.Diag("%s: symbol %s redeclared", ctxt.PosTable.Pos(start), s.Name)
		return
	}

	
	
	
	_, startLine := ctxt.getFileSymbolAndLine(start)

	
	name := strings.Replace(s.Name, "\"\"", ctxt.Pkgpath, -1)
	s.Func().FuncID = objabi.GetFuncID(name, flag&WRAPPER != 0 || flag&ABIWRAPPER != 0)
	s.Func().FuncFlag = ctxt.toFuncFlag(flag)
	s.Func().StartLine = startLine
	s.Set(AttrOnList, true)
	s.Set(AttrDuplicateOK, flag&DUPOK != 0)
	s.Set(AttrNoSplit, flag&NOSPLIT != 0)
	s.Set(AttrReflectMethod, flag&REFLECTMETHOD != 0)
	s.Set(AttrWrapper, flag&WRAPPER != 0)
	s.Set(AttrABIWrapper, flag&ABIWRAPPER != 0)
	s.Set(AttrNeedCtxt, flag&NEEDCTXT != 0)
	s.Set(AttrNoFrame, flag&NOFRAME != 0)
	s.Set(AttrPkgInit, flag&PKGINIT != 0)
	s.Type = objabi.STEXT
	ctxt.Text = append(ctxt.Text, s)

	
	ctxt.dwarfSym(s)
}

func (ctxt *Link) toFuncFlag(flag int) abi.FuncFlag {
	var out abi.FuncFlag
	if flag&TOPFRAME != 0 {
		out |= abi.FuncFlagTopFrame
	}
	if ctxt.IsAsm {
		out |= abi.FuncFlagAsm
	}
	return out
}

func (ctxt *Link) Globl(s *LSym, size int64, flag int) {
	ctxt.GloblPos(s, size, flag, src.NoXPos)
}
func (ctxt *Link) GloblPos(s *LSym, size int64, flag int, pos src.XPos) {
	if s.OnList() {
		
		ctxt.Diag("%s: symbol %s redeclared", ctxt.PosTable.Pos(pos), s.Name)
	}
	s.Set(AttrOnList, true)
	ctxt.Data = append(ctxt.Data, s)
	s.Size = size
	if s.Type == 0 {
		s.Type = objabi.SBSS
	}
	if flag&DUPOK != 0 {
		s.Set(AttrDuplicateOK, true)
	}
	if flag&RODATA != 0 {
		s.Type = objabi.SRODATA
	} else if flag&NOPTR != 0 {
		if s.Type == objabi.SDATA {
			s.Type = objabi.SNOPTRDATA
		} else {
			s.Type = objabi.SNOPTRBSS
		}
	} else if flag&TLSBSS != 0 {
		s.Type = objabi.STLSBSS
	}
}




func (ctxt *Link) EmitEntryLiveness(s *LSym, p *Prog, newprog ProgAlloc) *Prog {
	pcdata := ctxt.EmitEntryStackMap(s, p, newprog)
	pcdata = ctxt.EmitEntryUnsafePoint(s, pcdata, newprog)
	return pcdata
}


func (ctxt *Link) EmitEntryStackMap(s *LSym, p *Prog, newprog ProgAlloc) *Prog {
	pcdata := Appendp(p, newprog)
	pcdata.Pos = s.Func().Text.Pos
	pcdata.As = APCDATA
	pcdata.From.Type = TYPE_CONST
	pcdata.From.Offset = abi.PCDATA_StackMapIndex
	pcdata.To.Type = TYPE_CONST
	pcdata.To.Offset = -1 

	return pcdata
}


func (ctxt *Link) EmitEntryUnsafePoint(s *LSym, p *Prog, newprog ProgAlloc) *Prog {
	pcdata := Appendp(p, newprog)
	pcdata.Pos = s.Func().Text.Pos
	pcdata.As = APCDATA
	pcdata.From.Type = TYPE_CONST
	pcdata.From.Offset = abi.PCDATA_UnsafePoint
	pcdata.To.Type = TYPE_CONST
	pcdata.To.Offset = -1

	return pcdata
}





func (ctxt *Link) StartUnsafePoint(p *Prog, newprog ProgAlloc) *Prog {
	pcdata := Appendp(p, newprog)
	pcdata.As = APCDATA
	pcdata.From.Type = TYPE_CONST
	pcdata.From.Offset = abi.PCDATA_UnsafePoint
	pcdata.To.Type = TYPE_CONST
	pcdata.To.Offset = abi.UnsafePointUnsafe

	return pcdata
}





func (ctxt *Link) EndUnsafePoint(p *Prog, newprog ProgAlloc, oldval int64) *Prog {
	pcdata := Appendp(p, newprog)
	pcdata.As = APCDATA
	pcdata.From.Type = TYPE_CONST
	pcdata.From.Offset = abi.PCDATA_UnsafePoint
	pcdata.To.Type = TYPE_CONST
	pcdata.To.Offset = oldval

	return pcdata
}











func MarkUnsafePoints(ctxt *Link, p0 *Prog, newprog ProgAlloc, isUnsafePoint, isRestartable func(*Prog) bool) {
	if isRestartable == nil {
		
		isRestartable = func(*Prog) bool { return false }
	}
	prev := p0
	prevPcdata := int64(-1) 
	prevRestart := int64(0)
	for p := prev.Link; p != nil; p, prev = p.Link, p {
		if p.As == APCDATA && p.From.Offset == abi.PCDATA_UnsafePoint {
			prevPcdata = p.To.Offset
			continue
		}
		if prevPcdata == abi.UnsafePointUnsafe {
			continue 
		}
		if isUnsafePoint(p) {
			q := ctxt.StartUnsafePoint(prev, newprog)
			q.Pc = p.Pc
			q.Link = p
			
			for p.Link != nil && isUnsafePoint(p.Link) {
				p = p.Link
			}
			if p.Link == nil {
				break 
			}
			p = ctxt.EndUnsafePoint(p, newprog, prevPcdata)
			p.Pc = p.Link.Pc
			continue
		}
		if isRestartable(p) {
			val := int64(abi.UnsafePointRestart1)
			if val == prevRestart {
				val = abi.UnsafePointRestart2
			}
			prevRestart = val
			q := Appendp(prev, newprog)
			q.As = APCDATA
			q.From.Type = TYPE_CONST
			q.From.Offset = abi.PCDATA_UnsafePoint
			q.To.Type = TYPE_CONST
			q.To.Offset = val
			q.Pc = p.Pc
			q.Link = p

			if p.Link == nil {
				break 
			}
			if isRestartable(p.Link) {
				
				
				continue
			}
			p = Appendp(p, newprog)
			p.As = APCDATA
			p.From.Type = TYPE_CONST
			p.From.Offset = abi.PCDATA_UnsafePoint
			p.To.Type = TYPE_CONST
			p.To.Offset = prevPcdata
			p.Pc = p.Link.Pc
		}
	}
}
