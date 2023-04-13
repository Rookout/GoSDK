package namespaces

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type NoopNamespace struct{}

func NewNoopNamespace() *NoopNamespace {
	return &NoopNamespace{}
}

func (n *NoopNamespace) CallMethod(_ string, _ string) (Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(nil), nil
}

func (n *NoopNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (n *NoopNamespace) ReadAttribute(_ string) (Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(nil), nil
}

func (n *NoopNamespace) ReadKey(_ interface{}) (Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(nil), nil
}

func (n *NoopNamespace) GetObject() interface{} {
	return nil
}

func (n *NoopNamespace) Serialize(serializer Serializer) {
	dumpError(serializer, rookoutErrors.NewNotImplemented())
}
