// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dwarf generates DWARF debugging information.
// DWARF generation is split between the compiler and the linker,
// this package contains the shared code.
package dwarf

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
)


const InfoPrefix = "go:info."



const ConstInfoPrefix = "go:constinfo."



const CUInfoPrefix = "go:cuinfo."



const AbstractFuncSuffix = "$abstract"



var logDwarf bool


type Sym interface {
	Length(dwarfContext interface{}) int64
}


type Var struct {
	Name          string
	Abbrev        int 
	IsReturnValue bool
	IsInlFormal   bool
	DictIndex     uint16 
	StackOffset   int32
	
	
	PutLocationList func(listSym, startPC Sym)
	Scope           int32
	Type            Sym
	DeclFile        string
	DeclLine        uint
	DeclCol         uint
	InlIndex        int32 
	ChildIndex      int32 
	IsInAbstract    bool  
}







type Scope struct {
	Parent int32
	Ranges []Range
	Vars   []*Var
}


type Range struct {
	Start, End int64
}



type FnState struct {
	Name          string
	Importpath    string
	Info          Sym
	Filesym       Sym
	Loc           Sym
	Ranges        Sym
	Absfn         Sym
	StartPC       Sym
	Size          int64
	StartLine     int32
	External      bool
	Scopes        []Scope
	InlCalls      InlCalls
	UseBASEntries bool

	dictIndexToOffset []int64
}

func EnableLogging(doit bool) {
	logDwarf = doit
}



func MergeRanges(in1, in2 []Range) []Range {
	out := make([]Range, 0, len(in1)+len(in2))
	i, j := 0, 0
	for {
		var cur Range
		if i < len(in2) && j < len(in1) {
			if in2[i].Start < in1[j].Start {
				cur = in2[i]
				i++
			} else {
				cur = in1[j]
				j++
			}
		} else if i < len(in2) {
			cur = in2[i]
			i++
		} else if j < len(in1) {
			cur = in1[j]
			j++
		} else {
			break
		}

		if n := len(out); n > 0 && cur.Start <= out[n-1].End {
			out[n-1].End = cur.End
		} else {
			out = append(out, cur)
		}
	}

	return out
}


func (s *Scope) UnifyRanges(c *Scope) {
	s.Ranges = MergeRanges(s.Ranges, c.Ranges)
}



func (s *Scope) AppendRange(r Range) {
	if r.End <= r.Start {
		return
	}
	i := len(s.Ranges)
	if i > 0 && s.Ranges[i-1].End == r.Start {
		s.Ranges[i-1].End = r.End
		return
	}
	s.Ranges = append(s.Ranges, r)
}

type InlCalls struct {
	Calls []InlCall
}

type InlCall struct {
	
	InlIndex int

	
	CallFile Sym

	
	CallLine uint32

	
	AbsFunSym Sym

	
	Children []int

	
	
	InlVars []*Var

	
	Ranges []Range

	
	Root bool
}


type Context interface {
	PtrSize() int
	AddInt(s Sym, size int, i int64)
	AddBytes(s Sym, b []byte)
	AddAddress(s Sym, t interface{}, ofs int64)
	AddCURelativeAddress(s Sym, t interface{}, ofs int64)
	AddSectionOffset(s Sym, size int, t interface{}, ofs int64)
	AddDWARFAddrSectionOffset(s Sym, t interface{}, ofs int64)
	CurrentOffset(s Sym) int64
	RecordDclReference(from Sym, to Sym, dclIdx int, inlIndex int)
	RecordChildDieOffsets(s Sym, vars []*Var, offsets []int32)
	AddString(s Sym, v string)
	AddFileRef(s Sym, f interface{})
	Logf(format string, args ...interface{})
}


func AppendUleb128(b []byte, v uint64) []byte {
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		b = append(b, c)
		if c&0x80 == 0 {
			break
		}
	}
	return b
}


func AppendSleb128(b []byte, v int64) []byte {
	for {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7
		if (v != -1 || s == 0) && (v != 0 || s != 0) {
			c |= 0x80
		}
		b = append(b, c)
		if c&0x80 == 0 {
			break
		}
	}
	return b
}


var sevenbits = [...]byte{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
	0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
	0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
	0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
}



func sevenBitU(v int64) []byte {
	if uint64(v) < uint64(len(sevenbits)) {
		return sevenbits[v : v+1]
	}
	return nil
}



func sevenBitS(v int64) []byte {
	if uint64(v) <= 63 {
		return sevenbits[v : v+1]
	}
	if uint64(-v) <= 64 {
		return sevenbits[128+v : 128+v+1]
	}
	return nil
}


func Uleb128put(ctxt Context, s Sym, v int64) {
	b := sevenBitU(v)
	if b == nil {
		var encbuf [20]byte
		b = AppendUleb128(encbuf[:0], uint64(v))
	}
	ctxt.AddBytes(s, b)
}


func Sleb128put(ctxt Context, s Sym, v int64) {
	b := sevenBitS(v)
	if b == nil {
		var encbuf [20]byte
		b = AppendSleb128(encbuf[:0], v)
	}
	ctxt.AddBytes(s, b)
}


type dwAttrForm struct {
	attr uint16
	form uint8
}


const (
	DW_AT_go_kind = 0x2900
	DW_AT_go_key  = 0x2901
	DW_AT_go_elem = 0x2902
	
	
	DW_AT_go_embedded_field = 0x2903
	DW_AT_go_runtime_type   = 0x2904

	DW_AT_go_package_name = 0x2905 
	DW_AT_go_dict_index   = 0x2906 

	DW_AT_internal_location = 253 
)


const (
	DW_ABRV_NULL = iota
	DW_ABRV_COMPUNIT
	DW_ABRV_COMPUNIT_TEXTLESS
	DW_ABRV_FUNCTION
	DW_ABRV_WRAPPER
	DW_ABRV_FUNCTION_ABSTRACT
	DW_ABRV_FUNCTION_CONCRETE
	DW_ABRV_WRAPPER_CONCRETE
	DW_ABRV_INLINED_SUBROUTINE
	DW_ABRV_INLINED_SUBROUTINE_RANGES
	DW_ABRV_VARIABLE
	DW_ABRV_INT_CONSTANT
	DW_ABRV_AUTO
	DW_ABRV_AUTO_LOCLIST
	DW_ABRV_AUTO_ABSTRACT
	DW_ABRV_AUTO_CONCRETE
	DW_ABRV_AUTO_CONCRETE_LOCLIST
	DW_ABRV_PARAM
	DW_ABRV_PARAM_LOCLIST
	DW_ABRV_PARAM_ABSTRACT
	DW_ABRV_PARAM_CONCRETE
	DW_ABRV_PARAM_CONCRETE_LOCLIST
	DW_ABRV_LEXICAL_BLOCK_RANGES
	DW_ABRV_LEXICAL_BLOCK_SIMPLE
	DW_ABRV_STRUCTFIELD
	DW_ABRV_FUNCTYPEPARAM
	DW_ABRV_DOTDOTDOT
	DW_ABRV_ARRAYRANGE
	DW_ABRV_NULLTYPE
	DW_ABRV_BASETYPE
	DW_ABRV_ARRAYTYPE
	DW_ABRV_CHANTYPE
	DW_ABRV_FUNCTYPE
	DW_ABRV_IFACETYPE
	DW_ABRV_MAPTYPE
	DW_ABRV_PTRTYPE
	DW_ABRV_BARE_PTRTYPE 
	DW_ABRV_SLICETYPE
	DW_ABRV_STRINGTYPE
	DW_ABRV_STRUCTTYPE
	DW_ABRV_TYPEDECL
	DW_ABRV_DICT_INDEX
	DW_NABRV
)

type dwAbbrev struct {
	tag      uint8
	children uint8
	attr     []dwAttrForm
}

var abbrevsFinalized bool







func expandPseudoForm(form uint8) uint8 {
	
	if form != DW_FORM_udata_pseudo {
		return form
	}
	expandedForm := DW_FORM_udata
	if buildcfg.GOOS == "darwin" || buildcfg.GOOS == "ios" {
		expandedForm = DW_FORM_data4
	}
	return uint8(expandedForm)
}



func Abbrevs() []dwAbbrev {
	if abbrevsFinalized {
		return abbrevs[:]
	}
	for i := 1; i < DW_NABRV; i++ {
		for j := 0; j < len(abbrevs[i].attr); j++ {
			abbrevs[i].attr[j].form = expandPseudoForm(abbrevs[i].attr[j].form)
		}
	}
	abbrevsFinalized = true
	return abbrevs[:]
}





var abbrevs = [DW_NABRV]dwAbbrev{
	
	{0, 0, []dwAttrForm{}},

	
	{
		DW_TAG_compile_unit,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_language, DW_FORM_data1},
			{DW_AT_stmt_list, DW_FORM_sec_offset},
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_ranges, DW_FORM_sec_offset},
			{DW_AT_comp_dir, DW_FORM_string},
			{DW_AT_producer, DW_FORM_string},
			{DW_AT_go_package_name, DW_FORM_string},
		},
	},

	
	{
		DW_TAG_compile_unit,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_language, DW_FORM_data1},
			{DW_AT_comp_dir, DW_FORM_string},
			{DW_AT_producer, DW_FORM_string},
			{DW_AT_go_package_name, DW_FORM_string},
		},
	},

	
	{
		DW_TAG_subprogram,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_high_pc, DW_FORM_addr},
			{DW_AT_frame_base, DW_FORM_block1},
			{DW_AT_decl_file, DW_FORM_data4},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_external, DW_FORM_flag},
		},
	},

	
	{
		DW_TAG_subprogram,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_high_pc, DW_FORM_addr},
			{DW_AT_frame_base, DW_FORM_block1},
			{DW_AT_trampoline, DW_FORM_flag},
		},
	},

	
	{
		DW_TAG_subprogram,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_inline, DW_FORM_data1},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_external, DW_FORM_flag},
		},
	},

	
	{
		DW_TAG_subprogram,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_high_pc, DW_FORM_addr},
			{DW_AT_frame_base, DW_FORM_block1},
		},
	},

	
	{
		DW_TAG_subprogram,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_high_pc, DW_FORM_addr},
			{DW_AT_frame_base, DW_FORM_block1},
			{DW_AT_trampoline, DW_FORM_flag},
		},
	},

	
	{
		DW_TAG_inlined_subroutine,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_high_pc, DW_FORM_addr},
			{DW_AT_call_file, DW_FORM_data4},
			{DW_AT_call_line, DW_FORM_udata_pseudo}, 
		},
	},

	
	{
		DW_TAG_inlined_subroutine,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_ranges, DW_FORM_sec_offset},
			{DW_AT_call_file, DW_FORM_data4},
			{DW_AT_call_line, DW_FORM_udata_pseudo}, 
		},
	},

	
	{
		DW_TAG_variable,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_location, DW_FORM_block1},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_external, DW_FORM_flag},
		},
	},

	
	{
		DW_TAG_constant,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_const_value, DW_FORM_sdata},
		},
	},

	
	{
		DW_TAG_variable,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_block1},
		},
	},

	
	{
		DW_TAG_variable,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_sec_offset},
		},
	},

	
	{
		DW_TAG_variable,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_type, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_variable,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_block1},
		},
	},

	
	{
		DW_TAG_variable,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_sec_offset},
		},
	},

	
	{
		DW_TAG_formal_parameter,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_variable_parameter, DW_FORM_flag},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_block1},
		},
	},

	
	{
		DW_TAG_formal_parameter,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_variable_parameter, DW_FORM_flag},
			{DW_AT_decl_line, DW_FORM_udata},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_sec_offset},
		},
	},

	
	{
		DW_TAG_formal_parameter,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_variable_parameter, DW_FORM_flag},
			{DW_AT_type, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_formal_parameter,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_block1},
		},
	},

	
	{
		DW_TAG_formal_parameter,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_abstract_origin, DW_FORM_ref_addr},
			{DW_AT_location, DW_FORM_sec_offset},
		},
	},

	
	{
		DW_TAG_lexical_block,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_ranges, DW_FORM_sec_offset},
		},
	},

	
	{
		DW_TAG_lexical_block,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_low_pc, DW_FORM_addr},
			{DW_AT_high_pc, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_member,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_data_member_location, DW_FORM_udata},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_go_embedded_field, DW_FORM_flag},
		},
	},

	
	{
		DW_TAG_formal_parameter,
		DW_CHILDREN_no,

		
		[]dwAttrForm{
			{DW_AT_type, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_unspecified_parameters,
		DW_CHILDREN_no,
		[]dwAttrForm{},
	},

	
	{
		DW_TAG_subrange_type,
		DW_CHILDREN_no,

		
		[]dwAttrForm{
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_count, DW_FORM_udata},
		},
	},

	
	
	{
		DW_TAG_unspecified_type,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
		},
	},

	
	{
		DW_TAG_base_type,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_encoding, DW_FORM_data1},
			{DW_AT_byte_size, DW_FORM_data1},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	
	{
		DW_TAG_array_type,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_byte_size, DW_FORM_udata},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_typedef,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
			{DW_AT_go_elem, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_subroutine_type,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_byte_size, DW_FORM_udata},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_typedef,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_typedef,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
			{DW_AT_go_key, DW_FORM_ref_addr},
			{DW_AT_go_elem, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_pointer_type,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_pointer_type,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
		},
	},

	
	{
		DW_TAG_structure_type,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_byte_size, DW_FORM_udata},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
			{DW_AT_go_elem, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_structure_type,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_byte_size, DW_FORM_udata},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_structure_type,
		DW_CHILDREN_yes,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_byte_size, DW_FORM_udata},
			{DW_AT_go_kind, DW_FORM_data1},
			{DW_AT_go_runtime_type, DW_FORM_addr},
		},
	},

	
	{
		DW_TAG_typedef,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
		},
	},

	
	{
		DW_TAG_typedef,
		DW_CHILDREN_no,
		[]dwAttrForm{
			{DW_AT_name, DW_FORM_string},
			{DW_AT_type, DW_FORM_ref_addr},
			{DW_AT_go_dict_index, DW_FORM_udata},
		},
	},
}


func GetAbbrev() []byte {
	abbrevs := Abbrevs()
	var buf []byte
	for i := 1; i < DW_NABRV; i++ {
		
		buf = AppendUleb128(buf, uint64(i))
		buf = AppendUleb128(buf, uint64(abbrevs[i].tag))
		buf = append(buf, abbrevs[i].children)
		for _, f := range abbrevs[i].attr {
			buf = AppendUleb128(buf, uint64(f.attr))
			buf = AppendUleb128(buf, uint64(f.form))
		}
		buf = append(buf, 0, 0)
	}
	return append(buf, 0)
}









type DWAttr struct {
	Link  *DWAttr
	Atr   uint16 
	Cls   uint8  
	Value int64
	Data  interface{}
}


type DWDie struct {
	Abbrev int
	Link   *DWDie
	Child  *DWDie
	Attr   *DWAttr
	Sym    Sym
}

func putattr(ctxt Context, s Sym, abbrev int, form int, cls int, value int64, data interface{}) error {
	switch form {
	case DW_FORM_addr: 
		
		if data == nil && value == 0 {
			ctxt.AddInt(s, ctxt.PtrSize(), 0)
			break
		}
		if cls == DW_CLS_GO_TYPEREF {
			ctxt.AddSectionOffset(s, ctxt.PtrSize(), data, value)
			break
		}
		ctxt.AddAddress(s, data, value)

	case DW_FORM_block1: 
		if cls == DW_CLS_ADDRESS {
			ctxt.AddInt(s, 1, int64(1+ctxt.PtrSize()))
			ctxt.AddInt(s, 1, DW_OP_addr)
			ctxt.AddAddress(s, data, 0)
			break
		}

		value &= 0xff
		ctxt.AddInt(s, 1, value)
		p := data.([]byte)[:value]
		ctxt.AddBytes(s, p)

	case DW_FORM_block2: 
		value &= 0xffff

		ctxt.AddInt(s, 2, value)
		p := data.([]byte)[:value]
		ctxt.AddBytes(s, p)

	case DW_FORM_block4: 
		value &= 0xffffffff

		ctxt.AddInt(s, 4, value)
		p := data.([]byte)[:value]
		ctxt.AddBytes(s, p)

	case DW_FORM_block: 
		Uleb128put(ctxt, s, value)

		p := data.([]byte)[:value]
		ctxt.AddBytes(s, p)

	case DW_FORM_data1: 
		ctxt.AddInt(s, 1, value)

	case DW_FORM_data2: 
		ctxt.AddInt(s, 2, value)

	case DW_FORM_data4: 
		if cls == DW_CLS_PTR { 
			ctxt.AddDWARFAddrSectionOffset(s, data, value)
			break
		}
		ctxt.AddInt(s, 4, value)

	case DW_FORM_data8: 
		ctxt.AddInt(s, 8, value)

	case DW_FORM_sdata: 
		Sleb128put(ctxt, s, value)

	case DW_FORM_udata: 
		Uleb128put(ctxt, s, value)

	case DW_FORM_string: 
		str := data.(string)
		ctxt.AddString(s, str)
		
		for i := int64(len(str)); i < value; i++ {
			ctxt.AddInt(s, 1, 0)
		}

	case DW_FORM_flag: 
		if value != 0 {
			ctxt.AddInt(s, 1, 1)
		} else {
			ctxt.AddInt(s, 1, 0)
		}

	
	
	case DW_FORM_ref_addr: 
		fallthrough
	case DW_FORM_sec_offset: 
		if data == nil {
			return fmt.Errorf("dwarf: null reference in %d", abbrev)
		}
		ctxt.AddDWARFAddrSectionOffset(s, data, value)

	case DW_FORM_ref1, 
		DW_FORM_ref2,      
		DW_FORM_ref4,      
		DW_FORM_ref8,      
		DW_FORM_ref_udata, 

		DW_FORM_strp,     
		DW_FORM_indirect: 
		fallthrough
	default:
		return fmt.Errorf("dwarf: unsupported attribute form %d / class %d", form, cls)
	}
	return nil
}





func PutAttrs(ctxt Context, s Sym, abbrev int, attr *DWAttr) {
	abbrevs := Abbrevs()
Outer:
	for _, f := range abbrevs[abbrev].attr {
		for ap := attr; ap != nil; ap = ap.Link {
			if ap.Atr == f.attr {
				putattr(ctxt, s, abbrev, int(f.form), int(ap.Cls), ap.Value, ap.Data)
				continue Outer
			}
		}

		putattr(ctxt, s, abbrev, int(f.form), 0, 0, nil)
	}
}


func HasChildren(die *DWDie) bool {
	abbrevs := Abbrevs()
	return abbrevs[die.Abbrev].children != 0
}


func PutIntConst(ctxt Context, info, typ Sym, name string, val int64) {
	Uleb128put(ctxt, info, DW_ABRV_INT_CONSTANT)
	putattr(ctxt, info, DW_ABRV_INT_CONSTANT, DW_FORM_string, DW_CLS_STRING, int64(len(name)), name)
	putattr(ctxt, info, DW_ABRV_INT_CONSTANT, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, typ)
	putattr(ctxt, info, DW_ABRV_INT_CONSTANT, DW_FORM_sdata, DW_CLS_CONSTANT, val, nil)
}


func PutGlobal(ctxt Context, info, typ, gvar Sym, name string) {
	Uleb128put(ctxt, info, DW_ABRV_VARIABLE)
	putattr(ctxt, info, DW_ABRV_VARIABLE, DW_FORM_string, DW_CLS_STRING, int64(len(name)), name)
	putattr(ctxt, info, DW_ABRV_VARIABLE, DW_FORM_block1, DW_CLS_ADDRESS, 0, gvar)
	putattr(ctxt, info, DW_ABRV_VARIABLE, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, typ)
	putattr(ctxt, info, DW_ABRV_VARIABLE, DW_FORM_flag, DW_CLS_FLAG, 1, nil)
}




func PutBasedRanges(ctxt Context, sym Sym, ranges []Range) {
	ps := ctxt.PtrSize()
	
	for _, r := range ranges {
		ctxt.AddInt(sym, ps, r.Start)
		ctxt.AddInt(sym, ps, r.End)
	}
	
	ctxt.AddInt(sym, ps, 0)
	ctxt.AddInt(sym, ps, 0)
}



func (s *FnState) PutRanges(ctxt Context, ranges []Range) {
	ps := ctxt.PtrSize()
	sym, base := s.Ranges, s.StartPC

	if s.UseBASEntries {
		
		
		ctxt.AddInt(sym, ps, -1)
		ctxt.AddAddress(sym, base, 0)
		PutBasedRanges(ctxt, sym, ranges)
		return
	}

	
	for _, r := range ranges {
		ctxt.AddCURelativeAddress(sym, base, r.Start)
		ctxt.AddCURelativeAddress(sym, base, r.End)
	}
	
	ctxt.AddInt(sym, ps, 0)
	ctxt.AddInt(sym, ps, 0)
}




func isEmptyInlinedCall(slot int, calls *InlCalls) bool {
	ic := &calls.Calls[slot]
	if ic.InlIndex == -2 {
		return true
	}
	live := false
	for _, k := range ic.Children {
		if !isEmptyInlinedCall(k, calls) {
			live = true
		}
	}
	if len(ic.Ranges) > 0 {
		live = true
	}
	if !live {
		ic.InlIndex = -2
	}
	return !live
}



func inlChildren(slot int, calls *InlCalls) []int {
	var kids []int
	if slot != -1 {
		for _, k := range calls.Calls[slot].Children {
			if !isEmptyInlinedCall(k, calls) {
				kids = append(kids, k)
			}
		}
	} else {
		for k := 0; k < len(calls.Calls); k += 1 {
			if calls.Calls[k].Root && !isEmptyInlinedCall(k, calls) {
				kids = append(kids, k)
			}
		}
	}
	return kids
}

func inlinedVarTable(inlcalls *InlCalls) map[*Var]bool {
	vars := make(map[*Var]bool)
	for _, ic := range inlcalls.Calls {
		for _, v := range ic.InlVars {
			vars[v] = true
		}
	}
	return vars
}







func putPrunedScopes(ctxt Context, s *FnState, fnabbrev int) error {
	if len(s.Scopes) == 0 {
		return nil
	}
	scopes := make([]Scope, len(s.Scopes), len(s.Scopes))
	pvars := inlinedVarTable(&s.InlCalls)
	for k, s := range s.Scopes {
		var pruned Scope = Scope{Parent: s.Parent, Ranges: s.Ranges}
		for i := 0; i < len(s.Vars); i++ {
			_, found := pvars[s.Vars[i]]
			if !found {
				pruned.Vars = append(pruned.Vars, s.Vars[i])
			}
		}
		sort.Sort(byChildIndex(pruned.Vars))
		scopes[k] = pruned
	}

	s.dictIndexToOffset = putparamtypes(ctxt, s, scopes, fnabbrev)

	var encbuf [20]byte
	if putscope(ctxt, s, scopes, 0, fnabbrev, encbuf[:0]) < int32(len(scopes)) {
		return errors.New("multiple toplevel scopes")
	}
	return nil
}








func PutAbstractFunc(ctxt Context, s *FnState) error {

	if logDwarf {
		ctxt.Logf("PutAbstractFunc(%v)\n", s.Absfn)
	}

	abbrev := DW_ABRV_FUNCTION_ABSTRACT
	Uleb128put(ctxt, s.Absfn, int64(abbrev))

	fullname := s.Name
	if strings.HasPrefix(s.Name, "\"\".") {
		
		
		
		
		
		
		
		fullname = objabi.PathToPrefix(s.Importpath) + "." + s.Name[3:]
	}
	putattr(ctxt, s.Absfn, abbrev, DW_FORM_string, DW_CLS_STRING, int64(len(fullname)), fullname)

	
	putattr(ctxt, s.Absfn, abbrev, DW_FORM_data1, DW_CLS_CONSTANT, int64(DW_INL_inlined), nil)

	putattr(ctxt, s.Absfn, abbrev, DW_FORM_udata, DW_CLS_CONSTANT, int64(s.StartLine), nil)

	var ev int64
	if s.External {
		ev = 1
	}
	putattr(ctxt, s.Absfn, abbrev, DW_FORM_flag, DW_CLS_FLAG, ev, 0)

	
	var flattened []*Var

	
	
	var offsets []int32

	
	if len(s.Scopes) > 0 {
		
		
		
		pvars := inlinedVarTable(&s.InlCalls)
		for _, scope := range s.Scopes {
			for i := 0; i < len(scope.Vars); i++ {
				_, found := pvars[scope.Vars[i]]
				if found || !scope.Vars[i].IsInAbstract {
					continue
				}
				flattened = append(flattened, scope.Vars[i])
			}
		}
		if len(flattened) > 0 {
			sort.Sort(byChildIndex(flattened))

			if logDwarf {
				ctxt.Logf("putAbstractScope(%v): vars:", s.Info)
				for i, v := range flattened {
					ctxt.Logf(" %d:%s", i, v.Name)
				}
				ctxt.Logf("\n")
			}

			
			
			
			for _, v := range flattened {
				offsets = append(offsets, int32(ctxt.CurrentOffset(s.Absfn)))
				putAbstractVar(ctxt, s.Absfn, v)
			}
		}
	}
	ctxt.RecordChildDieOffsets(s.Absfn, flattened, offsets)

	Uleb128put(ctxt, s.Absfn, 0)
	return nil
}






func putInlinedFunc(ctxt Context, s *FnState, callIdx int) error {
	ic := s.InlCalls.Calls[callIdx]
	callee := ic.AbsFunSym

	abbrev := DW_ABRV_INLINED_SUBROUTINE_RANGES
	if len(ic.Ranges) == 1 {
		abbrev = DW_ABRV_INLINED_SUBROUTINE
	}
	Uleb128put(ctxt, s.Info, int64(abbrev))

	if logDwarf {
		ctxt.Logf("putInlinedFunc(callee=%v,abbrev=%d)\n", callee, abbrev)
	}

	
	putattr(ctxt, s.Info, abbrev, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, callee)

	if abbrev == DW_ABRV_INLINED_SUBROUTINE_RANGES {
		putattr(ctxt, s.Info, abbrev, DW_FORM_sec_offset, DW_CLS_PTR, s.Ranges.Length(ctxt), s.Ranges)
		s.PutRanges(ctxt, ic.Ranges)
	} else {
		st := ic.Ranges[0].Start
		en := ic.Ranges[0].End
		putattr(ctxt, s.Info, abbrev, DW_FORM_addr, DW_CLS_ADDRESS, st, s.StartPC)
		putattr(ctxt, s.Info, abbrev, DW_FORM_addr, DW_CLS_ADDRESS, en, s.StartPC)
	}

	
	ctxt.AddFileRef(s.Info, ic.CallFile)
	form := int(expandPseudoForm(DW_FORM_udata_pseudo))
	putattr(ctxt, s.Info, abbrev, form, DW_CLS_CONSTANT, int64(ic.CallLine), nil)

	
	vars := ic.InlVars
	sort.Sort(byChildIndex(vars))
	inlIndex := ic.InlIndex
	var encbuf [20]byte
	for _, v := range vars {
		if !v.IsInAbstract {
			continue
		}
		putvar(ctxt, s, v, callee, abbrev, inlIndex, encbuf[:0])
	}

	
	for _, sib := range inlChildren(callIdx, &s.InlCalls) {
		err := putInlinedFunc(ctxt, s, sib)
		if err != nil {
			return err
		}
	}

	Uleb128put(ctxt, s.Info, 0)
	return nil
}








func PutConcreteFunc(ctxt Context, s *FnState, isWrapper bool) error {
	if logDwarf {
		ctxt.Logf("PutConcreteFunc(%v)\n", s.Info)
	}
	abbrev := DW_ABRV_FUNCTION_CONCRETE
	if isWrapper {
		abbrev = DW_ABRV_WRAPPER_CONCRETE
	}
	Uleb128put(ctxt, s.Info, int64(abbrev))

	
	putattr(ctxt, s.Info, abbrev, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, s.Absfn)

	
	putattr(ctxt, s.Info, abbrev, DW_FORM_addr, DW_CLS_ADDRESS, 0, s.StartPC)
	putattr(ctxt, s.Info, abbrev, DW_FORM_addr, DW_CLS_ADDRESS, s.Size, s.StartPC)

	
	putattr(ctxt, s.Info, abbrev, DW_FORM_block1, DW_CLS_BLOCK, 1, []byte{DW_OP_call_frame_cfa})

	if isWrapper {
		putattr(ctxt, s.Info, abbrev, DW_FORM_flag, DW_CLS_FLAG, int64(1), 0)
	}

	
	if err := putPrunedScopes(ctxt, s, abbrev); err != nil {
		return err
	}

	
	for _, sib := range inlChildren(-1, &s.InlCalls) {
		err := putInlinedFunc(ctxt, s, sib)
		if err != nil {
			return err
		}
	}

	Uleb128put(ctxt, s.Info, 0)
	return nil
}






func PutDefaultFunc(ctxt Context, s *FnState, isWrapper bool) error {
	if logDwarf {
		ctxt.Logf("PutDefaultFunc(%v)\n", s.Info)
	}
	abbrev := DW_ABRV_FUNCTION
	if isWrapper {
		abbrev = DW_ABRV_WRAPPER
	}
	Uleb128put(ctxt, s.Info, int64(abbrev))

	
	name := s.Name
	if s.Importpath != "" {
		name = strings.Replace(name, "\"\".", objabi.PathToPrefix(s.Importpath)+".", -1)
	}

	putattr(ctxt, s.Info, DW_ABRV_FUNCTION, DW_FORM_string, DW_CLS_STRING, int64(len(name)), name)
	putattr(ctxt, s.Info, abbrev, DW_FORM_addr, DW_CLS_ADDRESS, 0, s.StartPC)
	putattr(ctxt, s.Info, abbrev, DW_FORM_addr, DW_CLS_ADDRESS, s.Size, s.StartPC)
	putattr(ctxt, s.Info, abbrev, DW_FORM_block1, DW_CLS_BLOCK, 1, []byte{DW_OP_call_frame_cfa})
	if isWrapper {
		putattr(ctxt, s.Info, abbrev, DW_FORM_flag, DW_CLS_FLAG, int64(1), 0)
	} else {
		ctxt.AddFileRef(s.Info, s.Filesym)
		putattr(ctxt, s.Info, abbrev, DW_FORM_udata, DW_CLS_CONSTANT, int64(s.StartLine), nil)

		var ev int64
		if s.External {
			ev = 1
		}
		putattr(ctxt, s.Info, abbrev, DW_FORM_flag, DW_CLS_FLAG, ev, 0)
	}

	
	if err := putPrunedScopes(ctxt, s, abbrev); err != nil {
		return err
	}

	
	for _, sib := range inlChildren(-1, &s.InlCalls) {
		err := putInlinedFunc(ctxt, s, sib)
		if err != nil {
			return err
		}
	}

	Uleb128put(ctxt, s.Info, 0)
	return nil
}


func putparamtypes(ctxt Context, s *FnState, scopes []Scope, fnabbrev int) []int64 {
	if fnabbrev == DW_ABRV_FUNCTION_CONCRETE {
		return nil
	}

	maxDictIndex := uint16(0)

	for i := range scopes {
		for _, v := range scopes[i].Vars {
			if v.DictIndex > maxDictIndex {
				maxDictIndex = v.DictIndex
			}
		}
	}

	if maxDictIndex == 0 {
		return nil
	}

	dictIndexToOffset := make([]int64, maxDictIndex)

	for i := range scopes {
		for _, v := range scopes[i].Vars {
			if v.DictIndex == 0 || dictIndexToOffset[v.DictIndex-1] != 0 {
				continue
			}

			dictIndexToOffset[v.DictIndex-1] = ctxt.CurrentOffset(s.Info)

			Uleb128put(ctxt, s.Info, int64(DW_ABRV_DICT_INDEX))
			n := fmt.Sprintf(".param%d", v.DictIndex-1)
			putattr(ctxt, s.Info, DW_ABRV_DICT_INDEX, DW_FORM_string, DW_CLS_STRING, int64(len(n)), n)
			putattr(ctxt, s.Info, DW_ABRV_DICT_INDEX, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, v.Type)
			putattr(ctxt, s.Info, DW_ABRV_DICT_INDEX, DW_FORM_udata, DW_CLS_CONSTANT, int64(v.DictIndex-1), nil)
		}
	}

	return dictIndexToOffset
}

func putscope(ctxt Context, s *FnState, scopes []Scope, curscope int32, fnabbrev int, encbuf []byte) int32 {

	if logDwarf {
		ctxt.Logf("putscope(%v,%d): vars:", s.Info, curscope)
		for i, v := range scopes[curscope].Vars {
			ctxt.Logf(" %d:%d:%s", i, v.ChildIndex, v.Name)
		}
		ctxt.Logf("\n")
	}

	for _, v := range scopes[curscope].Vars {
		putvar(ctxt, s, v, s.Absfn, fnabbrev, -1, encbuf)
	}
	this := curscope
	curscope++
	for curscope < int32(len(scopes)) {
		scope := scopes[curscope]
		if scope.Parent != this {
			return curscope
		}

		if len(scopes[curscope].Vars) == 0 {
			curscope = putscope(ctxt, s, scopes, curscope, fnabbrev, encbuf)
			continue
		}

		if len(scope.Ranges) == 1 {
			Uleb128put(ctxt, s.Info, DW_ABRV_LEXICAL_BLOCK_SIMPLE)
			putattr(ctxt, s.Info, DW_ABRV_LEXICAL_BLOCK_SIMPLE, DW_FORM_addr, DW_CLS_ADDRESS, scope.Ranges[0].Start, s.StartPC)
			putattr(ctxt, s.Info, DW_ABRV_LEXICAL_BLOCK_SIMPLE, DW_FORM_addr, DW_CLS_ADDRESS, scope.Ranges[0].End, s.StartPC)
		} else {
			Uleb128put(ctxt, s.Info, DW_ABRV_LEXICAL_BLOCK_RANGES)
			putattr(ctxt, s.Info, DW_ABRV_LEXICAL_BLOCK_RANGES, DW_FORM_sec_offset, DW_CLS_PTR, s.Ranges.Length(ctxt), s.Ranges)

			s.PutRanges(ctxt, scope.Ranges)
		}

		curscope = putscope(ctxt, s, scopes, curscope, fnabbrev, encbuf)

		Uleb128put(ctxt, s.Info, 0)
	}
	return curscope
}


func concreteVarAbbrev(varAbbrev int) int {
	switch varAbbrev {
	case DW_ABRV_AUTO:
		return DW_ABRV_AUTO_CONCRETE
	case DW_ABRV_PARAM:
		return DW_ABRV_PARAM_CONCRETE
	case DW_ABRV_AUTO_LOCLIST:
		return DW_ABRV_AUTO_CONCRETE_LOCLIST
	case DW_ABRV_PARAM_LOCLIST:
		return DW_ABRV_PARAM_CONCRETE_LOCLIST
	default:
		panic("should never happen")
	}
}


func determineVarAbbrev(v *Var, fnabbrev int) (int, bool, bool) {
	abbrev := v.Abbrev

	
	
	missing := false
	switch {
	case abbrev == DW_ABRV_AUTO_LOCLIST && v.PutLocationList == nil:
		missing = true
		abbrev = DW_ABRV_AUTO
	case abbrev == DW_ABRV_PARAM_LOCLIST && v.PutLocationList == nil:
		missing = true
		abbrev = DW_ABRV_PARAM
	}

	
	concrete := true
	switch fnabbrev {
	case DW_ABRV_FUNCTION, DW_ABRV_WRAPPER:
		concrete = false
	case DW_ABRV_FUNCTION_CONCRETE, DW_ABRV_WRAPPER_CONCRETE:
		
		
		
		if !v.IsInAbstract {
			concrete = false
		}
	case DW_ABRV_INLINED_SUBROUTINE, DW_ABRV_INLINED_SUBROUTINE_RANGES:
	default:
		panic("should never happen")
	}

	
	if concrete {
		abbrev = concreteVarAbbrev(abbrev)
	}

	return abbrev, missing, concrete
}

func abbrevUsesLoclist(abbrev int) bool {
	switch abbrev {
	case DW_ABRV_AUTO_LOCLIST, DW_ABRV_AUTO_CONCRETE_LOCLIST,
		DW_ABRV_PARAM_LOCLIST, DW_ABRV_PARAM_CONCRETE_LOCLIST:
		return true
	default:
		return false
	}
}


func putAbstractVar(ctxt Context, info Sym, v *Var) {
	
	abbrev := v.Abbrev
	switch abbrev {
	case DW_ABRV_AUTO, DW_ABRV_AUTO_LOCLIST:
		abbrev = DW_ABRV_AUTO_ABSTRACT
	case DW_ABRV_PARAM, DW_ABRV_PARAM_LOCLIST:
		abbrev = DW_ABRV_PARAM_ABSTRACT
	}

	Uleb128put(ctxt, info, int64(abbrev))
	putattr(ctxt, info, abbrev, DW_FORM_string, DW_CLS_STRING, int64(len(v.Name)), v.Name)

	
	if abbrev == DW_ABRV_PARAM_ABSTRACT {
		var isReturn int64
		if v.IsReturnValue {
			isReturn = 1
		}
		putattr(ctxt, info, abbrev, DW_FORM_flag, DW_CLS_FLAG, isReturn, nil)
	}

	
	if abbrev != DW_ABRV_PARAM_ABSTRACT {
		
		putattr(ctxt, info, abbrev, DW_FORM_udata, DW_CLS_CONSTANT, int64(v.DeclLine), nil)
	}

	
	putattr(ctxt, info, abbrev, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, v.Type)

	
}

func putvar(ctxt Context, s *FnState, v *Var, absfn Sym, fnabbrev, inlIndex int, encbuf []byte) {
	
	abbrev, missing, concrete := determineVarAbbrev(v, fnabbrev)

	Uleb128put(ctxt, s.Info, int64(abbrev))

	
	if concrete {
		
		
		
		
		putattr(ctxt, s.Info, abbrev, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, absfn)
		ctxt.RecordDclReference(s.Info, absfn, int(v.ChildIndex), inlIndex)
	} else {
		
		n := v.Name
		putattr(ctxt, s.Info, abbrev, DW_FORM_string, DW_CLS_STRING, int64(len(n)), n)
		if abbrev == DW_ABRV_PARAM || abbrev == DW_ABRV_PARAM_LOCLIST || abbrev == DW_ABRV_PARAM_ABSTRACT {
			var isReturn int64
			if v.IsReturnValue {
				isReturn = 1
			}
			putattr(ctxt, s.Info, abbrev, DW_FORM_flag, DW_CLS_FLAG, isReturn, nil)
		}
		putattr(ctxt, s.Info, abbrev, DW_FORM_udata, DW_CLS_CONSTANT, int64(v.DeclLine), nil)
		if v.DictIndex > 0 && s.dictIndexToOffset != nil && s.dictIndexToOffset[v.DictIndex-1] != 0 {
			
			putattr(ctxt, s.Info, abbrev, DW_FORM_ref_addr, DW_CLS_REFERENCE, s.dictIndexToOffset[v.DictIndex-1], s.Info)
		} else {
			putattr(ctxt, s.Info, abbrev, DW_FORM_ref_addr, DW_CLS_REFERENCE, 0, v.Type)
		}
	}

	if abbrevUsesLoclist(abbrev) {
		putattr(ctxt, s.Info, abbrev, DW_FORM_sec_offset, DW_CLS_PTR, s.Loc.Length(ctxt), s.Loc)
		v.PutLocationList(s.Loc, s.StartPC)
	} else {
		loc := encbuf[:0]
		switch {
		case missing:
			break 
		case v.StackOffset == 0:
			loc = append(loc, DW_OP_call_frame_cfa)
		default:
			loc = append(loc, DW_OP_fbreg)
			loc = AppendSleb128(loc, int64(v.StackOffset))
		}
		putattr(ctxt, s.Info, abbrev, DW_FORM_block1, DW_CLS_BLOCK, int64(len(loc)), loc)
	}

	
}


type byChildIndex []*Var

func (s byChildIndex) Len() int           { return len(s) }
func (s byChildIndex) Less(i, j int) bool { return s[i].ChildIndex < s[j].ChildIndex }
func (s byChildIndex) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }





func IsDWARFEnabledOnAIXLd(extld []string) (bool, error) {
	name, args := extld[0], extld[1:]
	args = append(args, "-Wl,-V")
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		
		
		
		if !bytes.Contains(out, []byte("0711-317")) {
			return false, fmt.Errorf("%s -Wl,-V failed: %v\n%s", extld, err, out)
		}
	}
	
	
	
	out = bytes.TrimPrefix(out, []byte("/usr/bin/ld: LD "))
	vers := string(bytes.Split(out, []byte("("))[0])
	subvers := strings.Split(vers, ".")
	if len(subvers) != 3 {
		return false, fmt.Errorf("cannot parse %s -Wl,-V (%s): %v\n", extld, out, err)
	}
	if v, err := strconv.Atoi(subvers[0]); err != nil || v < 7 {
		return false, nil
	} else if v > 7 {
		return true, nil
	}
	if v, err := strconv.Atoi(subvers[1]); err != nil || v < 2 {
		return false, nil
	} else if v > 2 {
		return true, nil
	}
	if v, err := strconv.Atoi(subvers[2]); err != nil || v < 2 {
		return false, nil
	}
	return true, nil
}
