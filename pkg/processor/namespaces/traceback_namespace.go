package namespaces

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
)

type TracebackNamespace struct {
	traceback []collection.Stackframe
	depth     int
}

func NewTracebackNamespace(traceback []collection.Stackframe, depth int) *TracebackNamespace {
	return &TracebackNamespace{
		traceback: traceback,
		depth:     depth,
	}
}

func (t *TracebackNamespace) ReadKey(key interface{}) (Namespace, rookoutErrors.RookoutError) {
	collectionService, err := collection.NewCollectionService(registers.OnStackRegisters{}, 0, []collection.Stackframe{t.traceback[key.(int)]}, nil, 0)
	if err != nil {
		return nil, err
	}
	return NewFrameNamespace(collectionService), nil
}

func (t *TracebackNamespace) CallMethod(name string, _ string) (Namespace, rookoutErrors.RookoutError) {
	if name == "size" {
		return NewGoObjectNamespace(t.depth), nil
	}

	return nil, rookoutErrors.NewRookMethodNotFound(name)
}

func (t *TracebackNamespace) ReadAttribute(_ string) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (t *TracebackNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (t *TracebackNamespace) GetObject() interface{} {
	return nil
}

func (t *TracebackNamespace) Serialize(serializer Serializer) {
	getFrame := func(i int) (int, string, string) {
		return t.traceback[i].Line, t.traceback[i].File, t.traceback[i].Function
	}
	tracebackLen := t.depth
	if tracebackLen > len(t.traceback) {
		tracebackLen = len(t.traceback)
	}
	serializer.dumpTraceback(getFrame, tracebackLen)
}

func (t *TracebackNamespace) GetTraceback() []collection.Stackframe {
	return t.traceback
}
