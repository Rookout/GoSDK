// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package abi

import (
	"unsafe"
)










type Type struct {
	Size_       uintptr
	PtrBytes    uintptr 
	Hash        uint32  
	TFlag       TFlag   
	Align_      uint8   
	FieldAlign_ uint8   
	Kind_       uint8   
	
	
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	
	
	
	GCData    *byte
	Str       NameOff 
	PtrToThis TypeOff 
}



type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Pointer
	Slice
	String
	Struct
	UnsafePointer
)

const (
	
	KindDirectIface = 1 << 5
	KindGCProg      = 1 << 6 
	KindMask        = (1 << 5) - 1
)



type TFlag uint8

const (
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	TFlagUncommon TFlag = 1 << 0

	
	
	
	
	TFlagExtraStar TFlag = 1 << 1

	
	TFlagNamed TFlag = 1 << 2

	
	
	TFlagRegularMemory TFlag = 1 << 3
)


type NameOff int32


type TypeOff int32


type TextOff int32


func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return kindNames[0]
}

var kindNames = []string{
	Invalid:       "invalid",
	Bool:          "bool",
	Int:           "int",
	Int8:          "int8",
	Int16:         "int16",
	Int32:         "int32",
	Int64:         "int64",
	Uint:          "uint",
	Uint8:         "uint8",
	Uint16:        "uint16",
	Uint32:        "uint32",
	Uint64:        "uint64",
	Uintptr:       "uintptr",
	Float32:       "float32",
	Float64:       "float64",
	Complex64:     "complex64",
	Complex128:    "complex128",
	Array:         "array",
	Chan:          "chan",
	Func:          "func",
	Interface:     "interface",
	Map:           "map",
	Pointer:       "ptr",
	Slice:         "slice",
	String:        "string",
	Struct:        "struct",
	UnsafePointer: "unsafe.Pointer",
}

func (t *Type) Kind() Kind { return Kind(t.Kind_ & KindMask) }

func (t *Type) HasName() bool {
	return t.TFlag&TFlagNamed != 0
}

func (t *Type) Pointers() bool { return t.PtrBytes != 0 }


func (t *Type) IfaceIndir() bool {
	return t.Kind_&KindDirectIface == 0
}


func (t *Type) IsDirectIface() bool {
	return t.Kind_&KindDirectIface != 0
}

func (t *Type) GcSlice(begin, end uintptr) []byte {
	return unsafeSliceFor(t.GCData, int(end))[begin:]
}


type Method struct {
	Name NameOff 
	Mtyp TypeOff 
	Ifn  TextOff 
	Tfn  TextOff 
}





type UncommonType struct {
	PkgPath NameOff 
	Mcount  uint16  
	Xcount  uint16  
	Moff    uint32  
	_       uint32  
}

func (t *UncommonType) Methods() []Method {
	if t.Mcount == 0 {
		return nil
	}
	return (*[1 << 16]Method)(addChecked(unsafe.Pointer(t), uintptr(t.Moff), "t.mcount > 0"))[:t.Mcount:t.Mcount]
}

func (t *UncommonType) ExportedMethods() []Method {
	if t.Xcount == 0 {
		return nil
	}
	return (*[1 << 16]Method)(addChecked(unsafe.Pointer(t), uintptr(t.Moff), "t.xcount > 0"))[:t.Xcount:t.Xcount]
}








func addChecked(p unsafe.Pointer, x uintptr, whySafe string) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}


type Imethod struct {
	Name NameOff 
	Typ  TypeOff 
}


type ArrayType struct {
	Type
	Elem  *Type 
	Slice *Type 
	Len   uintptr
}


func (t *Type) Len() int {
	if t.Kind() == Array {
		return int((*ArrayType)(unsafe.Pointer(t)).Len)
	}
	return 0
}

func (t *Type) Common() *Type {
	return t
}

type ChanDir int

const (
	RecvDir    ChanDir = 1 << iota         
	SendDir                                
	BothDir            = RecvDir | SendDir 
	InvalidDir ChanDir = 0
)


type ChanType struct {
	Type
	Elem *Type
	Dir  ChanDir
}

type structTypeUncommon struct {
	StructType
	u UncommonType
}


func (t *Type) ChanDir() ChanDir {
	if t.Kind() == Chan {
		ch := (*ChanType)(unsafe.Pointer(t))
		return ch.Dir
	}
	return InvalidDir
}


func (t *Type) Uncommon() *UncommonType {
	if t.TFlag&TFlagUncommon == 0 {
		return nil
	}
	switch t.Kind() {
	case Struct:
		return &(*structTypeUncommon)(unsafe.Pointer(t)).u
	case Pointer:
		type u struct {
			PtrType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Func:
		type u struct {
			FuncType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Slice:
		type u struct {
			SliceType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Array:
		type u struct {
			ArrayType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Chan:
		type u struct {
			ChanType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Map:
		type u struct {
			MapType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	case Interface:
		type u struct {
			InterfaceType
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	default:
		type u struct {
			Type
			u UncommonType
		}
		return &(*u)(unsafe.Pointer(t)).u
	}
}


func (t *Type) Elem() *Type {
	switch t.Kind() {
	case Array:
		tt := (*ArrayType)(unsafe.Pointer(t))
		return tt.Elem
	case Chan:
		tt := (*ChanType)(unsafe.Pointer(t))
		return tt.Elem
	case Map:
		tt := (*MapType)(unsafe.Pointer(t))
		return tt.Elem
	case Pointer:
		tt := (*PtrType)(unsafe.Pointer(t))
		return tt.Elem
	case Slice:
		tt := (*SliceType)(unsafe.Pointer(t))
		return tt.Elem
	}
	return nil
}


func (t *Type) StructType() *StructType {
	if t.Kind() != Struct {
		return nil
	}
	return (*StructType)(unsafe.Pointer(t))
}


func (t *Type) MapType() *MapType {
	if t.Kind() != Map {
		return nil
	}
	return (*MapType)(unsafe.Pointer(t))
}


func (t *Type) ArrayType() *ArrayType {
	if t.Kind() != Array {
		return nil
	}
	return (*ArrayType)(unsafe.Pointer(t))
}


func (t *Type) FuncType() *FuncType {
	if t.Kind() != Func {
		return nil
	}
	return (*FuncType)(unsafe.Pointer(t))
}


func (t *Type) InterfaceType() *InterfaceType {
	if t.Kind() != Interface {
		return nil
	}
	return (*InterfaceType)(unsafe.Pointer(t))
}


func (t *Type) Size() uintptr { return t.Size_ }


func (t *Type) Align() int { return int(t.Align_) }

func (t *Type) FieldAlign() int { return int(t.FieldAlign_) }

type InterfaceType struct {
	Type
	PkgPath Name      
	Methods []Imethod 
}

func (t *Type) ExportedMethods() []Method {
	ut := t.Uncommon()
	if ut == nil {
		return nil
	}
	return ut.ExportedMethods()
}

func (t *Type) NumMethod() int {
	if t.Kind() == Interface {
		tt := (*InterfaceType)(unsafe.Pointer(t))
		return tt.NumMethod()
	}
	return len(t.ExportedMethods())
}


func (t *InterfaceType) NumMethod() int { return len(t.Methods) }

type MapType struct {
	Type
	Key    *Type
	Elem   *Type
	Bucket *Type 
	
	Hasher     func(unsafe.Pointer, uintptr) uintptr
	KeySize    uint8  
	ValueSize  uint8  
	BucketSize uint16 
	Flags      uint32
}



func (mt *MapType) IndirectKey() bool { 
	return mt.Flags&1 != 0
}
func (mt *MapType) IndirectElem() bool { 
	return mt.Flags&2 != 0
}
func (mt *MapType) ReflexiveKey() bool { 
	return mt.Flags&4 != 0
}
func (mt *MapType) NeedKeyUpdate() bool { 
	return mt.Flags&8 != 0
}
func (mt *MapType) HashMightPanic() bool { 
	return mt.Flags&16 != 0
}

func (t *Type) Key() *Type {
	if t.Kind() == Map {
		return (*MapType)(unsafe.Pointer(t)).Key
	}
	return nil
}

type SliceType struct {
	Type
	Elem *Type 
}












type FuncType struct {
	Type
	InCount  uint16
	OutCount uint16 
}

func (t *FuncType) In(i int) *Type {
	return t.InSlice()[i]
}

func (t *FuncType) NumIn() int {
	return int(t.InCount)
}

func (t *FuncType) NumOut() int {
	return int(t.OutCount & (1<<15 - 1))
}

func (t *FuncType) Out(i int) *Type {
	return (t.OutSlice()[i])
}

func (t *FuncType) InSlice() []*Type {
	uadd := unsafe.Sizeof(*t)
	if t.TFlag&TFlagUncommon != 0 {
		uadd += unsafe.Sizeof(UncommonType{})
	}
	if t.InCount == 0 {
		return nil
	}
	return (*[1 << 16]*Type)(addChecked(unsafe.Pointer(t), uadd, "t.inCount > 0"))[:t.InCount:t.InCount]
}
func (t *FuncType) OutSlice() []*Type {
	outCount := uint16(t.NumOut())
	if outCount == 0 {
		return nil
	}
	uadd := unsafe.Sizeof(*t)
	if t.TFlag&TFlagUncommon != 0 {
		uadd += unsafe.Sizeof(UncommonType{})
	}
	return (*[1 << 17]*Type)(addChecked(unsafe.Pointer(t), uadd, "outCount > 0"))[t.InCount : t.InCount+outCount : t.InCount+outCount]
}

func (t *FuncType) IsVariadic() bool {
	return t.OutCount&(1<<15) != 0
}

type PtrType struct {
	Type
	Elem *Type 
}

type StructField struct {
	Name   Name    
	Typ    *Type   
	Offset uintptr 
}

func (f *StructField) Embedded() bool {
	return f.Name.IsEmbedded()
}

type StructType struct {
	Type
	PkgPath Name
	Fields  []StructField
}



























type Name struct {
	Bytes *byte
}



func (n Name) DataChecked(off int, whySafe string) *byte {
	return (*byte)(addChecked(unsafe.Pointer(n.Bytes), uintptr(off), whySafe))
}



func (n Name) Data(off int) *byte {
	return (*byte)(addChecked(unsafe.Pointer(n.Bytes), uintptr(off), "the runtime doesn't need to give you a reason"))
}


func (n Name) IsExported() bool {
	return (*n.Bytes)&(1<<0) != 0
}


func (n Name) HasTag() bool {
	return (*n.Bytes)&(1<<1) != 0
}


func (n Name) IsEmbedded() bool {
	return (*n.Bytes)&(1<<3) != 0
}



func (n Name) ReadVarint(off int) (int, int) {
	v := 0
	for i := 0; ; i++ {
		x := *n.DataChecked(off+i, "read varint")
		v += int(x&0x7f) << (7 * i)
		if x&0x80 == 0 {
			return i + 1, v
		}
	}
}


func (n Name) IsBlank() bool {
	if n.Bytes == nil {
		return false
	}
	_, l := n.ReadVarint(1)
	return l == 1 && *n.Data(2) == '_'
}




func writeVarint(buf []byte, n int) int {
	for i := 0; ; i++ {
		b := byte(n & 0x7f)
		n >>= 7
		if n == 0 {
			buf[i] = b
			return i + 1
		}
		buf[i] = b | 0x80
	}
}


func (n Name) Name() string {
	if n.Bytes == nil {
		return ""
	}
	i, l := n.ReadVarint(1)
	return unsafeStringFor(n.DataChecked(1+i, "non-empty string"), l)
}


func (n Name) Tag() string {
	if !n.HasTag() {
		return ""
	}
	i, l := n.ReadVarint(1)
	i2, l2 := n.ReadVarint(1 + i + l)
	return unsafeStringFor(n.DataChecked(1+i+l+i2, "non-empty string"), l2)
}

func NewName(n, tag string, exported, embedded bool) Name {
	if len(n) >= 1<<29 {
		panic("abi.NewName: name too long: " + n[:1024] + "...")
	}
	if len(tag) >= 1<<29 {
		panic("abi.NewName: tag too long: " + tag[:1024] + "...")
	}
	var nameLen [10]byte
	var tagLen [10]byte
	nameLenLen := writeVarint(nameLen[:], len(n))
	tagLenLen := writeVarint(tagLen[:], len(tag))

	var bits byte
	l := 1 + nameLenLen + len(n)
	if exported {
		bits |= 1 << 0
	}
	if len(tag) > 0 {
		l += tagLenLen + len(tag)
		bits |= 1 << 1
	}
	if embedded {
		bits |= 1 << 3
	}

	b := make([]byte, l)
	b[0] = bits
	copy(b[1:], nameLen[:nameLenLen])
	copy(b[1+nameLenLen:], n)
	if len(tag) > 0 {
		tb := b[1+nameLenLen+len(n):]
		copy(tb, tagLen[:tagLenLen])
		copy(tb[tagLenLen:], tag)
	}

	return Name{Bytes: &b[0]}
}
