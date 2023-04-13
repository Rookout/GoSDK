package namespaces

import (
	"container/list"
	"strconv"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
)

const defaultTracebackDepth = 1000

type StackNamespace struct {
	collectionService *collection.CollectionService
}

func NewStackNamespace(collectionService *collection.CollectionService) *StackNamespace {
	return &StackNamespace{
		collectionService: collectionService,
	}
}

func (s *StackNamespace) ReadKey(_ interface{}) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (s *StackNamespace) CallMethod(name string, args string) (Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "traceback":
		return s.Traceback(args)
	case "frames":
		return nil, rookoutErrors.NewNotImplemented()
	}
	return nil, rookoutErrors.NewRookMethodNotFound(name)
}

func (s *StackNamespace) Traceback(args string) (Namespace, rookoutErrors.RookoutError) {
	depth := 0
	if len(args) > 0 {
		var err error
		depth, err = strconv.Atoi(args)
		if err != nil {
			return nil, rookoutErrors.NewRookInvalidMethodArguments("traceback()", args)
		}
	} else {
		depth = defaultTracebackDepth
	}

	l := list.New()
	for i, stackFrame := range s.collectionService.StackTraceElements {
		if i > depth {
			break
		}

		containerNamespace := NewEmptyContainerNamespace()
		_ = containerNamespace.WriteAttribute("filename", NewGoObjectNamespace(stackFrame.File))
		_ = containerNamespace.WriteAttribute("module", NewGoObjectNamespace(stackFrame.File))
		_ = containerNamespace.WriteAttribute("line", NewGoObjectNamespace(stackFrame.Line))
		_ = containerNamespace.WriteAttribute("function", NewGoObjectNamespace(stackFrame.Function))

		l.PushBack(containerNamespace)
	}

	return NewTracebackNamespace(s.collectionService.StackTraceElements, depth), nil
}

func (s *StackNamespace) ReadAttribute(_ string) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (s *StackNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (s *StackNamespace) GetObject() interface{} {
	return nil
}

func (s *StackNamespace) Serialize(serializer Serializer) {
	dumpError(serializer, rookoutErrors.NewNotImplemented())
}
