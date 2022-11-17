package namespaces

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"io"
)

type ContainerNamespace struct {
	Obj     map[string]types.Namespace
	OnClose func() error
}

func NewEmptyContainerNamespace() *ContainerNamespace {
	c := &ContainerNamespace{
		Obj: make(map[string]types.Namespace),
	}

	return c
}

func NewContainerNamespace(initObj *map[string]types.Namespace) *ContainerNamespace {
	if nil == initObj {
		initObj = &map[string]types.Namespace{}
	}

	c := &ContainerNamespace{
		Obj: *initObj,
	}

	return c
}

func (c ContainerNamespace) CallMethod(name string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "size":
		return NewGoObjectNamespace(len(c.Obj)), nil

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (c ContainerNamespace) ReadAttribute(name string) (types.Namespace, rookoutErrors.RookoutError) {
	if val, ok := c.Obj[name]; ok {
		return val, nil
	}
	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (c ContainerNamespace) WriteAttribute(name string, value types.Namespace) rookoutErrors.RookoutError {
	c.Obj[name] = value
	return nil
}

func (c ContainerNamespace) ReadKey(_ interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (c ContainerNamespace) GetObject() interface{} {
	return &c.Obj
}

func (c ContainerNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer recoverFromPanic(recover(), v, logErrors)

	v.VariantType = pb.Variant_VARIANT_NAMESPACE

	attributes := make([]*pb.Variant_NamedValue, 0)
	for k, val := range c.Obj {

		dumpedValue := val.ToProtobuf(logErrors)

		attributes = append(attributes, &pb.Variant_NamedValue{
			Name:  k,
			Value: dumpedValue,
		})
	}

	v.Value = &pb.Variant_NamespaceValue{
		NamespaceValue: &pb.Variant_Namespace{
			Attributes: attributes}}

	return v
}

func (c ContainerNamespace) ToDict() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range c.Obj {
		result[k] = v.ToDict()
	}

	return result
}

func (c ContainerNamespace) ToSimpleDict() interface{} {
	result := make(map[string]interface{})

	for k, v := range c.Obj {
		result[k] = v.ToSimpleDict()
	}

	return result
}

func (c ContainerNamespace) Filter(filters []types.FieldFilter) rookoutErrors.RookoutError {
	for _, filter := range filters {
		for name, value := range c.Obj {
			if filter.FilterType == FilterFieldName {
				if filter.Pattern.Match([]byte(name)) {
					c.Obj[name] = NewGoObjectNamespace(FilteredFieldReplacement)
					continue
				}
			}

			err := value.Filter(filters)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c ContainerNamespace) Close() error {
	var err error
	if c.OnClose != nil {
		err = c.OnClose()
	}
	for _, v := range c.Obj {
		if closer, ok := v.(io.Closer); ok {
			_ = closer.Close()
		}
	}
	return err
}
