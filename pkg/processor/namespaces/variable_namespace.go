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
	currentDepth      int
}

func NewVariableNamespace(fullName string, o *variable.Variable, collectionService *collection.CollectionService) *VariableNamespace {
	g := &VariableNamespace{
		Obj:               o,
		ObjectDumpConf:    o.ObjectDumpConfig,
		CollectionService: collectionService,
		name:              fullName,
		currentDepth:      0,
	}

	return g
}

func (v *VariableNamespace) spawn(name string, obj *variable.Variable, checkMaxDepth bool) (*VariableNamespace, bool) {
	if checkMaxDepth && v.currentDepth >= v.ObjectDumpConf.MaxDepth {
		return nil, true
	}

	return &VariableNamespace{name: name, Obj: obj, ObjectDumpConf: obj.ObjectDumpConfig, CollectionService: v.CollectionService, currentDepth: v.currentDepth + 1}, false
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
			attr, _ := v.spawn(v.name+"."+name, child, false)
			return attr, nil
		}
	}

	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (v *VariableNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (v *VariableNamespace) readKeyFromArray(key int) (Namespace, rookoutErrors.RookoutError) {
	name := v.name + "[" + strconv.Itoa(key) + "]"

	if int(v.Obj.Len) <= key {
		return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
	}

	var child *variable.Variable
	if len(v.Obj.Children) > key {
		child = v.Obj.Children[key]
	} else {
		
		var err error
		child, err = v.Obj.LoadArrayValue(key)
		if err == nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, err)
		}
	}

	item, _ := v.spawn(name, child, false)
	return item, nil
}

func (v *VariableNamespace) readKeyFromMap(key string) (Namespace, rookoutErrors.RookoutError) {
	name := v.name + "[\"" + key + "\"]"

	
	for i := 0; i < len(v.Obj.Children); i += 2 {
		keyVar := v.Obj.Children[i]
		if keyVar.Kind == reflect.Interface || keyVar.Kind == reflect.Ptr {
			keyVar.ObjectDumpConfig.MaxCollectionDepth = 1
			keyVar.LoadValue()
			keyVar = keyVar.Children[0]
		}
		if keyVar.Kind != reflect.String {
			continue
		}

		keyVar.LoadValue()
		keyName := constant.StringVal(keyVar.Value)
		if key == keyName {
			item, _ := v.spawn(name, v.Obj.Children[i+1], false)
			return item, nil
		}
	}

	if int(v.Obj.Len) > len(v.Obj.Children) {
		
		value, err := v.Obj.LoadMapValue(key)
		if err != nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, err)
		}
		item, _ := v.spawn(name, value, false)
		return item, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
}

func (v *VariableNamespace) readKeyFromStruct(key string) (Namespace, rookoutErrors.RookoutError) {
	name := v.name + "." + key
	for _, child := range v.Obj.Children {
		if key == child.Name {
			item, _ := v.spawn(name, child, false)
			return item, nil
		}
	}

	if int(v.Obj.Len) > len(v.Obj.Children) {
		
		obj, err := v.Obj.LoadStructValue(name)
		if err != nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, err)
		}
		item, _ := v.spawn(name, obj, false)
		return item, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(v.name, key, nil)
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

	case reflect.Interface, reflect.Ptr:
		obj, _ := v.spawn(v.name, v.Obj.Children[0], false)
		
		if v.Obj.Kind == obj.Obj.Kind {
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
			n, maxDepth := v.spawn(v.name+"["+strconv.Itoa(i)+"]", child, true)
			if maxDepth {
				return values
			}
			values = append(values, n.GetObject())
		}
		return values
	case reflect.Struct:
		return StructTypeInstance
	case reflect.Map, reflect.Chan, reflect.Interface, reflect.Ptr, reflect.Func:
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
			getKeyValue := func(i int) (Namespace, Namespace, bool) {
				keyIndex := i * 2
				valueIndex := keyIndex + 1

				key := v.Obj.Children[keyIndex]
				keyNamespace, maxDepth := v.spawn(key.Name, key, true)
				if maxDepth {
					return nil, nil, false
				}

				value := v.Obj.Children[valueIndex]
				valueNamespace, maxDepth := v.spawn(value.Name, value, true)
				if maxDepth {
					return nil, nil, false
				}

				return keyNamespace, valueNamespace, true
			}

			serializer.dumpMap(getKeyValue, len(v.Obj.Children)/2, v.ObjectDumpConf)
		case reflect.Slice, reflect.Array, reflect.Chan:
			
			if v.Obj.Base == 0 && v.Obj.Len == 0 {
				serializer.dumpNil()
				return
			}

			getElem := func(i int) (Namespace, bool) {
				if i >= len(v.Obj.Children) {
					return nil, false
				}

				child := v.Obj.Children[i]
				spawned, maxDepth := v.spawn(child.Name, child, true)
				if maxDepth {
					return nil, false
				}
				return spawned, true
			}
			serializer.dumpArray(getElem, int(v.Obj.Len), v.ObjectDumpConf)
		case reflect.Ptr:
			if len(v.Obj.Children) == 0 || v.Obj.Children[0].Addr == 0 {
				serializer.dumpNil()
			} else if v.Obj.Children[0].OnlyAddr {
				dumpUint(serializer, v.Obj.Children[0].Addr, v.ObjectDumpConf)
			} else {
				child, maxDepth := v.spawn(v.Obj.Name, v.Obj.Children[0], true)
				if maxDepth {
					return
				}
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

			var child *VariableNamespace
			if data.Addr == v.Obj.Addr {
				
				var maxDepth bool
				child, maxDepth = v.spawn(data.Name, data, true)
				if maxDepth {
					return
				}
			} else {
				child, _ = v.spawn(data.Name, data, false)
				child.currentDepth = v.currentDepth
			}

			child.Serialize(serializer)
			serializer.dumpOriginalType(fmt.Sprintf("%s (%s)", PrettyTypeName(v.Obj.DwarfType), PrettyTypeName(child.Obj.DwarfType)))
		case reflect.Struct:
			getField := func(i int) (string, Namespace, bool) {
				child := v.Obj.Children[i]
				field, maxDepth := v.spawn(child.Name, child, true)
				if maxDepth {
					return "", nil, false
				}
				return child.Name, field, true
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
