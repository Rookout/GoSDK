package paths

import (
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"strings"
)

type pathOperation interface {
	Read(namespace namespaces.Namespace, create bool) (namespaces.Namespace, rookoutErrors.RookoutError)
}

type writeOperation interface {
	pathOperation
	Write(namespace namespaces.Namespace, value namespaces.Namespace) rookoutErrors.RookoutError
}

type lookupOperation struct {
	name interface{}
}

type methodOperation struct {
	methodName      string
	methodArguments string
}

type attributeOperation struct {
	name string
}

func newLookupOperation(name string) (pathOperation, rookoutErrors.RookoutError) {
	a := &lookupOperation{}

	switch {
	case strings.HasPrefix(name, "'"):
		a.name = strings.Trim(name, "'")
	case strings.HasPrefix(name, "\""):
		a.name = strings.Trim(name, "\"")
	default:
		d, err := utils.StringToInt(name)
		if err != nil {
			return nil, rookoutErrors.NewBadTypeException("type is not a string", err)
		}
		a.name = d
	}

	return a, nil
}

func (l lookupOperation) Read(namespace namespaces.Namespace, _ bool) (namespaces.Namespace, rookoutErrors.RookoutError) {
	return namespace.ReadKey(l.name)
}

func newMethodOperation(methodName string, methodArguments string) pathOperation {
	f := &methodOperation{
		methodName:      methodName,
		methodArguments: methodArguments,
	}

	return f
}

func (f methodOperation) Read(namespace namespaces.Namespace, _ bool) (namespaces.Namespace, rookoutErrors.RookoutError) {
	return namespace.CallMethod(f.methodName, f.methodArguments)
}

func newAttributeOperation(name string) pathOperation {
	a := &attributeOperation{
		name: name,
	}

	return a
}

func (a attributeOperation) Read(namespace namespaces.Namespace, create bool) (namespaces.Namespace, rookoutErrors.RookoutError) {
	if n, err := namespace.ReadAttribute(a.name); err == nil {
		return n, nil
	}

	if create {
		err := namespace.WriteAttribute(a.name, namespaces.NewEmptyContainerNamespace())
		if err != nil {
			return nil, err
		}
		return namespace.ReadAttribute(a.name)
	}

	return nil, rookoutErrors.NewRookAttributeNotFoundException(a.name)
}

func (a attributeOperation) Write(namespace namespaces.Namespace, value namespaces.Namespace) rookoutErrors.RookoutError {
	return namespace.WriteAttribute(a.name, value)
}
