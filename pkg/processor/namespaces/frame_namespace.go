package namespaces

import (
	"strconv"
	"strings"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
)

type FrameNamespace struct {
	collectionService *collection.CollectionService
	locals            map[string]*VariableNamespace
}

func NewFrameNamespace(collectionService *collection.CollectionService) *FrameNamespace {
	return &FrameNamespace{
		collectionService: collectionService,
		locals:            make(map[string]*VariableNamespace),
	}
}

func (f *FrameNamespace) CallMethod(name string, args string) (Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "filename":
		return NewGoObjectNamespace(f.collectionService.GetFrame().File), nil
	case "line":
		return NewGoObjectNamespace(f.collectionService.GetFrame().Line), nil
	case "method", "function":
		return NewGoObjectNamespace(f.collectionService.GetFrame().Function), nil
	case "locals":
		return f.GetLocals(args)
	case "dump":
		return f.GetDump(args)
	default:
		return nil, rookoutErrors.NewNotImplemented()
	}
}

func (f *FrameNamespace) getAllVariables(config config.ObjectDumpConfig) ([]*variable.Variable, rookoutErrors.RookoutError) {
	return f.collectionService.GetVariables(config), nil
}

func (f *FrameNamespace) variablesToLocals(vars []*variable.Variable, config config.ObjectDumpConfig) {
	for _, v := range vars {
		if _, ok := f.locals[v.Name]; !ok {
			obj := NewVariableNamespace(v.Name, v, f.collectionService)
			obj.ObjectDumpConf = config
			f.locals[v.Name] = obj
		}
	}
}

func (f *FrameNamespace) GetLocals(args string) (Namespace, rookoutErrors.RookoutError) {
	maxDepth := 0
	dumpConfig := config.GetDefaultDumpConfig()

	if len(args) > 0 {
		var err error
		if maxDepth, err = strconv.Atoi(args); err == nil {
			dumpConfig.MaxDepth = maxDepth
		} else {
			var ok bool
			dumpConfig, ok = config.GetObjectDumpConfig(strings.ToLower(args))
			if !ok {
				return nil, rookoutErrors.NewRookInvalidMethodArguments("locals()", args)
			}
		}
	}
	vars, err := f.getAllVariables(dumpConfig)
	if err != nil {
		return nil, err
	}

	f.variablesToLocals(vars, dumpConfig)
	locals := make(map[string]Namespace, len(f.locals))
	for name, local := range f.locals {
		locals[name] = local
	}

	return NewContainerNamespace(&locals), nil
}

func (f *FrameNamespace) GetDump(args string) (*ContainerNamespace, rookoutErrors.RookoutError) {
	c := NewEmptyContainerNamespace()

	locals, err := f.GetLocals(args)
	if err != nil {
		return nil, err
	}
	_ = c.WriteAttribute("locals", locals)
	_ = c.WriteAttribute("filename", NewGoObjectNamespace(f.collectionService.GetFrame().File))
	_ = c.WriteAttribute("module", NewGoObjectNamespace(f.collectionService.GetFrame().File))
	_ = c.WriteAttribute("line", NewGoObjectNamespace(f.collectionService.GetFrame().Line))
	_ = c.WriteAttribute("function", NewGoObjectNamespace(f.collectionService.GetFrame().Function))

	return c, nil
}

func (f *FrameNamespace) ReadAttribute(name string) (Namespace, rookoutErrors.RookoutError) {
	
	if local, ok := f.locals[name]; ok {
		if local.ObjectDumpConf.IsTailored {
			return local, nil
		}
	}

	dumpConfig := config.GetDefaultDumpConfig()
	dumpConfig.ShouldTailor = true

	
	v, err := f.collectionService.GetVariable(name, dumpConfig)
	if err != nil {
		return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
	}
	obj := NewVariableNamespace(name, v, f.collectionService)
	f.locals[name] = obj
	return obj, nil
}

func (f *FrameNamespace) WriteAttribute(_ string, _ Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (f *FrameNamespace) ReadKey(_ interface{}) (Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (f *FrameNamespace) GetObject() interface{} {
	return nil
}

func (f *FrameNamespace) Serialize(serializer Serializer) {
	dump, _ := f.GetDump("")
	dump.Serialize(serializer)
}
