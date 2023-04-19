package namespaces

import (
	"fmt"
	"go/constant"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)



type ReferenceType struct{}

var ReferenceTypeInstance = &ReferenceType{}



type StructType struct{}

var StructTypeInstance = &StructType{}

type VariableNamespace struct {
	Obj               *variable.Variable
	ObjectDumpConf    config.ObjectDumpConfig
	name              string
	CollectionService *collection.CollectionService
}

func NewVariableNamespace(fullName string, o *variable.Variable, collectionService *collection.CollectionService) *VariableNamespace {
	g := &VariableNamespace{
		Obj:               o,
		ObjectDumpConf:    config.GetDefaultDumpConfig(),
		CollectionService: collectionService,
		name:              fullName,
	}

	return g
}

func (v *VariableNamespace) spawn(name string, obj *variable.Variable) *VariableNamespace {
	return NewVariableNamespace(name, obj, v.CollectionService)
}

func (v *VariableNamespace) GetSize(_ string, _ interface{}) Namespace {
	switch v.Obj.Kind {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return NewGoObjectNamespace(v.Obj.Len)
	}
	return nil
}

func (v *VariableNamespace) CallMethod(name string, args string) (Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "type":
		return NewGoObjectNamespace(PrettyTypeName(v.Obj.DwarfType)), nil
	case "size":
		size := v.GetSize(name, args)
		if size == nil {
			return nil, rookoutErrors.NewObjectHasNoSizeException(v.GetObject())
		}
		return size, nil
	case "depth":
		maxDepth, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("depth()", args)
		}
		v.ObjectDumpConf.MaxDepth = maxDepth
		return v, nil
	case "width":
		maxWidth, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("width()", args)
		}
		v.ObjectDumpConf.MaxWidth = maxWidth
		return v, nil
	case "collection_dump":
		maxCollectionDepth, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("collection_dump()", args)
		}
		v.ObjectDumpConf.MaxCollectionDepth = maxCollectionDepth
		return v, nil
	case "string":
		maxString, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("string()", args)
		}
		v.ObjectDumpConf.MaxString = maxString
		return v, nil
	case "limit":
		if objectDumpConfig, ok := config.GetObjectDumpConfig(strings.ToLower(args)); ok {
			v.ObjectDumpConf = objectDumpConfig
			return v, nil
		}
		return nil, rookoutErrors.NewRookInvalidMethodArguments("limit()", args)

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (v *VariableNamespace) ReadAttribute(name string) (Namespace, rookoutErrors.RookoutError) {
	value := v.Obj

	
	if value.Kind == reflect.Interface {
		value = value.Children[0]
	}

	
	if value.Kind == reflect.Ptr && len(value.Children) == 1 && value.Children[0].Kind == reflect.Struct {
		value = value.Children[0]
	}

	for _, child := range value.Children {
		if name == child.Name {
			return v.spawn(v.name+"."+name, child), nil
		}
	}

	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (v *VariableNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (v *VariableNamespace) readKeyFromArray(key int) (Namespace, rookoutErrors.RookoutError) {
	name := v.name + "[" + strconv.Itoa(key) + "]"

	if int(v.Obj.Len) > key {
		if len(v.Obj.Children) > key {
			return v.spawn(name, v.Obj.Children[key]), nil
		}

		
		obj, err := v.tryLoadChild(name)
		if err == nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, err)
		}
		return obj, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
}

func (v *VariableNamespace) readKeyFromMap(key string) (Namespace, rookoutErrors.RookoutError) {
	name := v.name + "[\"" + key + "\"]"

	
	for i := 0; i < len(v.Obj.Children); i += 2 {
		childKey := v.Obj.Children[i]
		keyName := constant.StringVal(childKey.Value)
		if key == keyName {
			return v.spawn(name, v.Obj.Children[i+1]), nil
		}
	}

	if int(v.Obj.Len) > len(v.Obj.Children) {
		
		obj, err := v.tryLoadChild(name)
		if err != nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, err)
		}
		return obj, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
}

func (v *VariableNamespace) readKeyFromStruct(key string) (Namespace, rookoutErrors.RookoutError) {
	name := v.name + "." + key
	for _, child := range v.Obj.Children {
		if key == child.Name {
			return v.spawn(name, child), nil
		}
	}

	if int(v.Obj.Len) > len(v.Obj.Children) {
		
		obj, err := v.tryLoadChild(name)
		if err != nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, err)
		}
		return obj, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
}

func (v *VariableNamespace) tryLoadChild(name string) (Namespace, error) {
	objectDumpConfig := config.GetTailoredLimits(v.GetObject())
	child, err := v.CollectionService.GetVariable(name, objectDumpConfig)
	if err != nil {
		return nil, err
	}
	return v.spawn(name, child), nil
}

func (v *VariableNamespace) ReadKey(key interface{}) (Namespace, rookoutErrors.RookoutError) {
	switch v.Obj.Kind {
	case reflect.Array, reflect.Slice:
		keyAsInt, ok := key.(int)
		if !ok {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
		}
		return v.readKeyFromArray(keyAsInt)

	case reflect.Map:
		keyAsString, ok := key.(string)
		if !ok {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
		}
		return v.readKeyFromMap(keyAsString)

	case reflect.Struct:
		keyAsString, ok := key.(string)
		if !ok {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
		}
		return v.readKeyFromStruct(keyAsString)

	case reflect.Interface:
		obj := v.spawn(v.name, v.Obj.Children[0])
		if obj.Obj.Kind == reflect.Interface {
			return nil, rookoutErrors.NewInvalidInterfaceVariable(key)
		}
		return obj.ReadKey(key)
	}
	return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
}

func (v *VariableNamespace) GetObject() interface{} {
	switch v.Obj.Kind {
	case reflect.Bool:
		return constant.BoolVal(v.Obj.Value)
	case reflect.Float32:
		float, _ := constant.Float32Val(v.Obj.Value)
		return float
	case reflect.Float64:
		float, _ := constant.Float64Val(v.Obj.Value)
		return float
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, _ := constant.Int64Val(v.Obj.Value)
		switch v.Obj.Kind {
		case reflect.Int:
			return int(number)
		case reflect.Int8:
			return int8(number)
		case reflect.Int16:
			return int16(number)
		case reflect.Int32:
			return int32(number)
		case reflect.Int64:
			return int64(number)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		number, _ := constant.Uint64Val(v.Obj.Value)
		switch v.Obj.Kind {
		case reflect.Uint:
			return uint(number)
		case reflect.Uint8:
			return uint8(number)
		case reflect.Uint16:
			return uint16(number)
		case reflect.Uint32:
			return uint32(number)
		case reflect.Uint64:
			return uint64(number)
		case reflect.Uintptr:
			return uintptr(number)
		}
	case reflect.Complex64:
		str := v.Obj.Value.ExactString()
		str = strings.ReplaceAll(str, " ", "")
		number, _ := strconv.ParseComplex(str, 64)
		return complex64(number)
	case reflect.Complex128:
		str := v.Obj.Value.ExactString()
		str = strings.ReplaceAll(str, " ", "")
		number, _ := strconv.ParseComplex(str, 128)
		return number
	case reflect.String:
		return constant.StringVal(v.Obj.Value)
	case reflect.Array, reflect.Slice:
		if v.Obj.IsNil() {
			return nil
		}
		values := make([]interface{}, 0, len(v.Obj.Children))
		for i, child := range v.Obj.Children {
			n := v.spawn(v.name+"["+strconv.Itoa(i)+"]", child)
			values = append(values, n.GetObject())
		}
		return values
	case reflect.Struct:
		return StructTypeInstance
	case reflect.Map, reflect.Chan, reflect.Ptr, reflect.Interface, reflect.Func:
		if v.Obj.IsNil() {
			return nil
		}
		return ReferenceTypeInstance
	}
	return constant.Val(v.Obj.Value)
}

func PrettyTypeName(typ godwarf.Type) string {
	if typ == nil {
		return ""
	}
	if typ.Common().Name != "" {
		return typ.Common().Name
	}
	r := typ.String()
	if r == "*void" {
		return "unsafe.Pointer"
	}
	return r
}

func (v *VariableNamespace) Serialize(serializer Serializer) {
	defer serializer.dumpOriginalType(PrettyTypeName(v.Obj.DwarfType))

	if v.Obj.Value != nil {
		if cd := v.Obj.ConstDescr(); cd != "" {
			i, _ := constant.Int64Val(constant.ToInt(v.Obj.Value))
			serializer.dumpEnum(cd, int(i), PrettyTypeName(v.Obj.DwarfType))
			return
		}

		var val interface{}
		
		if v.Obj.Kind == reflect.Float32 || v.Obj.Kind == reflect.Float64 {
			val, _ = constant.Float64Val(v.Obj.Value)
		} else {
			val = constant.Val(v.Obj.Value)
		}
		dumpInterface(serializer, val, v.ObjectDumpConf)

		if v.Obj.Value.Kind() == constant.String {
			serializer.dumpStringLen(int(v.Obj.Len))
		}

		return
	}

	
	if v.Obj.DwarfType.Common().Name == "time.Time" {
		timeValue := reflect.NewAt(reflect.TypeOf(time.Time{}), unsafe.Pointer(uintptr(v.Obj.Addr)))
		dumpTimeValue(serializer, timeValue, v.ObjectDumpConf)
		return
	}

	if v.Obj.Unreadable != nil {
		dumpError(serializer, v.Obj.Unreadable)
	} else {
		switch v.Obj.Kind {
		case reflect.Map:
			getKeyValue := func(i int) (Namespace, Namespace) {
				keyIndex := i * 2
				valueIndex := keyIndex + 1

				key := v.Obj.Children[keyIndex]
				keyNamespace := v.spawn(key.Name, key)

				value := v.Obj.Children[valueIndex]
				valueNamespace := v.spawn(value.Name, value)

				return keyNamespace, valueNamespace
			}

			serializer.dumpMap(getKeyValue, len(v.Obj.Children)/2, v.ObjectDumpConf)
		case reflect.Slice, reflect.Array, reflect.Chan:
			
			if v.Obj.Base == 0 && v.Obj.Len == 0 {
				serializer.dumpNil()
				return
			}

			getElem := func(i int) Namespace {
				if i >= len(v.Obj.Children) {
					return nil
				}

				child := v.Obj.Children[i]
				return v.spawn(child.Name, child)
			}
			serializer.dumpArray(getElem, int(v.Obj.Len), v.ObjectDumpConf)
		case reflect.Ptr:
			if len(v.Obj.Children) == 0 || v.Obj.Children[0].Addr == 0 {
				serializer.dumpNil()
			} else if v.Obj.Children[0].OnlyAddr {
				dumpUint(serializer, v.Obj.Children[0].Addr, v.ObjectDumpConf)
			} else {
				child := v.spawn(v.Obj.Name, v.Obj.Children[0])
				child.Serialize(serializer)
			}
		case reflect.UnsafePointer:
			if len(v.Obj.Children) == 0 {
				serializer.dumpNil()
			}
			dumpUint(serializer, v.Obj.Children[0].Addr, v.ObjectDumpConf)
		case reflect.Func:
			serializer.dumpFunc(v.Obj.FunctionName, v.Obj.FileName, v.Obj.Line)
		case reflect.Interface:
			if v.Obj.Addr == 0 || len(v.Obj.Children) == 0 {
				
				
				serializer.dumpNil()
				return
			}

			data := v.Obj.Children[0]
			if data.OnlyAddr {
				dumpUint(serializer, data.Addr, v.ObjectDumpConf)
				return
			}

			
			child := v.spawn(data.Name, data)
			child.Serialize(serializer)
			serializer.dumpOriginalType(fmt.Sprintf("%s (%s)", PrettyTypeName(v.Obj.DwarfType), PrettyTypeName(child.Obj.DwarfType)))
		case reflect.Struct:
			getField := func(i int) (string, Namespace) {
				return v.Obj.Children[i].Name, v.spawn(v.Obj.Children[i].Name, v.Obj.Children[i])
			}
			serializer.dumpStruct(getField, len(v.Obj.Children), v.ObjectDumpConf)
		default:
			serializer.dumpUnsupported()
		}
	}
}

func (v *VariableNamespace) GetObjectDumpConfig() config.ObjectDumpConfig {
	return v.ObjectDumpConf
}

func (v *VariableNamespace) SetObjectDumpConfig(config config.ObjectDumpConfig) {
	v.ObjectDumpConf = config
}
