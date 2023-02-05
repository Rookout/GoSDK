package actions

import (
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type Processor interface {
	Process(namespace types.Namespace)
}

type ProcessorFactory interface {
	GetProcessor(configuration []types.AugConfiguration) (Processor, rookoutErrors.RookoutError)
}

type Action interface {
	Execute(augId types.AugId, reportId string, namespace types.Namespace, output com_ws.Output) rookoutErrors.RookoutError
}

type actionRunProcessor struct {
	processor     Processor
	postProcessor Processor
}

func NewActionRunProcessor(arguments types.AugConfiguration, processorFactory ProcessorFactory) (Action, rookoutErrors.RookoutError) {
	rawOps, ok := arguments["operations"].([]interface{})
	if !ok {
		return nil, rookoutErrors.NewRookInvalidActionConfiguration(arguments)
	}
	var ops []types.AugConfiguration
	for _, rawOp := range rawOps {
		ops = append(ops, rawOp.(map[string]interface{}))
	}
	p, err := processorFactory.GetProcessor(ops)
	if err != nil {
		return nil, err
	}

	var postProcessor Processor
	if rawPostOps, ok := arguments["post_operations"].([]interface{}); ok {
		var postOps []types.AugConfiguration
		for _, rawPostOp := range rawPostOps {
			postOps = append(postOps, rawPostOp.(map[string]interface{}))
		}
		postProcessor, _ = processorFactory.GetProcessor(postOps)
	}

	return &actionRunProcessor{
		processor:     p,
		postProcessor: postProcessor,
	}, nil
}

func (a *actionRunProcessor) Execute(augId types.AugId, reportId string, namespace types.Namespace, output com_ws.Output) (err rookoutErrors.RookoutError) {
	a.processor.Process(namespace)
	attribute, err := namespace.ReadAttribute("store")
	if err != nil {
		return err
	}

	if _, ok := attribute.(*namespaces.ContainerNamespace); ok {
		attribute.(*namespaces.ContainerNamespace).OnClose = namespace.(*namespaces.ContainerNamespace).OnClose
	}
	output.SendUserMessage(augId, reportId, attribute)

	if a.postProcessor != nil {
		a.postProcessor.Process(namespace)
	}

	return nil
}
