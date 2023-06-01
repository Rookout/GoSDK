// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DWARF type information structures.
// The format is heavily biased toward C, but for simplicity
// the String methods use a pseudo-Go syntax.

// Borrowed from golang.org/x/debug/dwarf/type.go

package godwarf

import (
	"debug/dwarf"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
)

const (
	AttrGoKind          dwarf.Attr = 0x2900
	AttrGoKey           dwarf.Attr = 0x2901
	AttrGoElem          dwarf.Attr = 0x2902
	AttrGoEmbeddedField dwarf.Attr = 0x2903
	AttrGoRuntimeType   dwarf.Attr = 0x2904
	AttrGoPackageName   dwarf.Attr = 0x2905
	AttrGoDictIndex     dwarf.Attr = 0x2906
)


const (
	encAddress        = 0x01
	encBoolean        = 0x02
	encComplexFloat   = 0x03
	encFloat          = 0x04
	encSigned         = 0x05
	encSignedChar     = 0x06
	encUnsigned       = 0x07
	encUnsignedChar   = 0x08
	encImaginaryFloat = 0x09
)

const cyclicalTypeStop = "<cyclical>" 

type recCheck map[dwarf.Offset]struct{}

func (recCheck recCheck) acquire(off dwarf.Offset) (release func()) {
	if _, rec := recCheck[off]; rec {
		return nil
	}
	recCheck[off] = struct{}{}
	return func() {
		delete(recCheck, off)
	}
}

func sizeAlignToSize(sz, align int64) int64 {
	return sz
}

func sizeAlignToAlign(sz, align int64) int64 {
	return align
}



type Type interface {
	Common() *CommonType
	String() string
	Size() int64
	Align() int64

	stringIntl(recCheck) string
	sizeAlignIntl(recCheck) (int64, int64)
}




type CommonType struct {
	Index       int          
	ByteSize    int64        
	Name        string       
	ReflectKind reflect.Kind 
	Offset      dwarf.Offset 
}

func (c *CommonType) Common() *CommonType { return c }

func (c *CommonType) Size() int64                           { return c.ByteSize }
func (c *CommonType) Align() int64                          { return c.ByteSize }
func (c *CommonType) sizeAlignIntl(recCheck) (int64, int64) { return c.ByteSize, c.ByteSize }




type BasicType struct {
	CommonType
	BitSize   int64
	BitOffset int64
}

func (b *BasicType) Basic() *BasicType { return b }

func (t *BasicType) String() string { return t.stringIntl(nil) }

func (t *BasicType) stringIntl(recCheck) string {
	if t.Name != "" {
		return t.Name
	}
	return "?"
}

func (t *BasicType) Align() int64 { return t.CommonType.ByteSize }


type CharType struct {
	BasicType
}


type UcharType struct {
	BasicType
}


type IntType struct {
	BasicType
}


type UintType struct {
	BasicType
}


type FloatType struct {
	BasicType
}


type ComplexType struct {
	BasicType
}


type BoolType struct {
	BasicType
}


type AddrType struct {
	BasicType
}


type UnspecifiedType struct {
	BasicType
}




type QualType struct {
	CommonType
	Qual string
	Type Type
}

func (t *QualType) String() string { return t.stringIntl(make(recCheck)) }

func (t *QualType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	return t.Qual + " " + t.Type.stringIntl(recCheck)
}

func (t *QualType) Size() int64 { return sizeAlignToSize(t.sizeAlignIntl(make(recCheck))) }

func (t *QualType) sizeAlignIntl(recCheck recCheck) (int64, int64) {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return t.CommonType.ByteSize, t.CommonType.ByteSize
	}
	defer release()
	return t.Type.sizeAlignIntl(recCheck)
}


type ArrayType struct {
	CommonType
	Type          Type
	StrideBitSize int64 
	Count         int64 
}

func (t *ArrayType) String() string { return t.stringIntl(make(recCheck)) }

func (t *ArrayType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	return "[" + strconv.FormatInt(t.Count, 10) + "]" + t.Type.stringIntl(recCheck)
}

func (t *ArrayType) Size() int64  { return sizeAlignToSize(t.sizeAlignIntl(make(recCheck))) }
func (t *ArrayType) Align() int64 { return sizeAlignToAlign(t.sizeAlignIntl(make(recCheck))) }

func (t *ArrayType) sizeAlignIntl(recCheck recCheck) (int64, int64) {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return t.CommonType.ByteSize, 1
	}
	defer release()
	sz, align := t.Type.sizeAlignIntl(recCheck)
	if t.CommonType.ByteSize != 0 {
		return t.CommonType.ByteSize, align
	}
	return sz * t.Count, align
}


type VoidType struct {
	CommonType
}

func (t *VoidType) String() string { return t.stringIntl(nil) }

func (t *VoidType) stringIntl(recCheck) string { return "void" }


type PtrType struct {
	CommonType
	Type Type
}

func (t *PtrType) String() string { return t.stringIntl(make(recCheck)) }

func (t *PtrType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	return "*" + t.Type.stringIntl(recCheck)
}


type StructType struct {
	CommonType
	StructName string
	Kind       string 
	Field      []*StructField
	Incomplete bool 
}


type StructField struct {
	Name       string
	Type       Type
	ByteOffset int64
	ByteSize   int64
	BitOffset  int64 
	BitSize    int64 
	Embedded   bool
}

func (t *StructType) String() string { return t.stringIntl(make(recCheck)) }

func (t *StructType) stringIntl(recCheck recCheck) string {
	if t.StructName != "" {
		return t.Kind + " " + t.StructName
	}
	return t.Defn(recCheck)
}

func (t *StructType) Defn(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	s := t.Kind
	if t.StructName != "" {
		s += " " + t.StructName
	}
	if t.Incomplete {
		s += " /*incomplete*/"
		return s
	}
	s += " {"
	for i, f := range t.Field {
		if i > 0 {
			s += "; "
		}
		s += f.Name + " " + f.Type.stringIntl(recCheck)
		s += "@" + strconv.FormatInt(f.ByteOffset, 10)
		if f.BitSize > 0 {
			s += " : " + strconv.FormatInt(f.BitSize, 10)
			s += "@" + strconv.FormatInt(f.BitOffset, 10)
		}
	}
	s += "}"
	return s
}

func (t *StructType) Size() int64  { return sizeAlignToSize(t.sizeAlignIntl(make(recCheck))) }
func (t *StructType) Align() int64 { return sizeAlignToAlign(t.sizeAlignIntl(make(recCheck))) }

func (t *StructType) sizeAlignIntl(recCheck recCheck) (int64, int64) {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return t.CommonType.ByteSize, 1
	}
	defer release()
	if len(t.Field) == 0 {
		return t.CommonType.ByteSize, 1
	}
	return t.CommonType.ByteSize, sizeAlignToAlign(t.Field[0].Type.sizeAlignIntl(recCheck))
}



type SliceType struct {
	StructType
	ElemType Type
}

func (t *SliceType) String() string { return t.stringIntl(make(recCheck)) }

func (t *SliceType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	if t.Name != "" {
		return t.Name
	}
	return "[]" + t.ElemType.stringIntl(recCheck)
}



type StringType struct {
	StructType
}

func (t *StringType) String() string { return t.stringIntl(nil) }

func (t *StringType) stringIntl(recCheck recCheck) string {
	if t.Name != "" {
		return t.Name
	}
	return "string"
}


type InterfaceType struct {
	TypedefType
}

func (t *InterfaceType) String() string { return t.stringIntl(nil) }

func (t *InterfaceType) stringIntl(recCheck recCheck) string {
	if t.Name != "" {
		return t.Name
	}
	return "Interface"
}




type EnumType struct {
	CommonType
	EnumName string
	Val      []*EnumValue
}


type EnumValue struct {
	Name string
	Val  int64
}

func (t *EnumType) String() string { return t.stringIntl(nil) }

func (t *EnumType) stringIntl(recCheck recCheck) string {
	s := "enum"
	if t.EnumName != "" {
		s += " " + t.EnumName
	}
	s += " {"
	for i, v := range t.Val {
		if i > 0 {
			s += "; "
		}
		s += v.Name + "=" + strconv.FormatInt(v.Val, 10)
	}
	s += "}"
	return s
}


type FuncType struct {
	CommonType
	ReturnType Type
	ParamType  []Type
}

func (t *FuncType) String() string { return t.stringIntl(make(recCheck)) }

func (t *FuncType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	s := "func("
	for i, t := range t.ParamType {
		if i > 0 {
			s += ", "
		}
		s += t.stringIntl(recCheck)
	}
	s += ")"
	if t.ReturnType != nil {
		s += " " + t.ReturnType.stringIntl(recCheck)
	}
	return s
}


type DotDotDotType struct {
	CommonType
}

func (t *DotDotDotType) String() string { return t.stringIntl(nil) }

func (t *DotDotDotType) stringIntl(recCheck recCheck) string { return "..." }


type TypedefType struct {
	CommonType
	Type Type
}

func (t *TypedefType) String() string { return t.stringIntl(nil) }

func (t *TypedefType) stringIntl(recCheck recCheck) string { return t.Name }

func (t *TypedefType) Size() int64 { sz, _ := t.sizeAlignIntl(make(recCheck)); return sz }

func (t *TypedefType) sizeAlignIntl(recCheck recCheck) (int64, int64) {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return t.CommonType.ByteSize, t.CommonType.ByteSize
	}
	defer release()
	if t.Type == nil {
		return 0, 1
	}
	return t.Type.sizeAlignIntl(recCheck)
}



type MapType struct {
	TypedefType
	KeyType  Type
	ElemType Type
}

func (t *MapType) String() string { return t.stringIntl(make(recCheck)) }

func (t *MapType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	if t.Name != "" {
		return t.Name
	}
	return "map[" + t.KeyType.String() + "]" + t.ElemType.String()
}


type ChanType struct {
	TypedefType
	ElemType Type
}

func (t *ChanType) String() string { return t.stringIntl(make(recCheck)) }

func (t *ChanType) stringIntl(recCheck recCheck) string {
	release := recCheck.acquire(t.CommonType.Offset)
	if release == nil {
		return cyclicalTypeStop
	}
	defer release()
	if t.Name != "" {
		return t.Name
	}
	return "chan " + t.ElemType.String()
}

type ParametricType struct {
	TypedefType
	DictIndex int64
}



type UnsupportedType struct {
	CommonType
	Tag dwarf.Tag
}

func (t *UnsupportedType) stringIntl(recCheck) string {
	if t.Name != "" {
		return t.Name
	}
	return fmt.Sprintf("(unsupported type %s)", t.Tag.String())
}

func (t *UnsupportedType) String() string { return t.stringIntl(nil) }


func ReadType(d *dwarf.Data, index int, off dwarf.Offset, typeCache *sync.Map) (Type, error) {
	typ, err := readType(d, "info", d.Reader(), off, typeCache, nil)
	if typ != nil {
		typ.Common().Index = index
	}
	return typ, err
}

func getKind(e *dwarf.Entry) reflect.Kind {
	integer, _ := e.Val(AttrGoKind).(int64)
	return reflect.Kind(integer)
}

type delayedSize struct {
	ct *CommonType 
	ut Type        
}



func readType(d *dwarf.Data, name string, r *dwarf.Reader, off dwarf.Offset, typeCache *sync.Map, delayedSizes *[]delayedSize) (Type, error) {
	if t, ok := typeCache.Load(off); ok {
		return t.(Type), nil
	}
	r.Seek(off)
	e, err := r.Next()
	if err != nil {
		return nil, err
	}
	addressSize := r.AddressSize()
	if e == nil || e.Offset != off {
		return nil, dwarf.DecodeError{Name: name, Offset: off, Err: "no type at offset"}
	}

	
	
	
	
	if delayedSizes == nil {
		var delayedSizeList []delayedSize
		defer func() {
			for _, ds := range delayedSizeList {
				ds.ct.ByteSize = ds.ut.Size()
			}
		}()
		delayedSizes = &delayedSizeList
	}

	
	
	
	var typ Type

	nextDepth := 0

	
	next := func() *dwarf.Entry {
		if !e.Children {
			return nil
		}
		
		
		
		
		
		for {
			kid, err1 := r.Next()
			if err1 != nil {
				err = err1
				return nil
			}
			if kid.Tag == 0 {
				if nextDepth > 0 {
					nextDepth--
					continue
				}
				return nil
			}
			if kid.Children {
				nextDepth++
			}
			if nextDepth > 0 {
				continue
			}
			return kid
		}
	}

	
	
	typeOf := func(e *dwarf.Entry, attr dwarf.Attr) Type {
		tval := e.Val(attr)
		var t Type
		switch toff := tval.(type) {
		case dwarf.Offset:
			if t, err = readType(d, name, d.Reader(), toff, typeCache, delayedSizes); err != nil {
				return nil
			}
		case uint64:
			err = dwarf.DecodeError{Name: name, Offset: e.Offset, Err: "DWARFv4 section debug_types unsupported"}
			return nil
		default:
			
			return new(VoidType)
		}
		return t
	}

	switch e.Tag {
	case dwarf.TagArrayType:
		
		
		
		
		
		
		
		
		
		t := new(ArrayType)
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.ReflectKind = getKind(e)
		typ = t
		typeCache.Store(off, t)
		if t.Type = typeOf(e, dwarf.AttrType); err != nil {
			goto Error
		}
		if bytes, ok := e.Val(dwarf.AttrStride).(int64); ok {
			t.StrideBitSize = 8 * bytes
		} else if bits, ok := e.Val(dwarf.AttrStrideSize).(int64); ok {
			t.StrideBitSize = bits
		} else {
			
			
			t.StrideBitSize = 8 * t.Type.Size()
		}

		
		ndim := 0
		for kid := next(); kid != nil; kid = next() {
			
			
			switch kid.Tag {
			case dwarf.TagSubrangeType:
				count, ok := kid.Val(dwarf.AttrCount).(int64)
				if !ok {
					
					count, ok = kid.Val(dwarf.AttrUpperBound).(int64)
					if ok {
						count++ 
					} else {
						count = -1 
					}
				}
				if ndim == 0 {
					t.Count = count
				} else {
					
					
					t.Type = &ArrayType{Type: t.Type, Count: count}
				}
				ndim++
			case dwarf.TagEnumerationType:
				err = dwarf.DecodeError{Name: name, Offset: kid.Offset, Err: "cannot handle enumeration type as array bound"}
				goto Error
			}
		}
		if ndim == 0 {
			
			t.Count = -1
		}

	case dwarf.TagBaseType:
		
		
		
		
		
		
		
		name, _ := e.Val(dwarf.AttrName).(string)
		enc, ok := e.Val(dwarf.AttrEncoding).(int64)
		if !ok {
			err = dwarf.DecodeError{Name: name, Offset: e.Offset, Err: "missing encoding attribute for " + name}
			goto Error
		}
		switch enc {
		default:
			err = dwarf.DecodeError{Name: name, Offset: e.Offset, Err: "unrecognized encoding attribute value"}
			goto Error

		case encAddress:
			typ = new(AddrType)
		case encBoolean:
			typ = new(BoolType)
		case encComplexFloat:
			typ = new(ComplexType)
			if name == "complex" {
				
				
				
				switch byteSize, _ := e.Val(dwarf.AttrByteSize).(int64); byteSize {
				case 8:
					name = "complex float"
				case 16:
					name = "complex double"
				}
			}
		case encFloat:
			typ = new(FloatType)
		case encSigned:
			typ = new(IntType)
		case encUnsigned:
			typ = new(UintType)
		case encSignedChar:
			typ = new(CharType)
		case encUnsignedChar:
			typ = new(UcharType)
		}
		typeCache.Store(off, typ)
		t := typ.(interface {
			Basic() *BasicType
		}).Basic()
		t.Name = name
		t.BitSize, _ = e.Val(dwarf.AttrBitSize).(int64)
		t.BitOffset, _ = e.Val(dwarf.AttrBitOffset).(int64)
		t.ReflectKind = getKind(e)

	case dwarf.TagClassType, dwarf.TagStructType, dwarf.TagUnionType:
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		
		t := new(StructType)
		t.ReflectKind = getKind(e)
		switch t.ReflectKind {
		case reflect.Slice:
			slice := new(SliceType)
			typ = slice
			typeCache.Store(off, slice)
			slice.ElemType = typeOf(e, AttrGoElem)
			t = &slice.StructType
		case reflect.String:
			str := new(StringType)
			t = &str.StructType
			str.ReflectKind = reflect.String
			typ = str
		default:
			typ = t
		}
		typeCache.Store(off, typ)
		switch e.Tag {
		case dwarf.TagClassType:
			t.Kind = "class"
		case dwarf.TagStructType:
			t.Kind = "struct"
		case dwarf.TagUnionType:
			t.Kind = "union"
		}
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.StructName, _ = e.Val(dwarf.AttrName).(string)
		t.Incomplete = e.Val(dwarf.AttrDeclaration) != nil
		t.Field = make([]*StructField, 0, 8)
		var lastFieldType Type
		var lastFieldBitOffset int64
		for kid := next(); kid != nil; kid = next() {
			if kid.Tag == dwarf.TagMember {
				f := new(StructField)
				if f.Type = typeOf(kid, dwarf.AttrType); err != nil {
					goto Error
				}
				switch loc := kid.Val(dwarf.AttrDataMemberLoc).(type) {
				case []byte:
					
					
					if len(loc) == 0 {
						
						break
					}
					b := util.MakeBuf(d, util.UnknownFormat{}, "location", 0, loc)
					op_ := op.Opcode(b.Uint8())
					switch op_ {
					case op.DW_OP_plus_uconst:
						
						f.ByteOffset = int64(b.Uint())
						b.AssertEmpty()
					case op.DW_OP_consts:
						
						f.ByteOffset = b.Int()
						op_ = op.Opcode(b.Uint8())
						if op_ != op.DW_OP_plus {
							err = dwarf.DecodeError{Name: name, Offset: kid.Offset, Err: fmt.Sprintf("unexpected opcode 0x%x", op_)}
							goto Error
						}
						b.AssertEmpty()
					default:
						err = dwarf.DecodeError{Name: name, Offset: kid.Offset, Err: fmt.Sprintf("unexpected opcode 0x%x", op_)}
						goto Error
					}
					if b.Err != nil {
						err = b.Err
						goto Error
					}
				case int64:
					f.ByteOffset = loc
				}

				haveBitOffset := false
				f.Name, _ = kid.Val(dwarf.AttrName).(string)
				f.ByteSize, _ = kid.Val(dwarf.AttrByteSize).(int64)
				f.BitOffset, haveBitOffset = kid.Val(dwarf.AttrBitOffset).(int64)
				f.BitSize, _ = kid.Val(dwarf.AttrBitSize).(int64)
				f.Embedded, _ = kid.Val(AttrGoEmbeddedField).(bool)
				t.Field = append(t.Field, f)

				bito := f.BitOffset
				if !haveBitOffset {
					bito = f.ByteOffset * 8
				}
				if bito == lastFieldBitOffset && t.Kind != "union" {
					
					
					zeroArray(lastFieldType)
				}
				lastFieldType = f.Type
				lastFieldBitOffset = bito
			}
		}
		if t.Kind != "union" {
			b, ok := e.Val(dwarf.AttrByteSize).(int64)
			if ok && b*8 == lastFieldBitOffset {
				
				zeroArray(lastFieldType)
			}
		}

	case dwarf.TagConstType, dwarf.TagVolatileType, dwarf.TagRestrictType:
		
		
		
		t := new(QualType)
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.ReflectKind = getKind(e)
		typ = t
		typeCache.Store(off, t)
		if t.Type = typeOf(e, dwarf.AttrType); err != nil {
			goto Error
		}
		switch e.Tag {
		case dwarf.TagConstType:
			t.Qual = "const"
		case dwarf.TagRestrictType:
			t.Qual = "restrict"
		case dwarf.TagVolatileType:
			t.Qual = "volatile"
		}

	case dwarf.TagEnumerationType:
		
		
		
		
		
		
		
		
		t := new(EnumType)
		t.ReflectKind = getKind(e)
		typ = t
		typeCache.Store(off, t)
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.EnumName, _ = e.Val(dwarf.AttrName).(string)
		t.Val = make([]*EnumValue, 0, 8)
		for kid := next(); kid != nil; kid = next() {
			if kid.Tag == dwarf.TagEnumerator {
				f := new(EnumValue)
				f.Name, _ = kid.Val(dwarf.AttrName).(string)
				f.Val, _ = kid.Val(dwarf.AttrConstValue).(int64)
				n := len(t.Val)
				if n >= cap(t.Val) {
					val := make([]*EnumValue, n, n*2)
					copy(val, t.Val)
					t.Val = val
				}
				t.Val = t.Val[0 : n+1]
				t.Val[n] = f
			}
		}

	case dwarf.TagPointerType:
		
		
		
		
		t := new(PtrType)
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.ReflectKind = getKind(e)
		typ = t
		typeCache.Store(off, t)
		if e.Val(dwarf.AttrType) == nil {
			t.Type = &VoidType{}
			break
		}
		t.Type = typeOf(e, dwarf.AttrType)

	case dwarf.TagSubroutineType:
		
		
		
		
		
		
		
		
		
		t := new(FuncType)
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.ReflectKind = getKind(e)
		typ = t
		typeCache.Store(off, t)
		if t.ReturnType = typeOf(e, dwarf.AttrType); err != nil {
			goto Error
		}
		t.ParamType = make([]Type, 0, 8)
		for kid := next(); kid != nil; kid = next() {
			var tkid Type
			switch kid.Tag {
			default:
				continue
			case dwarf.TagFormalParameter:
				if tkid = typeOf(kid, dwarf.AttrType); err != nil {
					goto Error
				}
			case dwarf.TagUnspecifiedParameters:
				tkid = &DotDotDotType{}
			}
			t.ParamType = append(t.ParamType, tkid)
		}

	case dwarf.TagTypedef:
		
		
		
		
		
		
		
		t := new(TypedefType)
		t.ReflectKind = getKind(e)
		switch t.ReflectKind {
		case reflect.Map:
			m := new(MapType)
			typ = m
			typeCache.Store(off, typ)
			m.KeyType = typeOf(e, AttrGoKey)
			m.ElemType = typeOf(e, AttrGoElem)
			t = &m.TypedefType
		case reflect.Chan:
			c := new(ChanType)
			typ = c
			typeCache.Store(off, typ)
			c.ElemType = typeOf(e, AttrGoElem)
			t = &c.TypedefType
		case reflect.Interface:
			it := new(InterfaceType)
			typ = it
			typeCache.Store(off, it)
			t = &it.TypedefType
		default:
			if dictIndex, ok := e.Val(AttrGoDictIndex).(int64); ok {
				pt := new(ParametricType)
				pt.DictIndex = dictIndex
				typ = pt
				typeCache.Store(off, pt)
				t = &pt.TypedefType
			} else {
				typ = t
			}
		}
		typeCache.Store(off, typ)
		t.Name, _ = e.Val(dwarf.AttrName).(string)
		t.Type = typeOf(e, dwarf.AttrType)

	case dwarf.TagUnspecifiedType:
		
		
		
		t := new(UnspecifiedType)
		typ = t
		typeCache.Store(off, t)
		t.Name, _ = e.Val(dwarf.AttrName).(string)

	default:
		
		
		
		t := new(UnsupportedType)
		typ = t
		typeCache.Store(off, t)
		t.Tag = e.Tag
		t.Name, _ = e.Val(dwarf.AttrName).(string)
	}

	if err != nil {
		goto Error
	}

	typ.Common().Offset = off

	{
		b, ok := e.Val(dwarf.AttrByteSize).(int64)
		if !ok {
			b = -1
			switch t := typ.(type) {
			case *TypedefType:
				*delayedSizes = append(*delayedSizes, delayedSize{typ.Common(), t.Type})
			case *MapType:
				*delayedSizes = append(*delayedSizes, delayedSize{typ.Common(), t.Type})
			case *ChanType:
				*delayedSizes = append(*delayedSizes, delayedSize{typ.Common(), t.Type})
			case *InterfaceType:
				*delayedSizes = append(*delayedSizes, delayedSize{typ.Common(), t.Type})
			case *PtrType:
				b = int64(addressSize)
			case *FuncType:
				
				b = int64(addressSize)
			}
		}
		typ.Common().ByteSize = b
	}
	return typ, nil

Error:
	
	
	
	typeCache.Delete(off)
	return nil, err
}

func zeroArray(t Type) {
	for {
		at, ok := t.(*ArrayType)
		if !ok {
			break
		}
		at.Count = 0
		t = at.Type
	}
}
