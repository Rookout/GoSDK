package namespaces

import (
	"fmt"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

type ErrorNamespace struct {
	Message    string
	Parameters types.Namespace
	Exc        types.Namespace
	Traceback  types.Namespace
}

func NewErrorNamespace(
	message string,
	parameters types.Namespace,
	exc types.Namespace,
	traceback types.Namespace) *ErrorNamespace {
	er := &ErrorNamespace{
		Message:    message,
		Parameters: parameters,
		Exc:        exc,
		Traceback:  traceback,
	}

	return er
}

func (e ErrorNamespace) CallMethod(name string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewRookMethodNotFound(name)
}

func (e ErrorNamespace) ReadAttribute(name string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "Message":
		return NewGoObjectNamespace(e.Message), nil
	case "Exception":
		return e.Exc, nil
	case "Traceback":
		return e.Traceback, nil
	case "Parameters":
		return e.Parameters, nil
	}

	return nil, rookoutErrors.NewNotImplemented()
}

func (e ErrorNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (e ErrorNamespace) ReadKey(_ interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (e ErrorNamespace) GetObject() interface{} {
	return nil
}

func (e ErrorNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer func() {
		if recoverMessage := recover(); recoverMessage != nil {
			var ok bool
			err, ok := recoverMessage.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", recoverMessage)
			}

			nilVariant := &pb.Variant{}
			nilVariant.VariantType = pb.Variant_VARIANT_NONE

			v.Reset()
			v.VariantType = pb.Variant_VARIANT_ERROR
			v.Value = &pb.Variant_ErrorValue{
				ErrorValue: &pb.Error{
					Message:   err.Error(),
					Type:      "error",
					Exc:       nilVariant,
					Traceback: NewGoObjectNamespace(string(debug.Stack())).ToProtobuf(logErrors),
				},
			}

			if logErrors {
				logrus.Error(err.Error())
			}
		}
	}()

	var parametersPB *pb.Variant
	if e.Parameters != nil {
		parametersPB = e.Parameters.ToProtobuf(logErrors)
	}

	v.VariantType = pb.Variant_VARIANT_ERROR
	v.Value = &pb.Variant_ErrorValue{
		ErrorValue: &pb.Error{
			Message:    e.Message,
			Type:       "error",
			Parameters: parametersPB,
			Exc:        e.Exc.ToProtobuf(logErrors),
			Traceback:  e.Traceback.ToProtobuf(logErrors),
		},
	}

	return v
}

func (e ErrorNamespace) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"@namespace":   "ErrorNamespace",
		"@common_type": "Namespace",
		"@value": map[string]interface{}{
			"message":    e.Message,
			"parameters": e.Parameters.ToDict(),
			"exc":        e.Exc.ToDict(),
		},
	}
}

func (e ErrorNamespace) ToSimpleDict() interface{} {
	return map[string]interface{}{
		"message":    e.Message,
		"parameters": e.Parameters.ToSimpleDict(),
		"exc":        e.Exc.ToSimpleDict(),
	}
}

func (e ErrorNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}
