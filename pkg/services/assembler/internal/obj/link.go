// Derived from Inferno utils/6l/l.h and related files.
// https://bitbucket.org/inferno-os/inferno-os/src/master/utils/6l/l.h
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
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

package obj

import (
	"bufio"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/dwarf"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/goobj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/src"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/sys"
	"encoding/binary"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/abi"
	"sync"
	"sync/atomic"
)





















































































































































type Addr struct {
	Reg    int16
	Index  int16
	Scale  int16 
	Type   AddrType
	Name   AddrName
	Class  int8
	Offset int64
	Sym    *LSym

	
	
	
	
	
	Val interface{}
}

type AddrName int8

const (
	NAME_NONE AddrName = iota
	NAME_EXTERN
	NAME_STATIC
	NAME_AUTO
	NAME_PARAM
	
	
	NAME_GOTREF
	
	NAME_TOCREF
)

//go:generate stringer -type AddrType

type AddrType uint8

const (
	TYPE_NONE AddrType = iota
	TYPE_BRANCH
	TYPE_TEXTSIZE
	TYPE_MEM
	TYPE_CONST
	TYPE_FCONST
	TYPE_SCONST
	TYPE_REG
	TYPE_ADDR
	TYPE_SHIFT
	TYPE_REGREG
	TYPE_REGREG2
	TYPE_INDIR
	TYPE_REGLIST
	TYPE_SPECIAL
)

func (a *Addr) Target() *Prog {
	if a.Type == TYPE_BRANCH && a.Val != nil {
		return a.Val.(*Prog)
	}
	return nil
}
func (a *Addr) SetTarget(t *Prog) {
	if a.Type != TYPE_BRANCH {
		panic("setting branch target when type is not TYPE_BRANCH")
	}
	a.Val = t
}

func (a *Addr) SetConst(v int64) {
	a.Sym = nil
	a.Type = TYPE_CONST
	a.Offset = v
}
































type Prog struct {
	Ctxt     *Link     
	Link     *Prog     
	From     Addr      
	RestArgs []AddrPos 
	To       Addr      
	Pool     *Prog     
	Forwd    *Prog     
	Rel      *Prog     
	Pc       int64     
	Pos      src.XPos  
	Spadj    int32     
	As       As        
	Reg      int16     
	RegTo2   int16     
	Mark     uint16    
	Optab    uint16    
	Scond    uint8     
	Back     uint8     
	Ft       uint8     
	Tt       uint8     
	Isize    uint8     
}


type AddrPos struct {
	Addr
	Pos OperandPos
}

type OperandPos int8

const (
	Source OperandPos = iota
	Destination
)





func (p *Prog) From3Type() AddrType {
	if p.RestArgs == nil {
		return TYPE_NONE
	}
	return p.RestArgs[0].Type
}










func (p *Prog) GetFrom3() *Addr {
	if p.RestArgs == nil {
		return nil
	}
	return &p.RestArgs[0].Addr
}





func (p *Prog) SetFrom3(a Addr) {
	p.RestArgs = []AddrPos{{a, Source}}
}




func (p *Prog) SetFrom3Reg(reg int16) {
	p.SetFrom3(Addr{Type: TYPE_REG, Reg: reg})
}




func (p *Prog) SetFrom3Const(off int64) {
	p.SetFrom3(Addr{Type: TYPE_CONST, Offset: off})
}



func (p *Prog) SetTo2(a Addr) {
	p.RestArgs = []AddrPos{{a, Destination}}
}


func (p *Prog) GetTo2() *Addr {
	if p.RestArgs == nil {
		return nil
	}
	return &p.RestArgs[0].Addr
}


func (p *Prog) SetRestArgs(args []Addr) {
	for i := range args {
		p.RestArgs = append(p.RestArgs, AddrPos{args[i], Source})
	}
}






type As int16


const (
	AXXX As = iota
	ACALL
	ADUFFCOPY
	ADUFFZERO
	AEND
	AFUNCDATA
	AJMP
	ANOP
	APCALIGN
	APCDATA
	ARET
	AGETCALLERPC
	ATEXT
	AUNDEF
	A_ARCHSPECIFIC
)








const (
	ABase386 = (1 + iota) << 11
	ABaseARM
	ABaseAMD64
	ABasePPC64
	ABaseARM64
	ABaseMIPS
	ABaseLoong64
	ABaseRISCV
	ABaseS390X
	ABaseWasm

	AllowedOpCodes = 1 << 11            
	AMask          = AllowedOpCodes - 1 
)



type LSym struct {
	Name string
	Type objabi.SymKind
	Attribute

	Size   int64
	Gotype *LSym
	P      []byte
	R      []Reloc

	Extra *interface{} 

	Pkg    string
	PkgIdx int32
	SymIdx int32
}


type FuncInfo struct {
	Args      int32
	Locals    int32
	Align     int32
	FuncID    abi.FuncID
	FuncFlag  abi.FuncFlag
	StartLine int32
	Text      *Prog
	Autot     map[*LSym]struct{}
	Pcln      Pcln
	InlMarks  []InlMark
	spills    []RegSpill

	dwarfInfoSym       *LSym
	dwarfLocSym        *LSym
	dwarfRangesSym     *LSym
	dwarfAbsFnSym      *LSym
	dwarfDebugLinesSym *LSym

	GCArgs             *LSym
	GCLocals           *LSym
	StackObjects       *LSym
	OpenCodedDeferInfo *LSym
	ArgInfo            *LSym 
	ArgLiveInfo        *LSym 
	WrapInfo           *LSym 
	JumpTables         []JumpTable

	FuncInfoSym   *LSym
	WasmImportSym *LSym
	WasmImport    *WasmImport

	sehUnwindInfoSym *LSym
}





type JumpTable struct {
	Sym     *LSym
	Targets []*Prog
}


func (s *LSym) NewFuncInfo() *FuncInfo {
	if s.Extra != nil {
		panic(fmt.Sprintf("invalid use of LSym - NewFuncInfo with Extra of type %T", *s.Extra))
	}
	f := new(FuncInfo)
	s.Extra = new(interface{})
	*s.Extra = f
	return f
}


func (s *LSym) Func() *FuncInfo {
	if s.Extra == nil {
		return nil
	}
	f, _ := (*s.Extra).(*FuncInfo)
	return f
}

type VarInfo struct {
	dwarfInfoSym *LSym
}


func (s *LSym) NewVarInfo() *VarInfo {
	if s.Extra != nil {
		panic(fmt.Sprintf("invalid use of LSym - NewVarInfo with Extra of type %T", *s.Extra))
	}
	f := new(VarInfo)
	s.Extra = new(interface{})
	*s.Extra = f
	return f
}


func (s *LSym) VarInfo() *VarInfo {
	if s.Extra == nil {
		return nil
	}
	f, _ := (*s.Extra).(*VarInfo)
	return f
}



type FileInfo struct {
	Name string 
	Size int64  
}


func (s *LSym) NewFileInfo() *FileInfo {
	if s.Extra != nil {
		panic(fmt.Sprintf("invalid use of LSym - NewFileInfo with Extra of type %T", *s.Extra))
	}
	f := new(FileInfo)
	s.Extra = new(interface{})
	*s.Extra = f
	return f
}


func (s *LSym) File() *FileInfo {
	if s.Extra == nil {
		return nil
	}
	f, _ := (*s.Extra).(*FileInfo)
	return f
}



type TypeInfo struct {
	Type interface{} 
}

func (s *LSym) NewTypeInfo() *TypeInfo {
	if s.Extra != nil {
		panic(fmt.Sprintf("invalid use of LSym - NewTypeInfo with Extra of type %T", *s.Extra))
	}
	t := new(TypeInfo)
	s.Extra = new(interface{})
	*s.Extra = t
	return t
}




type WasmImport struct {
	
	
	Module string
	
	
	Name string
	
	Params []WasmField
	
	Results []WasmField
}

func (wi *WasmImport) CreateSym(ctxt *Link) *LSym {
	var sym LSym

	var b [8]byte
	writeByte := func(x byte) {
		sym.WriteBytes(ctxt, sym.Size, []byte{x})
	}
	writeUint32 := func(x uint32) {
		binary.LittleEndian.PutUint32(b[:], x)
		sym.WriteBytes(ctxt, sym.Size, b[:4])
	}
	writeInt64 := func(x int64) {
		binary.LittleEndian.PutUint64(b[:], uint64(x))
		sym.WriteBytes(ctxt, sym.Size, b[:])
	}
	writeString := func(s string) {
		writeUint32(uint32(len(s)))
		sym.WriteString(ctxt, sym.Size, len(s), s)
	}
	writeString(wi.Module)
	writeString(wi.Name)
	writeUint32(uint32(len(wi.Params)))
	for _, f := range wi.Params {
		writeByte(byte(f.Type))
		writeInt64(f.Offset)
	}
	writeUint32(uint32(len(wi.Results)))
	for _, f := range wi.Results {
		writeByte(byte(f.Type))
		writeInt64(f.Offset)
	}

	return &sym
}

type WasmField struct {
	Type WasmFieldType
	
	
	
	Offset int64
}

type WasmFieldType byte

const (
	WasmI32 WasmFieldType = iota
	WasmI64
	WasmF32
	WasmF64
	WasmPtr
)

type InlMark struct {
	
	
	
	
	p  *Prog
	id int32
}





func (fi *FuncInfo) AddInlMark(p *Prog, id int32) {
	fi.InlMarks = append(fi.InlMarks, InlMark{p: p, id: id})
}


func (fi *FuncInfo) AddSpill(s RegSpill) {
	fi.spills = append(fi.spills, s)
}



func (fi *FuncInfo) RecordAutoType(gotype *LSym) {
	if fi.Autot == nil {
		fi.Autot = make(map[*LSym]struct{})
	}
	fi.Autot[gotype] = struct{}{}
}

//go:generate stringer -type ABI


type ABI uint8

const (
	
	
	
	
	
	ABI0 ABI = iota

	
	
	
	
	ABIInternal

	ABICount
)




func ParseABI(abistr string) (ABI, bool) {
	switch abistr {
	default:
		return ABI0, false
	case "ABI0":
		return ABI0, true
	case "ABIInternal":
		return ABIInternal, true
	}
}


type ABISet uint8

const (
	
	
	ABISetCallable ABISet = (1 << ABI0) | (1 << ABIInternal)
)


var _ ABISet = 1 << (ABICount - 1)

func ABISetOf(abi ABI) ABISet {
	return 1 << abi
}

func (a *ABISet) Set(abi ABI, value bool) {
	if value {
		*a |= 1 << abi
	} else {
		*a &^= 1 << abi
	}
}

func (a *ABISet) Get(abi ABI) bool {
	return (*a>>abi)&1 != 0
}

func (a ABISet) String() string {
	s := "{"
	for i := ABI(0); a != 0; i++ {
		if a&(1<<i) != 0 {
			if s != "{" {
				s += ","
			}
			s += i.String()
			a &^= 1 << i
		}
	}
	return s + "}"
}


type Attribute uint32

const (
	AttrDuplicateOK Attribute = 1 << iota
	AttrCFunc
	AttrNoSplit
	AttrLeaf
	AttrWrapper
	AttrNeedCtxt
	AttrNoFrame
	AttrOnList
	AttrStatic

	
	AttrMakeTypelink

	
	
	
	
	
	
	AttrReflectMethod

	
	
	
	
	
	
	AttrLocal

	
	
	AttrWasInlined

	
	
	AttrIndexed

	
	
	
	
	AttrUsedInIface

	
	AttrContentAddressable

	
	
	AttrABIWrapper

	
	AttrPcdata

	
	AttrPkgInit

	
	
	
	
	
	attrABIBase
)

func (a *Attribute) load() Attribute { return Attribute(atomic.LoadUint32((*uint32)(a))) }

func (a *Attribute) DuplicateOK() bool        { return a.load()&AttrDuplicateOK != 0 }
func (a *Attribute) MakeTypelink() bool       { return a.load()&AttrMakeTypelink != 0 }
func (a *Attribute) CFunc() bool              { return a.load()&AttrCFunc != 0 }
func (a *Attribute) NoSplit() bool            { return a.load()&AttrNoSplit != 0 }
func (a *Attribute) Leaf() bool               { return a.load()&AttrLeaf != 0 }
func (a *Attribute) OnList() bool             { return a.load()&AttrOnList != 0 }
func (a *Attribute) ReflectMethod() bool      { return a.load()&AttrReflectMethod != 0 }
func (a *Attribute) Local() bool              { return a.load()&AttrLocal != 0 }
func (a *Attribute) Wrapper() bool            { return a.load()&AttrWrapper != 0 }
func (a *Attribute) NeedCtxt() bool           { return a.load()&AttrNeedCtxt != 0 }
func (a *Attribute) NoFrame() bool            { return a.load()&AttrNoFrame != 0 }
func (a *Attribute) Static() bool             { return a.load()&AttrStatic != 0 }
func (a *Attribute) WasInlined() bool         { return a.load()&AttrWasInlined != 0 }
func (a *Attribute) Indexed() bool            { return a.load()&AttrIndexed != 0 }
func (a *Attribute) UsedInIface() bool        { return a.load()&AttrUsedInIface != 0 }
func (a *Attribute) ContentAddressable() bool { return a.load()&AttrContentAddressable != 0 }
func (a *Attribute) ABIWrapper() bool         { return a.load()&AttrABIWrapper != 0 }
func (a *Attribute) IsPcdata() bool           { return a.load()&AttrPcdata != 0 }
func (a *Attribute) IsPkgInit() bool          { return a.load()&AttrPkgInit != 0 }

func (a *Attribute) Set(flag Attribute, value bool) {
	for {
		v0 := a.load()
		v := v0
		if value {
			v |= flag
		} else {
			v &^= flag
		}
		if atomic.CompareAndSwapUint32((*uint32)(a), uint32(v0), uint32(v)) {
			break
		}
	}
}

func (a *Attribute) ABI() ABI { return ABI(a.load() / attrABIBase) }
func (a *Attribute) SetABI(abi ABI) {
	const mask = 1 
	for {
		v0 := a.load()
		v := (v0 &^ (mask * attrABIBase)) | Attribute(abi)*attrABIBase
		if atomic.CompareAndSwapUint32((*uint32)(a), uint32(v0), uint32(v)) {
			break
		}
	}
}

var textAttrStrings = [...]struct {
	bit Attribute
	s   string
}{
	{bit: AttrDuplicateOK, s: "DUPOK"},
	{bit: AttrMakeTypelink, s: ""},
	{bit: AttrCFunc, s: "CFUNC"},
	{bit: AttrNoSplit, s: "NOSPLIT"},
	{bit: AttrLeaf, s: "LEAF"},
	{bit: AttrOnList, s: ""},
	{bit: AttrReflectMethod, s: "REFLECTMETHOD"},
	{bit: AttrLocal, s: "LOCAL"},
	{bit: AttrWrapper, s: "WRAPPER"},
	{bit: AttrNeedCtxt, s: "NEEDCTXT"},
	{bit: AttrNoFrame, s: "NOFRAME"},
	{bit: AttrStatic, s: "STATIC"},
	{bit: AttrWasInlined, s: ""},
	{bit: AttrIndexed, s: ""},
	{bit: AttrContentAddressable, s: ""},
	{bit: AttrABIWrapper, s: "ABIWRAPPER"},
	{bit: AttrPkgInit, s: "PKGINIT"},
}


func (a Attribute) String() string {
	var s string
	for _, x := range textAttrStrings {
		if a&x.bit != 0 {
			if x.s != "" {
				s += x.s + "|"
			}
			a &^= x.bit
		}
	}
	switch a.ABI() {
	case ABI0:
	case ABIInternal:
		s += "ABIInternal|"
		a.SetABI(0) 
	}
	if a != 0 {
		s += fmt.Sprintf("UnknownAttribute(%d)|", a)
	}
	
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}


func (s *LSym) TextAttrString() string {
	attr := s.Attribute.String()
	if s.Func().FuncFlag&abi.FuncFlagTopFrame != 0 {
		if attr != "" {
			attr += "|"
		}
		attr += "TOPFRAME"
	}
	return attr
}

func (s *LSym) String() string {
	return s.Name
}


func (*LSym) CanBeAnSSASym() {}
func (*LSym) CanBeAnSSAAux() {}

type Pcln struct {
	
	Pcsp      *LSym
	Pcfile    *LSym
	Pcline    *LSym
	Pcinline  *LSym
	Pcdata    []*LSym
	Funcdata  []*LSym
	UsedFiles map[goobj.CUFileIndex]struct{} 
	InlTree   InlTree                        
}

type Reloc struct {
	Off  int32
	Siz  uint8
	Type objabi.RelocType
	Add  int64
	Sym  *LSym
}

type Auto struct {
	Asym    *LSym
	Aoffset int32
	Name    AddrName
	Gotype  *LSym
}






type RegSpill struct {
	Addr           Addr
	Reg            int16
	Spill, Unspill As
}



type Link struct {
	Headtype           objabi.HeadType
	Arch               *LinkArch
	Debugasm           int
	Debugvlog          bool
	Debugpcln          string
	Flag_shared        bool
	Flag_dynlink       bool
	Flag_linkshared    bool
	Flag_optimize      bool
	Flag_locationlists bool
	Flag_noRefName     bool   
	Retpoline          bool   
	Flag_maymorestack  string 
	Bso                *bufio.Writer
	Pathname           string
	Pkgpath            string           
	hashmu             sync.Mutex       
	hash               map[string]*LSym 
	funchash           map[string]*LSym 
	statichash         map[string]*LSym 
	PosTable           src.PosTable
	InlTree            InlTree 
	DwFixups           *DwarfFixupTable
	Imports            []goobj.ImportedPkg
	DiagFunc           func(string, ...interface{})
	DiagFlush          func()
	DebugInfo          func(fn *LSym, info *LSym, curfn interface{}) ([]dwarf.Scope, dwarf.InlCalls, src.XPos) 
	GenAbstractFunc    func(fn *LSym)
	Errors             int

	InParallel    bool 
	UseBASEntries bool 
	IsAsm         bool 

	
	Text []*LSym
	Data []*LSym

	
	
	
	
	constSyms []*LSym

	
	
	pkgIdx map[string]int32

	defs         []*LSym 
	hashed64defs []*LSym 
	hasheddefs   []*LSym 
	nonpkgdefs   []*LSym 
	nonpkgrefs   []*LSym 

	Fingerprint goobj.FingerprintType 
}

func (ctxt *Link) Diag(format string, args ...interface{}) {
	ctxt.Errors++
	ctxt.DiagFunc(format, args...)
}

func (ctxt *Link) Logf(format string, args ...interface{}) {
	fmt.Fprintf(ctxt.Bso, format, args...)
	ctxt.Bso.Flush()
}



func (fi *FuncInfo) SpillRegisterArgs(last *Prog, pa ProgAlloc) *Prog {
	
	for _, ra := range fi.spills {
		spill := Appendp(last, pa)
		spill.As = ra.Spill
		spill.From.Type = TYPE_REG
		spill.From.Reg = ra.Reg
		spill.To = ra.Addr
		last = spill
	}
	return last
}



func (fi *FuncInfo) UnspillRegisterArgs(last *Prog, pa ProgAlloc) *Prog {
	
	for _, ra := range fi.spills {
		unspill := Appendp(last, pa)
		unspill.As = ra.Unspill
		unspill.From = ra.Addr
		unspill.To.Type = TYPE_REG
		unspill.To.Reg = ra.Reg
		last = unspill
	}
	return last
}


type LinkArch struct {
	*sys.Arch
	Init           func(*Link)
	ErrorCheck     func(*Link, *LSym)
	Preprocess     func(*Link, *LSym, ProgAlloc)
	Assemble       func(*Link, *LSym, ProgAlloc)
	Progedit       func(*Link, *Prog, ProgAlloc)
	SEH            func(*Link, *LSym) *LSym
	UnaryDst       map[As]bool 
	DWARFRegisters map[int16]int16
}
