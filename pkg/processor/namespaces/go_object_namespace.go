package namespaces

import (
	"container/list"
	"github.com/Rookout/GoSDK/pkg/config"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"reflect"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type GoObjectNamespace struct {
	Obj            interface{}
	ObjectDumpConf config.ObjectDumpConfig
}

func NewGoObjectNamespace(o interface{}) *GoObjectNamespace {
	g := &GoObjectNamespace{
		Obj:            o,
		ObjectDumpConf: config.GetDefaultDumpConfig(),
	}

	return g
}

func (g *GoObjectNamespace) GetSize(_ string, _ string) types.Namespace {
	reflectedValue := reflect.ValueOf(g.Obj)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}

	switch reflectedValue.Kind() {
	case reflect.Array:
		return NewGoObjectNamespace(reflectedValue.Len())

	case reflect.Map:
		return NewGoObjectNamespace(len(reflectedValue.MapKeys()))

	default:

		switch reflect.Zero(reflectedValue.Type()).String() {
		case "<*list.List Value>":
			l := reflectedValue.Interface().(*list.List)

			return NewGoObjectNamespace(l.Len())
		}
	}
	return nil
}

func (g *GoObjectNamespace) CallMethod(name string, args string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "type":
		if nil == g.Obj {
			return NewGoObjectNamespace("nil"), nil
		}

		reflectedValue := reflect.ValueOf(g.Obj)

		if reflectedValue.Kind() == reflect.Ptr {
			reflectedValue = reflectedValue.Elem()
		}

		x := reflectedValue.Type().String()

		return NewGoObjectNamespace(x), nil
	case "size":
		size := g.GetSize(name, args)
		if size == nil {
			return nil, rookoutErrors.NewObjectHasNoSizeException(g.GetObject())
		}
		return size, nil

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (g *GoObjectNamespace) ReadAttribute(name string) (types.Namespace, rookoutErrors.RookoutError) {
	attribute, err := utils.GetStructAttribute(g.Obj, name)
	if nil == err {
		return NewGoObjectNamespace(attribute), nil
	}
	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (g *GoObjectNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (g *GoObjectNamespace) ReadComplexKey(key interface{}) types.Namespace {
	if g.Obj == nil {
		return nil
	}

	reflectedValue := reflect.ValueOf(g.Obj)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}

	switch reflect.Zero(reflectedValue.Type()).String() {
	case "<*list.List Value>":
		l := reflectedValue.Interface().(*list.List)

		return NewGoObjectNamespace(utils.GetElementInList(l, key.(int)))
	}

	return nil
}

func (g *GoObjectNamespace) ReadKey(key interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	if g.Obj == nil {
		return nil, rookoutErrors.NewAgentKeyNotFoundException("", key, nil)
	}

	reflectedValue := reflect.ValueOf(g.Obj)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}

	switch reflectedValue.Kind() {
	case reflect.Array:
		return NewGoObjectNamespace(reflectedValue.Index(key.(int)).Interface()), nil

	case reflect.Map:
		k := reflectedValue.MapKeys()

		
		for i := 0; i < len(k); i++ {
			if key.(string) == k[i].String() {
				x := reflectedValue.MapIndex(k[i]).Interface()
				return NewGoObjectNamespace(x), nil
			}
		}

	case reflect.Struct:
		return NewGoObjectNamespace(reflectedValue.FieldByName(key.(string))), nil

	case reflect.Slice:
		return NewGoObjectNamespace(reflectedValue.Index(key.(int)).Interface()), nil

	default:

		return g.ReadComplexKey(key), nil
	}
	return nil, rookoutErrors.NewAgentKeyNotFoundException("", key, nil)
}

func (g *GoObjectNamespace) GetObject() interface{} {
	return g.Obj
}

func getGoCommonType(o interface{}) string {
	if o == nil {
		return "nil"
	}

	reflectedValue := reflect.ValueOf(o)

	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}

	t := reflectedValue.Type().String()

	switch {
	case t == "bool":
		return "bool"
	case utils.Contains(utils.IntTypes, t):
		return "int"
	case utils.Contains(utils.StringTypes, t):
		return "string"
	case utils.Contains(utils.FloatTypes, t):
		return "float"
	case utils.Contains(utils.ComplexTypes, t):
		return "complex"
	case "time.Time" == t:
		return "datetime"
	case "list.List" == t:
		return "list"
	}

	switch reflectedValue.Kind() {
	case reflect.Array:
		return "array"
	case reflect.Map:
		return "dict"
	}

	return "no common type"
}

func (g *GoObjectNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer recoverFromPanic(recover(), v, logErrors)

	if err := dumpGoObject(g.GetObject(), v, 0, g.ObjectDumpConf, logErrors); nil != err {
		return GetErrorVariant(err, logErrors)
	}

	return v
}

func (g *GoObjectNamespace) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "GoObjectNamespace",
		"@common_type":   getGoCommonType(g.Obj),
		"@original_type": getGoCommonType(g.Obj),
		"@attributes":    map[string]interface{}{},
		"@value":         g.Obj,
	}
}

func (g *GoObjectNamespace) ToSimpleDict() interface{} {
	return g.Obj
}

func (g *GoObjectNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}

func (g *GoObjectNamespace) GetObjectDumpConfig() config.ObjectDumpConfig {
	return g.ObjectDumpConf
}

func (g *GoObjectNamespace) SetObjectDumpConfig(config config.ObjectDumpConfig) {
	g.ObjectDumpConf = config
}
