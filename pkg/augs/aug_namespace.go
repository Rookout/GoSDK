package augs

import (
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
)

type augNamespace struct {
	namespace namespaces.Namespace
}

func newAugNamespace(collectionService *collection.CollectionService) (*augNamespace, rookoutErrors.RookoutError) {
	frame := namespaces.NewFrameNamespace(collectionService)
	stack := namespaces.NewStackNamespace(collectionService)
	
	utils := namespaces.NewGoUtilsNameSpace(collectionService.GoroutineID())
	
	trace := namespaces.NewNoopNamespace()
	
	state := namespaces.NewEmptyContainerNamespace()
	
	extracted := namespaces.NewNoopNamespace()

	namespace := namespaces.NewEmptyContainerNamespace()
	namespace.OnClose = collectionService.Close
	err := namespace.WriteAttribute("frame", frame)
	if err != nil {
		return nil, err
	}
	err = namespace.WriteAttribute("stack", stack)
	if err != nil {
		return nil, err
	}
	err = namespace.WriteAttribute("store", namespaces.NewEmptyContainerNamespace())
	if err != nil {
		return nil, err
	}
	err = namespace.WriteAttribute("trace", trace)
	if err != nil {
		return nil, err
	}
	err = namespace.WriteAttribute("state", state)
	if err != nil {
		return nil, err
	}
	err = namespace.WriteAttribute("utils", utils)
	if err != nil {
		return nil, err
	}
	err = namespace.WriteAttribute("extracted", extracted)
	if err != nil {
		return nil, err
	}

	return &augNamespace{
		namespace: namespace,
	}, nil
}

func (a *augNamespace) GetAugNamespace() namespaces.Namespace {
	return a.namespace
}
