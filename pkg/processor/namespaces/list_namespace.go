package namespaces

import (
	"container/list"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"reflect"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

const listCommonType = "list"

type ListNamespace struct {
	Obj          *list.List
	originalType string
	originalSize int32
	attributes   map[string]types.Namespace
}

func NewListNamespace(initObj *list.List, origSize int32, attributes map[string]types.Namespace) types.Namespace {
	if nil == initObj {
		initObj = list.New()
	}

	c := &ListNamespace{
		Obj:          initObj,
		originalSize: origSize,
		attributes:   attributes,
	}

	reflectedValue := reflect.ValueOf(c.Obj)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}
	c.originalType = reflect.Zero(reflectedValue.Type()).String()

	return c
}

func (l ListNamespace) CallMethod(name string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "size":
		return NewGoObjectNamespace(l.Obj.Len()), nil

	case "type":
		return NewGoObjectNamespace(l.originalType), nil

	case "common_type":
		return NewGoObjectNamespace(listCommonType), nil

	case "original_size":
		return NewGoObjectNamespace(l.originalSize), nil

	case "max_depth":
		return NewGoObjectNamespace(false), nil
	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (l ListNamespace) ReadKey(key interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	element := utils.GetElementInList(l.Obj, key.(int))
	if nil != element {
		return element.(types.Namespace), nil
	}
	return nil, rookoutErrors.NewAgentKeyNotFoundException("", key, nil)
}

func (l ListNamespace) ReadAttribute(_ string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (l ListNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (l ListNamespace) GetObject() interface{} {
	return l.Obj
}

func (l ListNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer recoverFromPanic(recover(), v, logErrors)

	v.VariantType = pb.Variant_VARIANT_LIST
	v.OriginalType = "list"

	listVariant := &pb.Variant_List{
		Type:         "list",
		OriginalSize: (int32)(l.Obj.Len())}

	for e := l.Obj.Front(); e != nil; e = e.Next() {
		newVariant := e.Value.(types.Namespace).ToProtobuf(logErrors)
		listVariant.Values = append(listVariant.Values, newVariant)
	}

	v.Value = &pb.Variant_ListValue{
		ListValue: listVariant,
	}

	return v
}

func (l ListNamespace) ToDict() map[string]interface{} {
	dicts := make([]interface{}, l.Obj.Len())
	i := 0
	for e := l.Obj.Front(); e != nil; e = e.Next() {
		dicts[i] = NewGoObjectNamespace(e.Value).ToDict()
		i++
	}

	return map[string]interface{}{
		"@namespace":     "ListNamespace",
		"@common_type":   listCommonType,
		"@original_type": l.originalType,
		"@original_size": l.originalSize,
		"@max_depth":     false,
		"@attributes":    getAttributesDict(l.attributes),
		"@value":         dicts,
	}
}

func (l ListNamespace) ToSimpleDict() interface{} {
	simpleDicts := make([]interface{}, l.Obj.Len())
	i := 0
	for e := l.Obj.Front(); e != nil; e = e.Next() {
		namespace, ok := e.Value.(types.Namespace)

		if !ok {
			namespace = NewGoObjectNamespace(e.Value)
		}

		simpleDicts[i] = namespace.ToSimpleDict()
		i++
	}
	return simpleDicts
}

func (l ListNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}
