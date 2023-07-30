// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

// Writes dwarf information to object files.

package obj

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/dwarf"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/src"
	"fmt"
	"sort"
	"sync"
)



const (
	LINE_BASE   = -4
	LINE_RANGE  = 10
	PC_RANGE    = (255 - OPCODE_BASE) / LINE_RANGE
	OPCODE_BASE = 11
)









func (ctxt *Link) generateDebugLinesSymbol(s, lines *LSym) {
	dctxt := dwCtxt{ctxt}

	
	
	dctxt.AddUint8(lines, 0)
	dwarf.Uleb128put(dctxt, lines, 1+int64(ctxt.Arch.PtrSize))
	dctxt.AddUint8(lines, dwarf.DW_LNE_set_address)
	dctxt.AddAddress(lines, s, 0)

	
	
	stmt := true
	line := int64(1)
	pc := s.Func().Text.Pc
	var lastpc int64 
	name := ""
	prologue, wrotePrologue := false, false
	
	for p := s.Func().Text; p != nil; p = p.Link {
		prologue = prologue || (p.Pos.Xlogue() == src.PosPrologueEnd)
		
		if p.Pos.Line() == 0 || (p.Link != nil && p.Link.Pc == p.Pc) {
			continue
		}
		newStmt := p.Pos.IsStmt() != src.PosNotStmt
		newName, newLine := ctxt.getFileSymbolAndLine(p.Pos)

		
		wrote := false
		if name != newName {
			newFile := ctxt.PosTable.FileIndex(newName) + 1 
			dctxt.AddUint8(lines, dwarf.DW_LNS_set_file)
			dwarf.Uleb128put(dctxt, lines, int64(newFile))
			name = newName
			wrote = true
		}
		if prologue && !wrotePrologue {
			dctxt.AddUint8(lines, uint8(dwarf.DW_LNS_set_prologue_end))
			wrotePrologue = true
			wrote = true
		}
		if stmt != newStmt {
			dctxt.AddUint8(lines, uint8(dwarf.DW_LNS_negate_stmt))
			stmt = newStmt
			wrote = true
		}

		if line != int64(newLine) || wrote {
			pcdelta := p.Pc - pc
			lastpc = p.Pc
			putpclcdelta(ctxt, dctxt, lines, uint64(pcdelta), int64(newLine)-line)
			line, pc = int64(newLine), p.Pc
		}
	}

	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	lastlen := uint64(s.Size - (lastpc - s.Func().Text.Pc))
	dctxt.AddUint8(lines, dwarf.DW_LNS_advance_pc)
	dwarf.Uleb128put(dctxt, lines, int64(lastlen))
	dctxt.AddUint8(lines, 0) 
	dwarf.Uleb128put(dctxt, lines, 1)
	dctxt.AddUint8(lines, dwarf.DW_LNE_end_sequence)
}

func putpclcdelta(linkctxt *Link, dctxt dwCtxt, s *LSym, deltaPC uint64, deltaLC int64) {
	
	
	var opcode int64
	if deltaLC < LINE_BASE {
		if deltaPC >= PC_RANGE {
			opcode = OPCODE_BASE + (LINE_RANGE * PC_RANGE)
		} else {
			opcode = OPCODE_BASE + (LINE_RANGE * int64(deltaPC))
		}
	} else if deltaLC < LINE_BASE+LINE_RANGE {
		if deltaPC >= PC_RANGE {
			opcode = OPCODE_BASE + (deltaLC - LINE_BASE) + (LINE_RANGE * PC_RANGE)
			if opcode > 255 {
				opcode -= LINE_RANGE
			}
		} else {
			opcode = OPCODE_BASE + (deltaLC - LINE_BASE) + (LINE_RANGE * int64(deltaPC))
		}
	} else {
		if deltaPC <= PC_RANGE {
			opcode = OPCODE_BASE + (LINE_RANGE - 1) + (LINE_RANGE * int64(deltaPC))
			if opcode > 255 {
				opcode = 255
			}
		} else {
			
			
			
			
			
			
			
			
			
			
			switch deltaPC - PC_RANGE {
			
			
			
			
			
			
			
			
			case PC_RANGE, (1 << 7) - 1, (1 << 16) - 1, (1 << 21) - 1, (1 << 28) - 1,
				(1 << 35) - 1, (1 << 42) - 1, (1 << 49) - 1, (1 << 56) - 1, (1 << 63) - 1:
				opcode = 255
			default:
				opcode = OPCODE_BASE + LINE_RANGE*PC_RANGE - 1 
			}
		}
	}
	if opcode < OPCODE_BASE || opcode > 255 {
		panic(fmt.Sprintf("produced invalid special opcode %d", opcode))
	}

	
	deltaPC -= uint64((opcode - OPCODE_BASE) / LINE_RANGE)
	deltaLC -= (opcode-OPCODE_BASE)%LINE_RANGE + LINE_BASE

	
	if deltaPC != 0 {
		if deltaPC <= PC_RANGE {
			
			
			opcode -= LINE_RANGE * int64(PC_RANGE-deltaPC)
			if opcode < OPCODE_BASE {
				panic(fmt.Sprintf("produced invalid special opcode %d", opcode))
			}
			dctxt.AddUint8(s, dwarf.DW_LNS_const_add_pc)
		} else if (1<<14) <= deltaPC && deltaPC < (1<<16) {
			dctxt.AddUint8(s, dwarf.DW_LNS_fixed_advance_pc)
			dctxt.AddUint16(s, uint16(deltaPC))
		} else {
			dctxt.AddUint8(s, dwarf.DW_LNS_advance_pc)
			dwarf.Uleb128put(dctxt, s, int64(deltaPC))
		}
	}

	
	if deltaLC != 0 {
		dctxt.AddUint8(s, dwarf.DW_LNS_advance_line)
		dwarf.Sleb128put(dctxt, s, deltaLC)
	}

	
	dctxt.AddUint8(s, uint8(opcode))
}


type dwCtxt struct{ *Link }

func (c dwCtxt) PtrSize() int {
	return c.Arch.PtrSize
}
func (c dwCtxt) AddInt(s dwarf.Sym, size int, i int64) {
	ls := s.(*LSym)
	ls.WriteInt(c.Link, ls.Size, size, i)
}
func (c dwCtxt) AddUint16(s dwarf.Sym, i uint16) {
	c.AddInt(s, 2, int64(i))
}
func (c dwCtxt) AddUint8(s dwarf.Sym, i uint8) {
	b := []byte{byte(i)}
	c.AddBytes(s, b)
}
func (c dwCtxt) AddBytes(s dwarf.Sym, b []byte) {
	ls := s.(*LSym)
	ls.WriteBytes(c.Link, ls.Size, b)
}
func (c dwCtxt) AddString(s dwarf.Sym, v string) {
	ls := s.(*LSym)
	ls.WriteString(c.Link, ls.Size, len(v), v)
	ls.WriteInt(c.Link, ls.Size, 1, 0)
}
func (c dwCtxt) AddAddress(s dwarf.Sym, data interface{}, value int64) {
	ls := s.(*LSym)
	size := c.PtrSize()
	if data != nil {
		rsym := data.(*LSym)
		ls.WriteAddr(c.Link, ls.Size, size, rsym, value)
	} else {
		ls.WriteInt(c.Link, ls.Size, size, value)
	}
}
func (c dwCtxt) AddCURelativeAddress(s dwarf.Sym, data interface{}, value int64) {
	ls := s.(*LSym)
	rsym := data.(*LSym)
	ls.WriteCURelativeAddr(c.Link, ls.Size, rsym, value)
}
func (c dwCtxt) AddSectionOffset(s dwarf.Sym, size int, t interface{}, ofs int64) {
	panic("should be used only in the linker")
}
func (c dwCtxt) AddDWARFAddrSectionOffset(s dwarf.Sym, t interface{}, ofs int64) {
	size := 4
	if isDwarf64(c.Link) {
		size = 8
	}

	ls := s.(*LSym)
	rsym := t.(*LSym)
	ls.WriteAddr(c.Link, ls.Size, size, rsym, ofs)
	r := &ls.R[len(ls.R)-1]
	r.Type = objabi.R_DWARFSECREF
}

func (c dwCtxt) AddFileRef(s dwarf.Sym, f interface{}) {
	ls := s.(*LSym)
	rsym := f.(*LSym)
	fidx := c.Link.PosTable.FileIndex(rsym.Name)
	
	
	
	ls.WriteInt(c.Link, ls.Size, 4, int64(fidx+1))
}

func (c dwCtxt) CurrentOffset(s dwarf.Sym) int64 {
	ls := s.(*LSym)
	return ls.Size
}






func (c dwCtxt) RecordDclReference(from dwarf.Sym, to dwarf.Sym, dclIdx int, inlIndex int) {
	ls := from.(*LSym)
	tls := to.(*LSym)
	ridx := len(ls.R) - 1
	c.Link.DwFixups.ReferenceChildDIE(ls, ridx, tls, dclIdx, inlIndex)
}

func (c dwCtxt) RecordChildDieOffsets(s dwarf.Sym, vars []*dwarf.Var, offsets []int32) {
	ls := s.(*LSym)
	c.Link.DwFixups.RegisterChildDIEOffsets(ls, vars, offsets)
}

func (c dwCtxt) Logf(format string, args ...interface{}) {
	c.Link.Logf(format, args...)
}

func isDwarf64(ctxt *Link) bool {
	return ctxt.Headtype == objabi.Haix
}

func (ctxt *Link) dwarfSym(s *LSym) (dwarfInfoSym, dwarfLocSym, dwarfRangesSym, dwarfAbsFnSym, dwarfDebugLines *LSym) {
	if s.Type != objabi.STEXT {
		ctxt.Diag("dwarfSym of non-TEXT %v", s)
	}
	fn := s.Func()
	if fn.dwarfInfoSym == nil {
		fn.dwarfInfoSym = &LSym{
			Type: objabi.SDWARFFCN,
		}
		if ctxt.Flag_locationlists {
			fn.dwarfLocSym = &LSym{
				Type: objabi.SDWARFLOC,
			}
		}
		fn.dwarfRangesSym = &LSym{
			Type: objabi.SDWARFRANGE,
		}
		fn.dwarfDebugLinesSym = &LSym{
			Type: objabi.SDWARFLINES,
		}
		if s.WasInlined() {
			fn.dwarfAbsFnSym = ctxt.DwFixups.AbsFuncDwarfSym(s)
		}
	}
	return fn.dwarfInfoSym, fn.dwarfLocSym, fn.dwarfRangesSym, fn.dwarfAbsFnSym, fn.dwarfDebugLinesSym
}

func (s *LSym) Length(dwarfContext interface{}) int64 {
	return s.Size
}




func (ctxt *Link) fileSymbol(fn *LSym) *LSym {
	p := fn.Func().Text
	if p != nil {
		f, _ := ctxt.getFileSymbolAndLine(p.Pos)
		fsym := ctxt.Lookup(f)
		return fsym
	}
	return nil
}




func (ctxt *Link) populateDWARF(curfn interface{}, s *LSym, myimportpath string) {
	info, loc, ranges, absfunc, lines := ctxt.dwarfSym(s)
	if info.Size != 0 {
		ctxt.Diag("makeFuncDebugEntry double process %v", s)
	}
	var scopes []dwarf.Scope
	var inlcalls dwarf.InlCalls
	if ctxt.DebugInfo != nil {
		
		
		scopes, inlcalls, _ = ctxt.DebugInfo(s, info, curfn)
	}
	var err error
	dwctxt := dwCtxt{ctxt}
	filesym := ctxt.fileSymbol(s)
	fnstate := &dwarf.FnState{
		Name:          s.Name,
		Importpath:    myimportpath,
		Info:          info,
		Filesym:       filesym,
		Loc:           loc,
		Ranges:        ranges,
		Absfn:         absfunc,
		StartPC:       s,
		Size:          s.Size,
		StartLine:     s.Func().StartLine,
		External:      !s.Static(),
		Scopes:        scopes,
		InlCalls:      inlcalls,
		UseBASEntries: ctxt.UseBASEntries,
	}
	if absfunc != nil {
		err = dwarf.PutAbstractFunc(dwctxt, fnstate)
		if err != nil {
			ctxt.Diag("emitting DWARF for %s failed: %v", s.Name, err)
		}
		err = dwarf.PutConcreteFunc(dwctxt, fnstate, s.Wrapper())
	} else {
		err = dwarf.PutDefaultFunc(dwctxt, fnstate, s.Wrapper())
	}
	if err != nil {
		ctxt.Diag("emitting DWARF for %s failed: %v", s.Name, err)
	}
	
	ctxt.generateDebugLinesSymbol(s, lines)
}



func (ctxt *Link) DwarfIntConst(myimportpath, name, typename string, val int64) {
	if myimportpath == "" {
		return
	}
	s := ctxt.LookupInit(dwarf.ConstInfoPrefix+myimportpath, func(s *LSym) {
		s.Type = objabi.SDWARFCONST
		ctxt.Data = append(ctxt.Data, s)
	})
	dwarf.PutIntConst(dwCtxt{ctxt}, s, ctxt.Lookup(dwarf.InfoPrefix+typename), myimportpath+"."+name, val)
}



func (ctxt *Link) DwarfGlobal(myimportpath, typename string, varSym *LSym) {
	if myimportpath == "" || varSym.Local() {
		return
	}
	varname := varSym.Name
	dieSym := &LSym{
		Type: objabi.SDWARFVAR,
	}
	varSym.NewVarInfo().dwarfInfoSym = dieSym
	ctxt.Data = append(ctxt.Data, dieSym)
	typeSym := ctxt.Lookup(dwarf.InfoPrefix + typename)
	dwarf.PutGlobal(dwCtxt{ctxt}, dieSym, typeSym, varSym, varname)
}

func (ctxt *Link) DwarfAbstractFunc(curfn interface{}, s *LSym, myimportpath string) {
	absfn := ctxt.DwFixups.AbsFuncDwarfSym(s)
	if absfn.Size != 0 {
		ctxt.Diag("internal error: DwarfAbstractFunc double process %v", s)
	}
	if s.Func() == nil {
		s.NewFuncInfo()
	}
	scopes, _, startPos := ctxt.DebugInfo(s, absfn, curfn)
	_, startLine := ctxt.getFileSymbolAndLine(startPos)
	dwctxt := dwCtxt{ctxt}
	fnstate := dwarf.FnState{
		Name:          s.Name,
		Importpath:    myimportpath,
		Info:          absfn,
		Absfn:         absfn,
		StartLine:     startLine,
		External:      !s.Static(),
		Scopes:        scopes,
		UseBASEntries: ctxt.UseBASEntries,
	}
	if err := dwarf.PutAbstractFunc(dwctxt, &fnstate); err != nil {
		ctxt.Diag("emitting DWARF for %s failed: %v", s.Name, err)
	}
}



































type DwarfFixupTable struct {
	ctxt      *Link
	mu        sync.Mutex
	symtab    map[*LSym]int 
	svec      []symFixups
	precursor map[*LSym]fnState 
}

type symFixups struct {
	fixups   []relFixup
	doffsets []declOffset
	inlIndex int32
	defseen  bool
}

type declOffset struct {
	
	dclIdx int32
	
	offset int32
}

type relFixup struct {
	refsym *LSym
	relidx int32
	dclidx int32
}

type fnState struct {
	
	precursor interface{}
	
	absfn *LSym
}

func NewDwarfFixupTable(ctxt *Link) *DwarfFixupTable {
	return &DwarfFixupTable{
		ctxt:      ctxt,
		symtab:    make(map[*LSym]int),
		precursor: make(map[*LSym]fnState),
	}
}

func (ft *DwarfFixupTable) GetPrecursorFunc(s *LSym) interface{} {
	if fnstate, found := ft.precursor[s]; found {
		return fnstate.precursor
	}
	return nil
}

func (ft *DwarfFixupTable) SetPrecursorFunc(s *LSym, fn interface{}) {
	if _, found := ft.precursor[s]; found {
		ft.ctxt.Diag("internal error: DwarfFixupTable.SetPrecursorFunc double call on %v", s)
	}

	
	
	
	absfn := ft.ctxt.LookupDerived(s, dwarf.InfoPrefix+s.Name+dwarf.AbstractFuncSuffix)
	absfn.Set(AttrDuplicateOK, true)
	absfn.Type = objabi.SDWARFABSFCN
	ft.ctxt.Data = append(ft.ctxt.Data, absfn)

	
	
	
	
	if fn := s.Func(); fn != nil && fn.dwarfAbsFnSym == nil {
		fn.dwarfAbsFnSym = absfn
	}

	ft.precursor[s] = fnState{precursor: fn, absfn: absfn}
}



func (ft *DwarfFixupTable) ReferenceChildDIE(s *LSym, ridx int, tgt *LSym, dclidx int, inlIndex int) {
	
	ft.mu.Lock()
	defer ft.mu.Unlock()

	
	idx, found := ft.symtab[tgt]
	if !found {
		ft.svec = append(ft.svec, symFixups{inlIndex: int32(inlIndex)})
		idx = len(ft.svec) - 1
		ft.symtab[tgt] = idx
	}

	
	
	sf := &ft.svec[idx]
	if len(sf.doffsets) > 0 {
		found := false
		for _, do := range sf.doffsets {
			if do.dclIdx == int32(dclidx) {
				off := do.offset
				s.R[ridx].Add += int64(off)
				found = true
				break
			}
		}
		if !found {
			ft.ctxt.Diag("internal error: DwarfFixupTable.ReferenceChildDIE unable to locate child DIE offset for dclIdx=%d src=%v tgt=%v", dclidx, s, tgt)
		}
	} else {
		sf.fixups = append(sf.fixups, relFixup{s, int32(ridx), int32(dclidx)})
	}
}






func (ft *DwarfFixupTable) RegisterChildDIEOffsets(s *LSym, vars []*dwarf.Var, coffsets []int32) {
	
	if len(vars) != len(coffsets) {
		ft.ctxt.Diag("internal error: RegisterChildDIEOffsets vars/offsets length mismatch")
		return
	}

	
	doffsets := make([]declOffset, len(coffsets))
	for i := range coffsets {
		doffsets[i].dclIdx = vars[i].ChildIndex
		doffsets[i].offset = coffsets[i]
	}

	ft.mu.Lock()
	defer ft.mu.Unlock()

	
	idx, found := ft.symtab[s]
	if !found {
		sf := symFixups{inlIndex: -1, defseen: true, doffsets: doffsets}
		ft.svec = append(ft.svec, sf)
		ft.symtab[s] = len(ft.svec) - 1
	} else {
		sf := &ft.svec[idx]
		sf.doffsets = doffsets
		sf.defseen = true
	}
}

func (ft *DwarfFixupTable) processFixups(slot int, s *LSym) {
	sf := &ft.svec[slot]
	for _, f := range sf.fixups {
		dfound := false
		for _, doffset := range sf.doffsets {
			if doffset.dclIdx == f.dclidx {
				f.refsym.R[f.relidx].Add += int64(doffset.offset)
				dfound = true
				break
			}
		}
		if !dfound {
			ft.ctxt.Diag("internal error: DwarfFixupTable has orphaned fixup on %v targeting %v relidx=%d dclidx=%d", f.refsym, s, f.relidx, f.dclidx)
		}
	}
}



func (ft *DwarfFixupTable) AbsFuncDwarfSym(fnsym *LSym) *LSym {
	
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if fnstate, found := ft.precursor[fnsym]; found {
		return fnstate.absfn
	}
	ft.ctxt.Diag("internal error: AbsFuncDwarfSym requested for %v, not seen during inlining", fnsym)
	return nil
}







func (ft *DwarfFixupTable) Finalize(myimportpath string, trace bool) {
	if trace {
		ft.ctxt.Logf("DwarfFixupTable.Finalize invoked for %s\n", myimportpath)
	}

	
	
	fns := make([]*LSym, len(ft.precursor))
	idx := 0
	for fn := range ft.precursor {
		fns[idx] = fn
		idx++
	}
	sort.Sort(BySymName(fns))

	
	if ft.ctxt.InParallel {
		ft.ctxt.Diag("internal error: DwarfFixupTable.Finalize call during parallel backend")
	}

	
	for _, s := range fns {
		absfn := ft.AbsFuncDwarfSym(s)
		slot, found := ft.symtab[absfn]
		if !found || !ft.svec[slot].defseen {
			ft.ctxt.GenAbstractFunc(s)
		}
	}

	
	for _, s := range fns {
		absfn := ft.AbsFuncDwarfSym(s)
		slot, found := ft.symtab[absfn]
		if !found {
			ft.ctxt.Diag("internal error: DwarfFixupTable.Finalize orphan abstract function for %v", s)
		} else {
			ft.processFixups(slot, s)
		}
	}
}

type BySymName []*LSym

func (s BySymName) Len() int           { return len(s) }
func (s BySymName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s BySymName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
