package namespaces

import (
	"io"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type ContainerNamespace struct {
	Obj     map[string]Namespace
	OnClose func() error
}

func NewEmptyContainerNamespace() *ContainerNamespace {
	c := &ContainerNamespace{
		Obj: make(map[string]Namespace),
	}

	return c
}

func NewContainerNamespace(initObj *map[string]Namespace) *ContainerNamespace {
	if nil == initObj {
		initObj = &map[string]Namespace{}
	}

	c := &ContainerNamespace{
		Obj: *initObj,
	}

	return c
}

func (c *ContainerNamespace) CallMethod(name string, _ string) (Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "size":
		return NewGoObjectNamespace(len(c.Obj)), nil

	default:
		return nil, rookoutErrors.NewRookMethodNotFound(name)
	}
}

func (c *ContainerNamespace) ReadAttribute(name string) (Namespace, rookoutErrors.RookoutError) {
	if val, ok := c.Obj[name]; ok {
		return val, nil
	}
	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (c *ContainerNamespace) WriteAttribute(name string, value Namespace) rookoutErrors.RookoutError {
	c.Obj[name] = value
	return nil
}

func (c *ContainerNamespace) ReadKey(_ interface{}) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (c *ContainerNamespace) GetObject() interface{} {
	return &c.Obj
}

func (c *ContainerNamespace) Serialize(serializer Serializer) {
	names := make([]string, 0, len(c.Obj))
	for k := range c.Obj {
		names = append(names, k)
	}

	getNamedValue := func(i int) (string, Namespace) {
		return names[i], c.Obj[names[i]]
	}
	serializer.dumpNamespace(getNamedValue, len(c.Obj))
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
