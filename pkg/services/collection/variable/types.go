package variable

import (
	"bytes"
	"errors"
	"fmt"
	"go/constant"
	"reflect"
	"strings"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/services/collection/memory"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
)










func runtimeTypeToDIE(_type *Variable, dataAddr uint64) (typ godwarf.Type, kind int64, err error) {
	bi := _type.bi

	_type = _type.MaybeDereference()

	// go 1.11 implementation: use extended attribute in debug_info

	md := module.FindModuleDataForType(_type.Addr)
	if md != nil {
		so := imageOfPC(bi, md.GetFirstPC())
		if so != nil {
			if rtdie, ok := so.RuntimeTypeToDIE[_type.Addr-md.GetTypesAddr()]; ok {
				typ, err := godwarf.ReadType(so.Dwarf, so.Index, rtdie.Offset, &so.TypeCache)
				if err != nil {
					return nil, 0, fmt.Errorf("invalid interface type: %v", err)
				}
				if rtdie.Kind == -1 {
					if kindField := _type.loadFieldNamed("kind"); kindField != nil && kindField.Value != nil {
						rtdie.Kind, _ = constant.Int64Val(kindField.Value)
					}
				}
				return typ, rtdie.Kind, nil
			}
		}
	}

	// go1.7 to go1.10 implementation: convert runtime._type structs to type names

	if binary_info.GoVersionAfterOrEqual(1, 17) {
		
		
		
		
		
		return nil, 0, fmt.Errorf("could not resolve interface type")
	}

	typename, kind, err := nameOfRuntimeType(_type)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid interface type: %v", err)
	}

	typ, err = bi.FindType(typename)
	if err != nil {
		return nil, 0, fmt.Errorf("interface type %q not found for %#x: %v", typename, dataAddr, err)
	}

	return typ, kind, nil
}


func imageOfPC(bi *binary_info.BinaryInfo, pc uint64) *binary_info.Image {
	fn := bi.PCToFunc(pc)
	if fn != nil {
		return bi.FuncToImage(fn)
	}

	
	var so *binary_info.Image
	for i := range bi.Images {
		if int64(bi.Images[i].StaticBase) > int64(pc) {
			continue
		}
		if so == nil || int64(bi.Images[i].StaticBase) > int64(so.StaticBase) {
			so = bi.Images[i]
		}
	}
	return so
}



const (
	tflagUncommon  = 1 << 0
	tflagExtraStar = 1 << 1
	tflagNamed     = 1 << 2
)




func nameOfRuntimeType(_type *Variable) (typename string, kind int64, err error) {
	if e, ok := _type.bi.NameOfRuntimeType[_type.Addr]; ok {
		return e.Typename, e.Kind, nil
	}

	var tflag int64

	if tflagField := _type.loadFieldNamed("tflag"); tflagField != nil && tflagField.Value != nil {
		tflag, _ = constant.Int64Val(tflagField.Value)
	}
	if kindField := _type.loadFieldNamed("kind"); kindField != nil && kindField.Value != nil {
		kind, _ = constant.Int64Val(kindField.Value)
	}

	
	
	if tflag&tflagNamed != 0 {
		typename, err = nameOfNamedRuntimeType(_type, kind, tflag)
		if err == nil {
			_type.bi.NameOfRuntimeType[_type.Addr] = binary_info.NameOfRuntimeTypeEntry{Typename: typename, Kind: kind}
		}
		return typename, kind, err
	}

	typename, err = nameOfUnnamedRuntimeType(_type, kind, tflag)
	if err == nil {
		_type.bi.NameOfRuntimeType[_type.Addr] = binary_info.NameOfRuntimeTypeEntry{Typename: typename, Kind: kind}
	}
	return typename, kind, err
}

func fieldToType(_type *Variable, fieldName string) (string, error) {
	typeField, err := _type.structMember(fieldName)
	if err != nil {
		return "", err
	}
	typeField = typeField.MaybeDereference()
	typename, _, err := nameOfRuntimeType(typeField)
	return typename, err
}

func nameOfUnnamedRuntimeType(_type *Variable, kind, tflag int64) (string, error) {
	_type, err := specificRuntimeType(_type, kind)
	if err != nil {
		return "", err
	}

	
	switch reflect.Kind(kind & kindMask) {
	case reflect.Array:
		var len int64
		if lenField := _type.loadFieldNamed("len"); lenField != nil && lenField.Value != nil {
			len, _ = constant.Int64Val(lenField.Value)
		}
		elemname, err := fieldToType(_type, "elem")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[%d]%s", len, elemname), nil
	case reflect.Chan:
		elemname, err := fieldToType(_type, "elem")
		if err != nil {
			return "", err
		}
		return "chan " + elemname, nil
	case reflect.Func:
		return nameOfFuncRuntimeType(_type, tflag, true)
	case reflect.Interface:
		return nameOfInterfaceRuntimeType(_type, kind, tflag)
	case reflect.Map:
		keyname, err := fieldToType(_type, "key")
		if err != nil {
			return "", err
		}
		elemname, err := fieldToType(_type, "elem")
		if err != nil {
			return "", err
		}
		return "map[" + keyname + "]" + elemname, nil
	case reflect.Ptr:
		elemname, err := fieldToType(_type, "elem")
		if err != nil {
			return "", err
		}
		return "*" + elemname, nil
	case reflect.Slice:
		elemname, err := fieldToType(_type, "elem")
		if err != nil {
			return "", err
		}
		return "[]" + elemname, nil
	case reflect.Struct:
		return nameOfStructRuntimeType(_type, kind, tflag)
	default:
		return nameOfNamedRuntimeType(_type, kind, tflag)
	}
}





func nameOfFuncRuntimeType(_type *Variable, tflag int64, anonymous bool) (string, error) {
	rtyp, err := _type.bi.FindType("runtime._type")
	if err != nil {
		return "", err
	}
	prtyp := pointerTo(rtyp, _type.bi)

	uadd := _type.RealType.Common().ByteSize
	if ut := uncommon(_type, tflag); ut != nil {
		uadd += ut.RealType.Common().ByteSize
	}

	var inCount, outCount int64
	if inCountField := _type.loadFieldNamed("inCount"); inCountField != nil && inCountField.Value != nil {
		inCount, _ = constant.Int64Val(inCountField.Value)
	}
	if outCountField := _type.loadFieldNamed("outCount"); outCountField != nil && outCountField.Value != nil {
		outCount, _ = constant.Int64Val(outCountField.Value)
		
		outCount = outCount & (1<<15 - 1)
	}

	cursortyp := _type.spawn("", _type.Addr+uint64(uadd), prtyp, _type.Mem)
	var buf bytes.Buffer
	if anonymous {
		buf.WriteString("func(")
	} else {
		buf.WriteString("(")
	}

	for i := int64(0); i < inCount; i++ {
		argtype := cursortyp.MaybeDereference()
		cursortyp.Addr += uint64(_type.bi.PointerSize)
		argtypename, _, err := nameOfRuntimeType(argtype)
		if err != nil {
			return "", err
		}
		buf.WriteString(argtypename)
		if i != inCount-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(")")

	switch outCount {
	case 0:
		
	case 1:
		buf.WriteString(" ")
		argtype := cursortyp.MaybeDereference()
		argtypename, _, err := nameOfRuntimeType(argtype)
		if err != nil {
			return "", err
		}
		buf.WriteString(argtypename)
	default:
		buf.WriteString(" (")
		for i := int64(0); i < outCount; i++ {
			argtype := cursortyp.MaybeDereference()
			cursortyp.Addr += uint64(_type.bi.PointerSize)
			argtypename, _, err := nameOfRuntimeType(argtype)
			if err != nil {
				return "", err
			}
			buf.WriteString(argtypename)
			if i != inCount-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString(")")
	}
	return buf.String(), nil
}





const (
	imethodFieldName       = "name"
	imethodFieldItyp       = "ityp"
	interfacetypeFieldMhdr = "mhdr"
)

func resolveTypeOff(bi *binary_info.BinaryInfo, typeAddr, off uint64, mem memory.MemoryReader) (*Variable, error) {
	
	md := module.FindModuleDataForType(typeAddr)

	rtyp, err := bi.FindType("runtime._type")
	if err != nil {
		return nil, err
	}

	if md == nil {
		v, err := reflectOffsMapAccess(bi, off, mem)
		if err != nil {
			return nil, err
		}
		v.LoadValue()
		addr, _ := constant.Int64Val(v.Value)
		return v.spawn(v.Name, uint64(addr), rtyp, mem), nil
	}

	if t, ok := md.GetTypeMap()[module.TypeOff(off)]; ok {
		tVar := NewVariable("", uint64(t), nil, mem, bi, config.GetDefaultDumpConfig(), 0, map[VariablesCacheKey]VariablesCacheValue{})
		tVar.Value = constant.MakeUint64(uint64(t))
		return tVar, nil
	}

	res := md.GetTypesAddr() + off

	return NewVariable("", uint64(res), rtyp, mem, bi, config.GetDefaultDumpConfig(), 0, map[VariablesCacheKey]VariablesCacheValue{}), nil
}

func nameOfInterfaceRuntimeType(_type *Variable, kind, tflag int64) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("interface {")

	methods, _ := _type.structMember(interfacetypeFieldMhdr)
	methods.loadArrayValues(0)
	if methods.Unreadable != nil {
		return "", nil
	}

	if len(methods.Children) == 0 {
		buf.WriteString("}")
		return buf.String(), nil
	}
	buf.WriteString(" ")

	for i, im := range methods.Children {
		var methodname, methodtype string
		for i := range im.Children {
			switch im.Children[i].Name {
			case imethodFieldName:
				nameoff, _ := constant.Int64Val(im.Children[i].Value)
				var err error
				methodname, _, _, err = resolveNameOff(_type.bi, _type.Addr, uint64(nameoff), _type.Mem)
				if err != nil {
					return "", err
				}

			case imethodFieldItyp:
				typeoff, _ := constant.Int64Val(im.Children[i].Value)
				typ, err := resolveTypeOff(_type.bi, _type.Addr, uint64(typeoff), _type.Mem)
				if err != nil {
					return "", err
				}
				typ, err = specificRuntimeType(typ, int64(reflect.Func))
				if err != nil {
					return "", err
				}
				var tflag int64
				if tflagField := typ.loadFieldNamed("tflag"); tflagField != nil && tflagField.Value != nil {
					tflag, _ = constant.Int64Val(tflagField.Value)
				}
				methodtype, err = nameOfFuncRuntimeType(typ, tflag, false)
				if err != nil {
					return "", err
				}
			}
		}

		buf.WriteString(methodname)
		buf.WriteString(methodtype)

		if i != len(methods.Children)-1 {
			buf.WriteString("; ")
		} else {
			buf.WriteString(" }")
		}
	}
	return buf.String(), nil
}

func nameOfStructRuntimeType(_type *Variable, kind, tflag int64) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("struct {")

	fields, _ := _type.structMember("fields")
	fields.loadArrayValues(0)
	if fields.Unreadable != nil {
		return "", fields.Unreadable
	}

	if len(fields.Children) == 0 {
		buf.WriteString("}")
		return buf.String(), nil
	}
	buf.WriteString(" ")

	for i, field := range fields.Children {
		var fieldname, fieldtypename string
		var typeField *Variable
		isembed := false
		for i := range field.Children {
			switch field.Children[i].Name {
			case "name":
				var nameoff int64
				switch field.Children[i].Kind {
				case reflect.Struct:
					nameoff = int64(field.Children[i].fieldVariable("bytes").Children[0].Addr)
				default:
					nameoff, _ = constant.Int64Val(field.Children[i].Value)
				}

				var err error
				fieldname, _, _, err = loadName(uint64(nameoff), _type.Mem)
				if err != nil {
					return "", err
				}

			case "typ":
				typeField = field.Children[i].MaybeDereference()
				var err error
				fieldtypename, _, err = nameOfRuntimeType(typeField)
				if err != nil {
					return "", err
				}

			case "offsetAnon":
				
				
				
				
				
				
				
				offsetAnon, _ := constant.Int64Val(field.Children[i].Value)
				isembed = offsetAnon%2 != 0
			}
		}

		
		if fieldname != "" && !isembed {
			buf.WriteString(fieldname)
			buf.WriteString(" ")
		}
		buf.WriteString(fieldtypename)
		if i != len(fields.Children)-1 {
			buf.WriteString("; ")
		} else {
			buf.WriteString(" }")
		}
	}

	return buf.String(), nil
}

















func nameOfNamedRuntimeType(_type *Variable, kind, tflag int64) (typename string, err error) {
	var strOff int64
	if strField := _type.loadFieldNamed("str"); strField != nil && strField.Value != nil {
		strOff, _ = constant.Int64Val(strField.Value)
	} else {
		return "", errors.New("could not find str field")
	}

	
	
	

	typename, _, _, err = resolveNameOff(_type.bi, _type.Addr, uint64(strOff), _type.Mem)
	if err != nil {
		return "", err
	}

	if tflag&tflagExtraStar != 0 {
		typename = typename[1:]
	}

	if i := strings.Index(typename, "."); i >= 0 {
		typename = typename[i+1:]
	} else {
		return typename, nil
	}

	
	

	_type, err = specificRuntimeType(_type, kind)
	if err != nil {
		return "", err
	}

	if ut := uncommon(_type, tflag); ut != nil {
		if pkgPathField := ut.loadFieldNamed("pkgpath"); pkgPathField != nil && pkgPathField.Value != nil {
			pkgPathOff, _ := constant.Int64Val(pkgPathField.Value)
			pkgPath, _, _, err := resolveNameOff(_type.bi, _type.Addr, uint64(pkgPathOff), _type.Mem)
			if err != nil {
				return "", err
			}
			if slash := strings.LastIndex(pkgPath, "/"); slash >= 0 {
				fixedName := strings.Replace(pkgPath[slash+1:], ".", "%2e", -1)
				if fixedName != pkgPath[slash+1:] {
					pkgPath = pkgPath[:slash+1] + fixedName
				}
			}
			typename = pkgPath + "." + typename
		}
	}

	return typename, nil
}

func specificRuntimeType(_type *Variable, kind int64) (*Variable, error) {
	typ, err := typeForKind(kind, _type.bi)
	if err != nil {
		return nil, err
	}
	if typ == nil {
		return _type, nil
	}

	return _type.spawn(_type.Name, _type.Addr, typ, _type.Mem), nil
}

var kindToRuntimeTypeName = map[reflect.Kind]string{
	reflect.Array:     "runtime.arraytype",
	reflect.Chan:      "runtime.chantype",
	reflect.Func:      "runtime.functype",
	reflect.Interface: "runtime.interfacetype",
	reflect.Map:       "runtime.maptype",
	reflect.Ptr:       "runtime.ptrtype",
	reflect.Slice:     "runtime.slicetype",
	reflect.Struct:    "runtime.structtype",
}




func typeForKind(kind int64, bi *binary_info.BinaryInfo) (*godwarf.StructType, error) {
	typename, ok := kindToRuntimeTypeName[reflect.Kind(kind&kindMask)]
	if !ok {
		return nil, nil
	}
	typ, err := bi.FindType(typename)
	if err != nil {
		return nil, err
	}
	typ = resolveTypedef(typ)
	return typ.(*godwarf.StructType), nil
}


func uncommon(_type *Variable, tflag int64) *Variable {
	if tflag&tflagUncommon == 0 {
		return nil
	}

	typ, err := _type.bi.FindType("runtime.uncommontype")
	if err != nil {
		return nil
	}

	return _type.spawn(_type.Name, _type.Addr+uint64(_type.RealType.Size()), typ, _type.Mem)
}

func resolveNameOff(bi *binary_info.BinaryInfo, typeAddr, off uint64, mem memory.MemoryReader) (name, tag string, pkgpathoff int32, err error) {
	
	if md := module.FindModuleDataForType(typeAddr); md != nil {
		return loadName(md.GetTypesAddr()+off, mem)
	}

	v, err := reflectOffsMapAccess(bi, off, mem)
	if err != nil {
		return "", "", 0, err
	}

	resv := v.MaybeDereference()
	if resv.Unreadable != nil {
		return "", "", 0, resv.Unreadable
	}

	return loadName(resv.Addr, mem)
}

func reflectOffsMapAccess(bi *binary_info.BinaryInfo, off uint64, mem memory.MemoryReader) (*Variable, error) {
	v := NewVariable("", 0, nil, mem, bi, config.GetDefaultDumpConfig(), 0, map[VariablesCacheKey]VariablesCacheValue{})
	v.Value = constant.MakeUint64(uint64(uintptr(reflectOffs.m[int32(off)])))
	v.Addr = uint64(uintptr(reflectOffs.m[int32(off)]))
	return v, nil
}

type lockRankStruct struct {
}






type mutex struct {
	
	lockRankStruct
	
	
	
	key uintptr
}

//go:linkname reflectOffs runtime.reflectOffs
var reflectOffs struct {
	lock mutex
	next int32
	m    map[int32]unsafe.Pointer
	minv map[unsafe.Pointer]int32
}

const (
	
	nameflagExported = 1 << 0
	nameflagHasTag   = 1 << 1
	nameflagHasPkg   = 1 << 2
)

func loadName(addr uint64, mem memory.MemoryReader) (name, tag string, pkgpathoff int32, err error) {
	off := addr
	namedata := make([]byte, 3)
	_, err = mem.ReadMemory(namedata, off)
	off += 3
	if err != nil {
		return "", "", 0, err
	}

	namelen := uint16(namedata[1])<<8 | uint16(namedata[2])

	rawstr := make([]byte, int(namelen))
	_, err = mem.ReadMemory(rawstr, off)
	off += uint64(namelen)
	if err != nil {
		return "", "", 0, err
	}

	name = string(rawstr)

	if namedata[0]&nameflagHasTag != 0 {
		taglendata := make([]byte, 2)
		_, err = mem.ReadMemory(taglendata, off)
		off += 2
		if err != nil {
			return "", "", 0, err
		}
		taglen := uint16(taglendata[0])<<8 | uint16(taglendata[1])

		rawstr := make([]byte, int(taglen))
		_, err = mem.ReadMemory(rawstr, off)
		off += uint64(taglen)
		if err != nil {
			return "", "", 0, err
		}

		tag = string(rawstr)
	}

	if namedata[0]&nameflagHasPkg != 0 {
		pkgdata := make([]byte, 4)
		_, err = mem.ReadMemory(pkgdata, off)
		if err != nil {
			return "", "", 0, err
		}

		
		copy((*[4]byte)(unsafe.Pointer(&pkgpathoff))[:], pkgdata)
	}

	return name, tag, pkgpathoff, nil
}



func resolveParametricType(bi *binary_info.BinaryInfo, mem memory.MemoryReader, t godwarf.Type, dictAddr uint64) (godwarf.Type, error) {
	ptyp, _ := t.(*godwarf.ParametricType)
	if ptyp == nil {
		return t, nil
	}
	if dictAddr == 0 {
		return ptyp.TypedefType.Type, errors.New("parametric type without a dictionary")
	}
	rtypeAddr, err := readUintRaw(mem, dictAddr+uint64(ptyp.DictIndex*int64(bi.PointerSize)), int64(bi.PointerSize))
	if err != nil {
		return ptyp.TypedefType.Type, err
	}
	runtimeType, err := bi.FindType("runtime._type")
	if err != nil {
		return ptyp.TypedefType.Type, err
	}
	_type := NewVariable("", rtypeAddr, runtimeType, mem, bi, config.GetDefaultDumpConfig(), dictAddr, map[VariablesCacheKey]VariablesCacheValue{})

	typ, _, err := runtimeTypeToDIE(_type, 0)
	if err != nil {
		return ptyp.TypedefType.Type, err
	}

	return typ, nil
}
