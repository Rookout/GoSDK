package namespaces

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type NoopNamespace struct{}

func NewNoopNamespace() *NoopNamespace {
	return &NoopNamespace{}
}

func (n *NoopNamespace) CallMethod(_ string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(nil), nil
}

func (n *NoopNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (n *NoopNamespace) ReadAttribute(_ string) (types.Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(nil), nil
}

func (n *NoopNamespace) ReadKey(_ interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(nil), nil
}

func (n *NoopNamespace) GetObject() interface{} {
	return nil
}

func (n *NoopNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	return GetErrorVariant(rookoutErrors.NewNotImplemented(), logErrors)
}

func (n *NoopNamespace) ToDict() map[string]interface{} {
	panic("not implemented")
}

func (n *NoopNamespace) ToSimpleDict() interface{} {
	panic("not implemented")
}

func (n *NoopNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}
