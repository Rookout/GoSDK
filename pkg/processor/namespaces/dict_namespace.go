package namespaces

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"reflect"
)

const dictCommonType = "dict"

type DictNamespace struct {
	Obj          *map[interface{}]interface{}
	originalType string
	originalSize int
	attributes   map[string]types.Namespace
}

func NewDictNamespace(
	initObj *map[interface{}]interface{},
	origSize int,
	attributes map[string]types.Namespace) types.Namespace {

	if nil == initObj {
		initObj = &map[interface{}]interface{}{}
	}

	d := &DictNamespace{
		Obj:          initObj,
		originalSize: origSize,
		attributes:   attributes,
	}

	reflectedValue := reflect.ValueOf(d.Obj)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}
	d.originalType = reflect.Zero(reflectedValue.Type()).String()

	return d
}

func (d DictNamespace) CallMethod(name string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "size":
		return NewGoObjectNamespace(len(*d.Obj)), nil

	case "type":
		return NewGoObjectNamespace(d.originalType), nil

	case "common_type":
		return NewGoObjectNamespace(dictCommonType), nil

	case "original_size":
		return NewGoObjectNamespace(d.originalSize), nil

	case "max_depth":
		return NewGoObjectNamespace(false), nil

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (d DictNamespace) ReadKey(key interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	obj, ok := (*d.Obj)[key]
	if ok == true {
		return obj.(types.Namespace), nil
	}
	return nil, rookoutErrors.NewAgentKeyNotFoundException("", key, nil)
}

func (d DictNamespace) ReadAttribute(_ string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (d DictNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (d DictNamespace) GetObject() interface{} {
	return d.Obj
}

func (d DictNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer recoverFromPanic(recover(), v, logErrors)

	v.VariantType = pb.Variant_VARIANT_MAP

	pairs := make([]*pb.Variant_Pair, 0)
	for k, val := range *d.Obj {
		dumpedK := k.(types.Namespace).ToProtobuf(logErrors)

		dumpedVal := val.(types.Namespace).ToProtobuf(logErrors)

		pairs = append(pairs, &pb.Variant_Pair{
			First:  dumpedK,
			Second: dumpedVal,
		})
	}

	v.Value = &pb.Variant_MapValue{
		MapValue: &pb.Variant_Map{
			OriginalSize: int32(len(*d.Obj)),
			Pairs:        pairs}}

	return v
}

func (d DictNamespace) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "DictNamespace",
		"@common_type":   dictCommonType,
		"@original_type": d.originalType,
		"@original_size": d.originalSize,
		"@max_depth":     false,
		"@attributes":    getAttributesDict(d.attributes),
		"@value":         d.Obj,
	}
}

func (d DictNamespace) ToSimpleDict() interface{} {
	return d.Obj
}

func (d DictNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}
