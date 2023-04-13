package namespaces

import (
	"fmt"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type GoUtilsNamespace struct {
	goid          int
	goroutineName string
}

func (g *GoUtilsNamespace) CallMethod(name string, args string) (Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "thread_id":
		return NewGoObjectNamespace(g.goid), nil

	case "thread_name":
		return NewGoObjectNamespace(g.goroutineName), nil
	default:
		
		return NewGoObjectNamespace(nil), nil
	}
}

func (g *GoUtilsNamespace) WriteAttribute(name string, value Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (g *GoUtilsNamespace) ReadAttribute(name string) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (g *GoUtilsNamespace) ReadKey(key interface{}) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (g *GoUtilsNamespace) GetObject() interface{} {
	return nil
}

func (g *GoUtilsNamespace) Serialize(serializer Serializer) {
	dumpError(serializer, rookoutErrors.NewNotImplemented())
}

func NewGoUtilsNameSpace(goid int) *GoUtilsNamespace {
	return &GoUtilsNamespace{goid: goid, goroutineName: fmt.Sprintf("goroutine %d", goid)}
}
