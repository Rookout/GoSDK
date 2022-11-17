package namespaces

import (
	"container/list"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/types"
	"strconv"
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

func (s *StackNamespace) ReadKey(_ interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (s *StackNamespace) CallMethod(name string, args string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "traceback":
		return s.Traceback(args)
	case "frames":
		return nil, rookoutErrors.NewNotImplemented()
	}
	return nil, rookoutErrors.NewRookMethodNotFound(name)
}

func (s *StackNamespace) Traceback(args string) (types.Namespace, rookoutErrors.RookoutError) {
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
		_ = containerNamespace.WriteAttribute("function", NewGoObjectNamespace(stackFrame.Function.Name))

		l.PushBack(containerNamespace)
	}

	return NewListNamespace(l, int32(depth), map[string]types.Namespace{}), nil
}

func (s *StackNamespace) ReadAttribute(_ string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (s *StackNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (s *StackNamespace) GetObject() interface{} {
	return nil
}

func (s *StackNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	return GetErrorVariant(rookoutErrors.NewNotImplemented(), logErrors)
}

func (s *StackNamespace) ToDict() map[string]interface{} {
	panic("not implemented")
}

func (s *StackNamespace) ToSimpleDict() interface{} {
	panic("not implemented")
}

func (s StackNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}
