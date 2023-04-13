package namespaces

import (
	"container/list"
	"reflect"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/config"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type ValueNamespace struct {
	value          reflect.Value
	ObjectDumpConf config.ObjectDumpConfig
}

func NewValueNamespace(value reflect.Value) *ValueNamespace {
	g := &ValueNamespace{
		value:          value,
		ObjectDumpConf: config.GetDefaultDumpConfig(),
	}

	return g
}

func (v *ValueNamespace) GetSize(_ string, _ string) Namespace {
	value := v.value
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Array:
		return NewGoObjectNamespace(value.Len())

	case reflect.Map:
		return NewGoObjectNamespace(len(value.MapKeys()))

	default:

		switch reflect.Zero(value.Type()).String() {
		case "<*list.List Value>":
			l := value.Interface().(*list.List)

			return NewGoObjectNamespace(l.Len())
		}
	}
	return nil
}

func (v *ValueNamespace) CallMethod(name string, args string) (n Namespace, err rookoutErrors.RookoutError) {
	defer func() {
		if recover() != nil {
			n = NewGoObjectNamespace("nil")
		}
	}()

	switch name {
	case "type":
		return NewGoObjectNamespace(v.value.Type().String()), nil
	case "size":
		size := v.GetSize(name, args)
		if size == nil {
			return nil, rookoutErrors.NewObjectHasNoSizeException(v.value.Type().String())
		}
		return size, nil

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (v *ValueNamespace) ReadAttribute(name string) (n Namespace, err rookoutErrors.RookoutError) {
	value := v.value
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	_, ok := value.Type().FieldByName(name)
	if !ok {
		return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
	}

	return NewValueNamespace(value.FieldByName(name)), nil
}

func (v *ValueNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (v *ValueNamespace) ReadKey(key interface{}) (Namespace, rookoutErrors.RookoutError) {
	value := v.value
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Array:
		return NewValueNamespace(value.Index(key.(int))), nil

	case reflect.Map:
		k := value.MapKeys()

		for i := 0; i < len(k); i++ {
			if key.(string) == k[i].String() {
				return NewValueNamespace(value.MapIndex(k[i])), nil
			}
		}

	case reflect.Struct:
		return NewValueNamespace(value.FieldByName(key.(string))), nil

	case reflect.Slice:
		return NewValueNamespace(value.Index(key.(int))), nil
	}

	return nil, rookoutErrors.NewAgentKeyNotFoundException("", key, nil)
}

func (v *ValueNamespace) GetObject() (i interface{}) {
	defer func() {
		if recover() != nil {
			if v.value.CanAddr() {
				valueCopy := reflect.NewAt(v.value.Type(), unsafe.Pointer(v.value.UnsafeAddr()))
				valueCopy = valueCopy.Elem()
				i = valueCopy.Interface()
			}
		}
	}()

	return v.value.Interface()
}

func (v *ValueNamespace) Serialize(serializer Serializer) {
	dumpValue(serializer, v.value, v.ObjectDumpConf)
}

func (v *ValueNamespace) GetObjectDumpConfig() config.ObjectDumpConfig {
	return v.ObjectDumpConf
}

func (v *ValueNamespace) SetObjectDumpConfig(config config.ObjectDumpConfig) {
	v.ObjectDumpConf = config
}
