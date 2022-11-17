package namespaces

import (
	"github.com/Rookout/GoSDK/pkg/config"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/types"
	"go/constant"
	"reflect"
	"strconv"
	"strings"
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

func (d *VariableNamespace) spawn(name string, obj *variable.Variable) *VariableNamespace {
	return NewVariableNamespace(name, obj, d.CollectionService)
}

func (d *VariableNamespace) GetSize(_ string, _ interface{}) types.Namespace {
	switch d.Obj.Kind {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return NewGoObjectNamespace(d.Obj.Len)
	}
	return nil
}

func (d *VariableNamespace) CallMethod(name string, args string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "type":
		return NewGoObjectNamespace(PrettyTypeName(d.Obj.DwarfType)), nil
	case "size":
		size := d.GetSize(name, args)
		if size == nil {
			return nil, rookoutErrors.NewObjectHasNoSizeException(d.GetObject())
		}
		return size, nil
	case "depth":
		maxDepth, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("depth()", args)
		}
		d.ObjectDumpConf.MaxDepth = maxDepth
		return d, nil
	case "width":
		maxWidth, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("width()", args)
		}
		d.ObjectDumpConf.MaxWidth = maxWidth
		return d, nil
	case "collection_dump":
		maxCollectionDepth, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("collection_dump()", args)
		}
		d.ObjectDumpConf.MaxCollectionDepth = maxCollectionDepth
		return d, nil
	case "string":
		maxString, err := strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("string()", args)
		}
		d.ObjectDumpConf.MaxString = maxString
		return d, nil
	case "limit":
		if objectDumpConfig, ok := config.GetObjectDumpConfig(strings.ToLower(args)); ok {
			d.ObjectDumpConf = objectDumpConfig
			return d, nil
		}
		return nil, rookoutErrors.NewRookInvalidMethodArguments("limit()", args)

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (d *VariableNamespace) ReadAttribute(name string) (types.Namespace, rookoutErrors.RookoutError) {
	var children []*variable.Variable
	
	if d.Obj.Kind == reflect.Ptr {
		if len(d.Obj.Children) == 1 && d.Obj.Children[0].Kind == reflect.Struct {
			children = d.Obj.Children[0].Children
		}
	} else {
		children = d.Obj.Children
	}

	for _, child := range children {
		if name == child.Name {
			return d.spawn(d.name+"."+name, child), nil
		}
	}

	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (d *VariableNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (d *VariableNamespace) readKeyFromArray(key int) (types.Namespace, rookoutErrors.RookoutError) {
	name := d.name + "[" + strconv.Itoa(key) + "]"

	if int(d.Obj.Len) > key {
		if len(d.Obj.Children) > key {
			return d.spawn(name, d.Obj.Children[key]), nil
		}

		
		obj, err := d.tryLoadChild(name)
		if err == nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, err)
		}
		return obj, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
}

func (d *VariableNamespace) readKeyFromMap(key string) (types.Namespace, rookoutErrors.RookoutError) {
	name := d.name + "[\"" + key + "\"]"

	
	for i := 0; i < len(d.Obj.Children); i += 2 {
		childKey := d.Obj.Children[i]
		keyName := constant.StringVal(childKey.Value)
		if key == keyName {
			return d.spawn(name, d.Obj.Children[i+1]), nil
		}
	}

	if int(d.Obj.Len) > len(d.Obj.Children) {
		
		obj, err := d.tryLoadChild(name)
		if err != nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, err)
		}
		return obj, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
}

func (d *VariableNamespace) readKeyFromStruct(key string) (types.Namespace, rookoutErrors.RookoutError) {
	name := d.name + "." + key
	for _, child := range d.Obj.Children {
		if key == child.Name {
			return d.spawn(name, child), nil
		}
	}

	if int(d.Obj.Len) > len(d.Obj.Children) {
		
		obj, err := d.tryLoadChild(name)
		if err != nil {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, err)
		}
		return obj, nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
}

func (d *VariableNamespace) tryLoadChild(name string) (types.Namespace, error) {
	objectDumpConfig := config.GetTailoredLimits(d.GetObject())
	v, err := d.CollectionService.GetVariable(name, objectDumpConfig)
	if err != nil {
		return nil, err
	}
	return d.spawn(name, v), nil
}

func (d *VariableNamespace) ReadKey(key interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	switch d.Obj.Kind {
	case reflect.Array, reflect.Slice:
		keyAsInt, ok := key.(int)
		if !ok {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
		}
		return d.readKeyFromArray(keyAsInt)

	case reflect.Map:
		keyAsString, ok := key.(string)
		if !ok {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
		}
		return d.readKeyFromMap(keyAsString)

	case reflect.Struct:
		keyAsString, ok := key.(string)
		if !ok {
			return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
		}
		return d.readKeyFromStruct(keyAsString)

	case reflect.Interface:
		obj := d.spawn(d.name, d.Obj.Children[0])
		if obj.Obj.Kind == reflect.Interface {
			return nil, rookoutErrors.NewInvalidInterfaceVariable(key)
		}
		return obj.ReadKey(key)
	}
	return nil, rookoutErrors.NewAgentKeyNotFoundException(d.name, key, nil)
}

func (d *VariableNamespace) GetObject() interface{} {
	switch d.Obj.Kind {
	case reflect.Bool:
		return constant.BoolVal(d.Obj.Value)
	case reflect.Float32:
		float, _ := constant.Float32Val(d.Obj.Value)
		return float
	case reflect.Float64:
		float, _ := constant.Float64Val(d.Obj.Value)
		return float
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, _ := constant.Int64Val(d.Obj.Value)
		switch d.Obj.Kind {
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
		number, _ := constant.Uint64Val(d.Obj.Value)
		switch d.Obj.Kind {
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
		str := d.Obj.Value.ExactString()
		str = strings.ReplaceAll(str, " ", "")
		number, _ := strconv.ParseComplex(str, 64)
		return complex64(number)
	case reflect.Complex128:
		str := d.Obj.Value.ExactString()
		str = strings.ReplaceAll(str, " ", "")
		number, _ := strconv.ParseComplex(str, 128)
		return number
	case reflect.String:
		return constant.StringVal(d.Obj.Value)
	case reflect.Array, reflect.Slice:
		if d.Obj.IsNil() {
			return nil
		}
		values := make([]interface{}, 0, len(d.Obj.Children))
		for i, child := range d.Obj.Children {
			n := d.spawn(d.name+"["+strconv.Itoa(i)+"]", child)
			values = append(values, n.GetObject())
		}
		return values
	case reflect.Struct:
		return StructTypeInstance
	case reflect.Map, reflect.Chan, reflect.Ptr, reflect.Interface, reflect.Func:
		if d.Obj.IsNil() {
			return nil
		}
		return ReferenceTypeInstance
	}
	return constant.Val(d.Obj.Value)
}

func (d *VariableNamespace) getGoCommonType() string {
	switch d.Obj.Kind {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "int"
	case reflect.String:
		return "string"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Complex64, reflect.Complex128:
		return "complex"
	case reflect.Array, reflect.Slice:
		return "list"
	case reflect.Map:
		return "dict"
	case reflect.Chan:
		return "list"
	}

	switch d.prettyType() {
	case "time.Time":
		return "datetime"
	case "list.List":
		return "list"
	}

	return "no common type"
}

func (d *VariableNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer recoverFromPanic(recover(), v, logErrors)

	return dumpVariable(d.Obj)
}

func (d *VariableNamespace) prettyType() string {
	return PrettyTypeName(d.Obj.DwarfType)
}

func (d *VariableNamespace) ToDict() map[string]interface{} {
	panic("not implemented")
}

func (d *VariableNamespace) ToSimpleDict() interface{} {
	panic("not implemented")
}

func (d *VariableNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}

func (d *VariableNamespace) GetObjectDumpConfig() config.ObjectDumpConfig {
	return d.ObjectDumpConf
}

func (d *VariableNamespace) SetObjectDumpConfig(config config.ObjectDumpConfig) {
	d.ObjectDumpConf = config
}
