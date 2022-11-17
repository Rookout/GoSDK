package namespaces

import (
	"fmt"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type GoUtilsNamespace struct {
	goid          int
	goroutineName string
}

func (g *GoUtilsNamespace) CallMethod(name string, args string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "thread_id":
		return NewGoObjectNamespace(g.goid), nil

	case "thread_name":
		return NewGoObjectNamespace(g.goroutineName), nil
	default:
		
		return NewGoObjectNamespace(nil), nil
	}
}

func (g *GoUtilsNamespace) WriteAttribute(name string, value types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (g *GoUtilsNamespace) ReadAttribute(name string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (g *GoUtilsNamespace) ReadKey(key interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (g *GoUtilsNamespace) GetObject() interface{} {
	return nil
}

func (g *GoUtilsNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	return GetErrorVariant(rookoutErrors.NewNotImplemented(), logErrors)
}

func (g *GoUtilsNamespace) ToDict() map[string]interface{} {
	panic("not implemented")
}

func (g *GoUtilsNamespace) ToSimpleDict() interface{} {
	panic("not implemented")
}

func (g *GoUtilsNamespace) Filter(filters []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}

func NewGoUtilsNameSpace(goid int) *GoUtilsNamespace {
	return &GoUtilsNamespace{goid: goid, goroutineName: fmt.Sprintf("goroutine %d", goid)}
}
