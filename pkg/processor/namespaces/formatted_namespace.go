package namespaces

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type FormattedNamespace struct {
	message string
}

func NewFormattedNamespace(message string) types.Namespace {
	formattedNamespace := &FormattedNamespace{
		message: message,
	}

	return formattedNamespace
}

func (f FormattedNamespace) CallMethod(name string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewRookMethodNotFound(name)
}

func (f FormattedNamespace) ReadAttribute(_ string) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (f FormattedNamespace) WriteAttribute(_ string, _ types.Namespace) rookoutErrors.RookoutError {
	return rookoutErrors.NewNotImplemented()
}

func (f FormattedNamespace) ReadKey(_ interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	return nil, rookoutErrors.NewNotImplemented()
}

func (f FormattedNamespace) GetObject() interface{} {
	return f.message
}

func (f FormattedNamespace) ToProtobuf(logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	defer recoverFromPanic(recover(), v, logErrors)

	v.VariantType = pb.Variant_VARIANT_FORMATTED_MESSAGE
	v.Value = &pb.Variant_MessageValue{
		MessageValue: &pb.Variant_FormattedMessage{
			Message: f.message,
		},
	}

	return v
}

func (f FormattedNamespace) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"@namespace":   "FormattedNamespace",
		"@common_type": "Namespace",
		"@value":       f.message,
	}
}

func (f FormattedNamespace) ToSimpleDict() interface{} {
	return f.message
}

func (f FormattedNamespace) Filter(_ []types.FieldFilter) rookoutErrors.RookoutError {
	return nil
}
