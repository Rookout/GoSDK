package namespaces

import (
	"encoding/json"
	"fmt"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/sirupsen/logrus"
	"math"
	"reflect"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

const intMin = -2147483648
const intMax = 2147483647
const longMin = -(1 << 63)
const longMax = (1 << 63) - 1

const (
	FilterFieldName  types.FieldFilterType = iota
	FilterFieldValue                       = iota
)

const FilteredValueReplacement = "****"
const FilteredFieldReplacement = "[REDACTED]"

type ProtobufNamespace struct {
	variant *pb.Variant
}

func NewProtobufNamespace(v *pb.Variant) (p types.Namespace) {
	if v == nil {
		p = &ProtobufNamespace{
			variant: &pb.Variant{VariantType: pb.Variant_VARIANT_NONE},
		}

	} else {
		p = &ProtobufNamespace{
			variant: v,
		}
	}

	return p
}

func (p ProtobufNamespace) loadAttributes(v *pb.Variant) map[string]types.Namespace {
	m := make(map[string]types.Namespace)

	for i := range v.Attributes {
		m[v.Attributes[i].Name] = NewProtobufNamespace(v.Attributes[i].Value)
	}

	return m
}

func (p ProtobufNamespace) GetName() (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_CODE_OBJECT:
		return NewGoObjectNamespace(p.variant.GetCodeValue().Name), nil
	case pb.Variant_VARIANT_STRING:
		return NewGoObjectNamespace(p.variant.GetStringValue().Value), nil
	case pb.Variant_VARIANT_OBJECT:
		return NewGoObjectNamespace(p.variant.OriginalType), nil

	}
	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) GetFileName() (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_CODE_OBJECT:
		return NewGoObjectNamespace(p.variant.GetCodeValue().Filename), nil
	}
	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) GetLineNumber() (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_CODE_OBJECT:
		return NewGoObjectNamespace(p.variant.GetCodeValue().Lineno), nil
	}
	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) GetModuleName() (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_CODE_OBJECT:
		return NewGoObjectNamespace(p.variant.GetCodeValue().Module), nil
	case pb.Variant_VARIANT_STRING:
		return NewGoObjectNamespace(p.variant.GetStringValue().Value), nil
	}
	return nil, rookoutErrors.NewNotImplemented()
}
func (p ProtobufNamespace) GetMaxDepth() (types.Namespace, rookoutErrors.RookoutError) {
	return NewGoObjectNamespace(p.variant.MaxDepth), nil
}
func (p ProtobufNamespace) GetCommonType() (types.Namespace, rookoutErrors.RookoutError) {
	commonType := ""
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_NONE:
		commonType = "nil"
	case pb.Variant_VARIANT_INT:
		if "bool" == p.variant.OriginalType {
			commonType = "bool"
		} else {
			commonType = "int"
		}
	case pb.Variant_VARIANT_LONG:
		commonType = "long"
	case pb.Variant_VARIANT_DOUBLE:
		if "float32" == p.variant.OriginalType {
			commonType = "float32"
		} else {
			commonType = "float64"
		}
	case pb.Variant_VARIANT_COMPLEX:
		if "complex64" == p.variant.OriginalType {
			commonType = "complex64"
		} else {
			commonType = "complex128"
		}
	case pb.Variant_VARIANT_STRING:
		commonType = "string"
	case pb.Variant_VARIANT_BINARY:
		commonType = "binary"
	case pb.Variant_VARIANT_TIME:
		commonType = "Time"
	case pb.Variant_VARIANT_LIST:
		commonType = "list"
	case pb.Variant_VARIANT_MAP:
		commonType = "Map"
	case pb.Variant_VARIANT_ENUM:
		commonType = "Enum"
	}

	if "" != commonType {
		return NewGoObjectNamespace(commonType), nil
	}
	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) GetOriginalSizeOfVariant() (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_STRING:
		return NewGoObjectNamespace(p.variant.GetStringValue().OriginalSize), nil
	case pb.Variant_VARIANT_BINARY:
		return NewGoObjectNamespace(p.variant.GetBinaryValue().OriginalSize), nil
	case pb.Variant_VARIANT_LIST:
		return NewGoObjectNamespace(p.variant.GetListValue().OriginalSize), nil
	case pb.Variant_VARIANT_MAP:
		return NewGoObjectNamespace(p.variant.GetMapValue().OriginalSize), nil
	}

	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) GetSizeOfVariant() (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_STRING:
		return NewGoObjectNamespace(len(p.variant.GetStringValue().Value)), nil
	case pb.Variant_VARIANT_NAMESPACE:
		ns := p.variant.GetNamespaceValue()

		if ns != nil {
			return NewGoObjectNamespace(len(ns.Attributes)), nil
		}

		return nil, rookoutErrors.NewNilProtobufNamespaceException()

	case pb.Variant_VARIANT_LIST:
		ns := p.variant.GetListValue()

		if ns != nil {
			return NewGoObjectNamespace(len(ns.Values)), nil
		}

		return nil, rookoutErrors.NewNilProtobufNamespaceException()

	case pb.Variant_VARIANT_MAP:
		ns := p.variant.GetMapValue()

		if ns != nil {
			return NewGoObjectNamespace(len(ns.Pairs)), nil
		}

		return nil, rookoutErrors.NewNilProtobufNamespaceException()

	case pb.Variant_VARIANT_BINARY:
		return NewGoObjectNamespace(len(p.variant.GetBinaryValue().Value)), nil
	}

	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) CallMethod(name string, _ string) (types.Namespace, rookoutErrors.RookoutError) {
	switch name {
	case "size":
		return p.GetSizeOfVariant()

	case "original_size":
		return p.GetOriginalSizeOfVariant()

	case "type":
		return NewGoObjectNamespace(p.variant.OriginalType), nil

	case "common_type":
		return p.GetCommonType()

	case "max_depth":
		return p.GetMaxDepth()

	case "name":
		return p.GetName()

	case "filename":
		return p.GetFileName()

	case "lineno":
		return p.GetLineNumber()

	case "module":
		return p.GetModuleName()

	case "toJson":
		if p.variant.VariantType == pb.Variant_VARIANT_NAMESPACE {
			result := p.ToSimpleDict()
			jsonString, err := json.Marshal(result)
			if err != nil {
				return nil, rookoutErrors.NewJsonMarshallingException(result, err)
			}
			return NewGoObjectNamespace(jsonString), nil
		}

		return nil, rookoutErrors.NewNotImplemented()

	case "format_stack":
		if result := p.tracebackToSimpleDict(); result != nil {
			return NewGoObjectNamespace(p.tracebackToSimpleDict()), nil
		}

		return nil, rookoutErrors.NewBadVariantType("Variant is not a traceback", p.variant)

	}
	return nil, rookoutErrors.NewRookMethodNotFound(name)
}


func (p ProtobufNamespace) ReadAttribute(name string) (types.Namespace, rookoutErrors.RookoutError) {
	if nil == p.variant {
		return nil, rookoutErrors.NewNilProtobufNamespaceException()
	}

	switch p.variant.VariantType {
	case pb.Variant_VARIANT_NAMESPACE:
		ns := p.variant.GetNamespaceValue()

		if ns == nil {
			return nil, rookoutErrors.NewNilProtobufNamespaceException()
		}

		for _, attribute := range ns.Attributes {
			if attribute.Name == name {
				return NewProtobufNamespace(attribute.Value), nil
			}
		}
	case pb.Variant_VARIANT_ERROR:
		errorValue := p.variant.GetErrorValue()

		switch name {
		case "Message":
			return NewGoObjectNamespace(errorValue.Message), nil
		case "Exception":
			return NewProtobufNamespace(errorValue.Exc), nil
		case "Traceback":
			return NewProtobufNamespace(errorValue.Traceback), nil
		case "Parameters":
			return NewProtobufNamespace(errorValue.Parameters), nil
		}
	}

	for _, attrib := range p.variant.Attributes {
		if attrib.Name == name {
			return NewProtobufNamespace(attrib.Value), nil
		}
	}
	return nil, rookoutErrors.NewRookAttributeNotFoundException(name)
}

func (p ProtobufNamespace) WriteAttribute(name string, value types.Namespace) rookoutErrors.RookoutError {
	if p.variant.VariantType != pb.Variant_VARIANT_NAMESPACE {
		return rookoutErrors.NewRookOperationReadOnlyException("ProtobufWriteOperation")
	}

	namespace := p.variant.GetNamespaceValue()
	attrs := make([]*pb.Variant_NamedValue, 0)
	for _, attr := range namespace.Attributes {
		if attr.Name == name {
			continue
		}
		attrs = append(attrs, attr)
	}
	namespace.Attributes = append(attrs, &pb.Variant_NamedValue{Name: name, Value: value.ToProtobuf(true)})

	return nil
}

func (p ProtobufNamespace) ReadKey(key interface{}) (types.Namespace, rookoutErrors.RookoutError) {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_LIST:
		if keyAsInt, ok := key.(int); true == ok {
			ns := p.variant.GetListValue()

			if ns == nil {
				return nil, rookoutErrors.NewNilProtobufNamespaceException()
			}

			if (keyAsInt >= 0) && (len(ns.Values) > keyAsInt) {
				return NewProtobufNamespace(ns.Values[keyAsInt]), nil
			}
		}

	case pb.Variant_VARIANT_MAP:
		for _, pair := range p.variant.GetMapValue().Pairs {
			if reflect.DeepEqual(NewProtobufNamespace(pair.First).ToSimpleDict(), key) {
				return NewProtobufNamespace(pair.Second), nil
			}
		}
		return nil, rookoutErrors.NewAgentKeyNotFoundException("", key, nil)
	}

	return nil, rookoutErrors.NewNotImplemented()
}

func (p ProtobufNamespace) namespaceToSimpleDict(v *pb.Variant) map[string]interface{} {
	m := make(map[string]interface{})
	ns := v.GetNamespaceValue()

	if ns == nil {
		return m
	}

	for _, attribute := range ns.Attributes {
		m[attribute.Name] = NewProtobufNamespace(attribute.Value).ToSimpleDict()
	}

	return m
}

func (p ProtobufNamespace) namespaceToDict(v *pb.Variant) map[string]interface{} {
	m := make(map[string]interface{})

	if namespaceValue := v.GetNamespaceValue(); namespaceValue != nil {
		for _, attribute := range namespaceValue.Attributes {
			m[attribute.Name] = NewProtobufNamespace(attribute.Value).ToDict()
		}
	}

	return m
}

func (p ProtobufNamespace) mapToSimpleDict(v *pb.Variant) *map[string]interface{} {
	m := make(map[string]interface{})
	ns := v.GetMapValue()

	if ns == nil {
		return &m
	}

	for _, pair := range ns.Pairs {
		keyAsSimpleDict := NewProtobufNamespace(pair.First).ToSimpleDict()
		if keyAsString, ok := keyAsSimpleDict.(string); ok {
			m[keyAsString] = NewProtobufNamespace(pair.Second).ToSimpleDict()
		} else {
			m[fmt.Sprintf("%v", keyAsSimpleDict)] = NewProtobufNamespace(pair.Second).ToSimpleDict()
		}
	}

	return &m
}

func (p ProtobufNamespace) mapToDict(v *pb.Variant) map[string]interface{} {
	ns := v.GetMapValue()
	if ns == nil {
		return make(map[string]interface{})
	}
	items := make([][]interface{}, len(ns.Pairs))

	for i, enumeratedVariant := range ns.Pairs {
		key := NewProtobufNamespace(enumeratedVariant.First).ToDict()
		value := NewProtobufNamespace(enumeratedVariant.Second).ToDict()
		keyValueTuple := []interface{}{key, value}
		items[i] = keyValueTuple
	}
	utils.MapStringToMapInterface(p.mapToSimpleDict(v))
	return map[string]interface{}{
		"@namespace":     "DictNamespace",
		"@common_type":   dictCommonType,
		"@original_type": v.OriginalType,
		"@max_depth":     v.MaxDepth,
		"@original_size": ns.OriginalSize,
		"@attributes":    getAttributesDict(p.loadAttributes(v)),
		"@value":         items,
	}
}

func (p ProtobufNamespace) listToSimpleDict(v *pb.Variant) interface{} {
	ns := v.GetListValue()
	retVal := make([]interface{}, len(ns.Values))

	if ns == nil {
		return make(map[string]interface{})
	}

	for i, enumeratedVariant := range ns.Values {
		retVal[i] = NewProtobufNamespace(enumeratedVariant).ToSimpleDict()
	}

	return retVal
}

func (p ProtobufNamespace) listToDict(v *pb.Variant) map[string]interface{} {
	ns := v.GetListValue()

	if ns == nil {
		return make(map[string]interface{})
	}
	l := make([]interface{}, len(ns.Values))
	for i, enumeratedVariant := range ns.Values {
		l[i] = NewProtobufNamespace(enumeratedVariant).ToDict()
	}

	commonType := v.GetListValue().GetType()
	return map[string]interface{}{
		"@namespace":     "ListNamespace",
		"@common_type":   commonType,
		"@original_type": v.OriginalType,
		"@max_depth":     v.MaxDepth,
		"@original_size": ns.OriginalSize,
		"@attributes":    getAttributesDict(p.loadAttributes(v)),
		"@value":         l,
	}
}

func (p ProtobufNamespace) GetObject() interface{} {
	return p.ToSimpleDict()
}

func (p ProtobufNamespace) ToProtobuf(_ bool) *pb.Variant {
	return p.variant
}

func (p ProtobufNamespace) primitiveTypeToDict(
	value interface{},
	commonType string,
	variant *pb.Variant) map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "DumpedPrimitiveNamespace",
		"@common_type":   commonType,
		"@original_type": variant.OriginalType,
		"@max_depth":     variant.MaxDepth,
		"@attributes":    getAttributesDict(p.loadAttributes(variant)),
		"@value":         value,
	}
}

func (p ProtobufNamespace) bufferTypeToDict(
	value interface{},
	originalSize int,
	commonType string,
	variant *pb.Variant) map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "StringNamespace",
		"@common_type":   commonType,
		"@original_size": originalSize,
		"@original_type": p.variant.OriginalType,
		"@max_depth":     p.variant.MaxDepth,
		"@attributes":    getAttributesDict(p.loadAttributes(variant)),
		"@value":         value,
	}
}

func (p ProtobufNamespace) EnumTypeToDict(value, typeName string, ordinalValue int32, commonType string, variant *pb.Variant) map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "EnumNamespace",
		"@common_type":   commonType,
		"@original_type": variant.OriginalType,
		"@max_depth":     variant.MaxDepth,
		"@attributes":    getAttributesDict(p.loadAttributes(variant)),
		"@value": map[string]interface{}{
			"@ordinal_value": ordinalValue,
			"@type_name":     typeName,
			"@value":         value,
		},
	}
}

func (p ProtobufNamespace) ToDict() map[string]interface{} {
	var finalValue interface{}
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_NONE:
		return p.primitiveTypeToDict(nil, "null", p.variant)
	case pb.Variant_VARIANT_INT:
		if "bool" == p.variant.OriginalType {
			finalValue = utils.IntAsBool(p.variant.GetIntValue())
		} else {
			finalValue = p.variant.GetIntValue()
		}
		return p.primitiveTypeToDict(finalValue, "int", p.variant)
	case pb.Variant_VARIANT_LONG:
		return p.primitiveTypeToDict(int64ToSafeJsNumber(p.variant.GetLongValue()), "int", p.variant)
	case pb.Variant_VARIANT_DOUBLE:
		if "float32" == p.variant.OriginalType {
			finalValue = float32(p.variant.GetDoubleValue())
			if math.IsNaN(float64(finalValue.(float32))) {
				finalValue = "NaN"
			}
		} else {
			finalValue = p.variant.GetDoubleValue()
			if math.IsNaN(finalValue.(float64)) {
				finalValue = "NaN"
			} else if math.IsInf(finalValue.(float64), 1) {
				finalValue = "+Infinity"
			} else if math.IsInf(finalValue.(float64), -1) {
				finalValue = "-Infinity"
			}
		}
		return p.primitiveTypeToDict(finalValue, "float", p.variant)
	case pb.Variant_VARIANT_STRING:
		return p.bufferTypeToDict(p.variant.GetStringValue().Value,
			int(p.variant.GetStringValue().GetOriginalSize()),
			"string",
			p.variant)
	case pb.Variant_VARIANT_BINARY:
		return p.bufferTypeToDict(getBinaryValueWithDefault(p.variant.GetBinaryValue().Value),
			int(p.variant.GetBinaryValue().GetOriginalSize()),
			"binary",
			p.variant)
	case pb.Variant_VARIANT_TIME:
		return p.primitiveTypeToDict(utils.ProtoToBackendCompatibleIsoTimeFormat(p.variant.GetTimeValue()),
			"datetime", p.variant)
	case pb.Variant_VARIANT_LARGE_INT:
		return p.primitiveTypeToDict(p.variant.GetLargeIntValue().Value, "int", p.variant)
	case pb.Variant_VARIANT_COMPLEX:
		if "complex64" == p.variant.OriginalType {
			finalValue = complex(float32(p.variant.GetComplexValue().Real), float32(p.variant.GetComplexValue().Imaginary))
		} else {
			finalValue = complex(p.variant.GetComplexValue().Real, p.variant.GetComplexValue().Imaginary)
		}
		return p.primitiveTypeToDict(finalValue, "complex", p.variant)
	case pb.Variant_VARIANT_UNDEFINED:
		return p.primitiveTypeToDict(nil, "null", p.variant)
	case pb.Variant_VARIANT_ENUM:
		ev := p.variant.GetEnumValue()
		return p.EnumTypeToDict(ev.GetStringValue(), ev.GetTypeName(), ev.GetOrdinalValue(), "Enum", p.variant)

	

	case pb.Variant_VARIANT_LIST:
		return p.listToDict(p.variant)
	case pb.Variant_VARIANT_MAP:
		return p.mapToDict(p.variant)
	case pb.Variant_VARIANT_OBJECT:
		return UserObjectToDict(p.loadAttributes(p.variant), p.variant.GetOriginalType(), p.variant.GetMaxDepth())
	case pb.Variant_VARIANT_NAMESPACE:
		return p.namespaceToDict(p.variant)
	case pb.Variant_VARIANT_ERROR:
		rookError := p.variant.GetErrorValue()
		return NewErrorNamespace(
			rookError.Message,
			NewProtobufNamespace(rookError.Parameters),
			NewProtobufNamespace(rookError.Exc),
			NewProtobufNamespace(rookError.Traceback)).ToDict()
	case pb.Variant_VARIANT_MAX_DEPTH:
		return maxDepthToDict()
	case pb.Variant_VARIANT_FORMATTED_MESSAGE:
		return NewFormattedNamespace(p.variant.GetMessageValue().Message).ToDict()
	case pb.Variant_VARIANT_UKNOWN_OBJECT:
		return UnknownObjectToDict(p.loadAttributes(p.variant), p.variant.GetOriginalType(), p.variant.GetMaxDepth())
	case pb.Variant_VARIANT_CODE_OBJECT:
		codeObject := p.variant.GetCodeValue()
		return codeObjectToDict(
			p.loadAttributes(p.variant),
			p.variant.GetOriginalType(),
			codeObject.Name,
			codeObject.Module,
			codeObject.Filename,
			codeObject.Lineno,
			p.variant.GetMaxDepth())
	case pb.Variant_VARIANT_DYNAMIC:
		return dynamicObjectToDict()
	}

	logrus.Warningf("Bad variant type - %d", p.variant.VariantType)
	return p.primitiveTypeToDict(nil, "null", p.variant)
}


func (p ProtobufNamespace) tracebackToSimpleDict() interface{} {
	var result []string

	basicList := p.listToSimpleDict(p.variant)
	if asMap, ok := basicList.([]interface{}); ok {
		for _, frame := range asMap {
			if frameAsMap, ok := frame.(map[string]interface{}); ok {
				if len(frameAsMap) != 4 {
					return nil
				}

				filename, ok := frameAsMap["filename"]
				if !ok {
					return nil
				}
				lineno, ok := frameAsMap["line"]
				if !ok {
					return nil
				}
				functionName, ok := frameAsMap["function"]
				if !ok {
					return nil
				}

				result = append(result, fmt.Sprintf("File %s, line %d, in %s", filename, lineno, functionName))
			} else {
				return nil
			}
		}
	} else {
		return nil
	}

	return result
}

func (p ProtobufNamespace) ToSimpleDict() interface{} {
	switch p.variant.VariantType {
	case pb.Variant_VARIANT_NONE:
		return nil
	case pb.Variant_VARIANT_INT:
		if "bool" == p.variant.OriginalType {
			return utils.IntAsBool(p.variant.GetIntValue())
		}
		return p.variant.GetIntValue()
	case pb.Variant_VARIANT_LONG:
		return int64ToSafeJsNumber(p.variant.GetLongValue())
	case pb.Variant_VARIANT_DOUBLE:
		if "float32" == p.variant.OriginalType {
			return float32(p.variant.GetDoubleValue())
		}
		return p.variant.GetDoubleValue()
	case pb.Variant_VARIANT_STRING:
		return p.variant.GetStringValue().Value
	case pb.Variant_VARIANT_BINARY:
		return getBinaryValueWithDefault(p.variant.GetBinaryValue().Value)
	case pb.Variant_VARIANT_TIME:
		return utils.ProtoToBackendCompatibleIsoTimeFormat(p.variant.GetTimeValue())
	case pb.Variant_VARIANT_LARGE_INT:
		return p.variant.GetLargeIntValue().Value
	case pb.Variant_VARIANT_COMPLEX:
		if "complex64" == p.variant.OriginalType {
			return complex(float32(p.variant.GetComplexValue().Real), float32(p.variant.GetComplexValue().Imaginary))
		}
		return complex(p.variant.GetComplexValue().Real, p.variant.GetComplexValue().Imaginary)
	case pb.Variant_VARIANT_UNDEFINED:
		return nil
	case pb.Variant_VARIANT_LIST:
		return p.listToSimpleDict(p.variant)
	case pb.Variant_VARIANT_MAP:
		return p.mapToSimpleDict(p.variant)
	case pb.Variant_VARIANT_OBJECT:
		return UserObjectToSimpleDict(p.loadAttributes(p.variant))
	case pb.Variant_VARIANT_NAMESPACE:
		return p.namespaceToSimpleDict(p.variant)
	case pb.Variant_VARIANT_ERROR:
		rookError := p.variant.GetErrorValue()
		if rookError == nil {
			logrus.Errorf("Error is nil: %v", p.variant)
			return NewErrorNamespace(
				"Failed to convert error",
				NewProtobufNamespace(nil),
				NewProtobufNamespace(nil),
				NewProtobufNamespace(nil)).ToSimpleDict()
		}
		return NewErrorNamespace(
			rookError.Message,
			NewProtobufNamespace(rookError.Parameters),
			NewProtobufNamespace(rookError.Exc),
			NewProtobufNamespace(rookError.Traceback)).ToSimpleDict()
	case pb.Variant_VARIANT_MAX_DEPTH:
		return maxDepthToSimpleDict()
	case pb.Variant_VARIANT_FORMATTED_MESSAGE:
		return NewFormattedNamespace(p.variant.GetMessageValue().Message).ToSimpleDict()
	case pb.Variant_VARIANT_UKNOWN_OBJECT:
		return UnknownObjectToSimpleDict()
	case pb.Variant_VARIANT_CODE_OBJECT:
		codeObject := p.variant.GetCodeValue()
		return codeObjectToSimpleDict(
			codeObject.Name,
			codeObject.Module)
	case pb.Variant_VARIANT_DYNAMIC:
		return dynamicObjectToSimpleDict()
	case pb.Variant_VARIANT_ENUM:
		return p.variant.GetEnumValue().GetStringValue()
	}

	logrus.Warningf("Bad variant type - %d", p.variant.VariantType)
	return p.variant
}

func getBinaryValueWithDefault(b []byte) interface{} {
	
	if len(b) == 0 {
		return ""
	}
	return b
}

func (p ProtobufNamespace) Filter(filters []types.FieldFilter) rookoutErrors.RookoutError {
	_, err := filterVariant(p.variant, filters)
	return err
}

func filteredValue(originalType string) *pb.Variant {
	filtered := new(pb.Variant)
	filtered.VariantType = pb.Variant_VARIANT_STRING
	filtered.OriginalType = originalType
	filtered.Value = &pb.Variant_StringValue{
		StringValue: &pb.Variant_String{
			OriginalSize: 0,
			Value:        FilteredFieldReplacement}}
	return filtered
}

func filterValue(v *pb.Variant) {
	v.VariantType = pb.Variant_VARIANT_STRING
	v.Value = &pb.Variant_StringValue{
		StringValue: &pb.Variant_String{
			OriginalSize: 0,
			Value:        FilteredFieldReplacement}}
}

func doesMatch(str []byte, wantedType types.FieldFilterType, filters []types.FieldFilter) int {
	for i, filter := range filters {
		if wantedType == filter.FilterType {
			if filter.Pattern.Match(str) {
				return i
			}
		}
	}
	return -1
}


func filterVariant(variant *pb.Variant, filters []types.FieldFilter) (bool, rookoutErrors.RookoutError) {
	if variant == nil {
		
		
		return false, nil
	}

	filtered := false

	if variant.Value != nil {
		switch variant.VariantType {
		case pb.Variant_VARIANT_STRING:
			value := []byte(variant.GetStringValue().Value)

			if i := doesMatch(value, FilterFieldValue, filters); i != -1 {
				value = filters[i].Pattern.ReplaceAll(value, []byte(FilteredValueReplacement))
				variant.GetStringValue().Value = string(value)
			}

		case pb.Variant_VARIANT_LIST:
			for _, value := range variant.GetListValue().Values {
				if internalFiltered, err := filterVariant(value, filters); err == nil {
					if filters[0].Whitelist {
						
						if !internalFiltered {
							
							filterValue(value)
						} else {
							filtered = true
						}
					}
				} else {
					return false, err
				}
			}

		case pb.Variant_VARIANT_MAP:
			ns := variant.GetMapValue()

			if ns == nil {
				return false, rookoutErrors.NewNilProtobufNamespaceException()
			}

			pairs := ns.Pairs

			for i, pair := range pairs {
				k := pair.First
				v := pair.Second
				if k.VariantType == pb.Variant_VARIANT_STRING {
					if index := doesMatch([]byte(k.GetStringValue().Value), FilterFieldName, filters); index != -1 {
						if filters[index].Whitelist == false {
							pairs[i] = &pb.Variant_Pair{First: k, Second: filteredValue(v.OriginalType)}
						} else {
							pairs[i] = &pb.Variant_Pair{First: k, Second: v}
							filtered = true
						}
						continue
					}
				}

				
				if internalFiltered, err := filterVariant(v, filters); err == nil {
					if filters[0].Whitelist {
						if !internalFiltered {
							
							pairs[i] = &pb.Variant_Pair{First: k, Second: filteredValue(v.OriginalType)}
						} else {
							filtered = true
						}
					}
				} else {
					return false, err
				}
			}

		case pb.Variant_VARIANT_NAMESPACE:
			ns := variant.GetNamespaceValue()

			if ns == nil {
				return false, rookoutErrors.NewNilProtobufNamespaceException()
			}

			attrs := ns.Attributes

			for i, attr := range attrs {
				name := attr.Name
				value := attr.Value

				if index := doesMatch([]byte(name), FilterFieldName, filters); index != -1 {
					if filters[index].Whitelist == false {
						attrs[i] = &pb.Variant_NamedValue{Name: name, Value: filteredValue(value.OriginalType)}
					} else {
						attrs[i] = &pb.Variant_NamedValue{Name: name, Value: value}
						filtered = true
					}
					continue
				}

				
				if internalFiltered, err := filterVariant(value, filters); err == nil {
					if filters[0].Whitelist {
						if !internalFiltered {
							
							attrs[i] = &pb.Variant_NamedValue{Name: name, Value: filteredValue(value.OriginalType)}
						} else {
							filtered = true
						}
					}
				} else {
					return false, err
				}
			}

		case pb.Variant_VARIANT_FORMATTED_MESSAGE:
			value := []byte(variant.GetMessageValue().Message)
			if index := doesMatch(value, FilterFieldValue, filters); index != -1 {
				value = filters[index].Pattern.ReplaceAll(value, []byte(FilteredValueReplacement))
				variant.GetStringValue().Value = string(value)
			}
		}
	}

	for _, attr := range variant.Attributes {
		name := attr.Name
		value := attr.Value
		if index := doesMatch([]byte(name), FilterFieldName, filters); index != -1 {
			if filters[index].Whitelist == false {
				
				newValue := filteredValue(value.OriginalType)
				attr.Value = newValue
			} else {
				
				filtered = true
			}
			continue
		}

		
		
		if internalFiltered, err := filterVariant(value, filters); err == nil {
			if filters[0].Whitelist {
				if !internalFiltered {
					
					newValue := filteredValue(value.OriginalType)
					attr.Value = newValue
				} else {
					filtered = true
				}
			} else {
				if _, err := filterVariant(value, filters); err != nil {
					return false, err
				}
			}
		} else {
			return false, err
		}

	}

	return filtered, nil
}
