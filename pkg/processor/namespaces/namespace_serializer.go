package namespaces

import (
	"reflect"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/config"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NamespaceSerializer struct { 
	*pb.Variant
	logErrors    bool
	currentDepth int
}

func NewNamespaceSerializer(fromNamespace Namespace, logErrors bool) *NamespaceSerializer {
	g := &NamespaceSerializer{Variant: &pb.Variant{}, logErrors: logErrors}
	fromNamespace.Serialize(g)
	return g
}

func (n *NamespaceSerializer) getCurrentDepth() int {
	return n.currentDepth
}

func (n *NamespaceSerializer) spawn(fromNamespace Namespace) *NamespaceSerializer {
	spawned := &NamespaceSerializer{Variant: &pb.Variant{}, logErrors: n.logErrors, currentDepth: n.currentDepth + 1}

	fromNamespace.Serialize(spawned)
	return spawned
}

func (n *NamespaceSerializer) spawnAtSameDepth(fromNamespace Namespace) *NamespaceSerializer {
	spawned := &NamespaceSerializer{Variant: &pb.Variant{}, logErrors: n.logErrors, currentDepth: n.currentDepth}
	fromNamespace.Serialize(spawned)
	return spawned
}

func (n *NamespaceSerializer) dumpOriginalType(originalType string) {
	n.OriginalType = originalType
}

func (n *NamespaceSerializer) dumpType(variantType pb.Variant_Type) {
	n.VariantType = variantType
}

func (n *NamespaceSerializer) dumpMaxDepth(maxDepth bool) {
	n.MaxDepth = maxDepth
}

func (n *NamespaceSerializer) dumpComplex(c complex128) {
	n.dumpType(pb.Variant_VARIANT_COMPLEX)
	n.Value = &pb.Variant_ComplexValue{ComplexValue: &pb.Variant_Complex{Real: float64(real(c)), Imaginary: float64(imag(c))}}
}

func (n *NamespaceSerializer) dumpFloat(f float64) {
	n.dumpType(pb.Variant_VARIANT_DOUBLE)
	n.Value = &pb.Variant_DoubleValue{DoubleValue: float64(f)}
}

func (n *NamespaceSerializer) dumpString(s string, config config.ObjectDumpConfig) {
	if !utf8.ValidString(s) {
		n.dumpBinary([]byte(s), config)
		return
	}

	n.dumpType(pb.Variant_VARIANT_STRING)
	originalSize := len(s)
	if len(s) > config.MaxString {
		s = s[:config.MaxString]
	}
	n.Value = &pb.Variant_StringValue{
		StringValue: &pb.Variant_String{
			OriginalSize: int32(originalSize),
			Value:        s,
		},
	}
}

func (n *NamespaceSerializer) dumpStringLen(stringLen int) {
	if _, ok := n.Value.(*pb.Variant_StringValue); ok {
		n.Value.(*pb.Variant_StringValue).StringValue.OriginalSize = int32(stringLen)
	}
	if _, ok := n.Value.(*pb.Variant_BinaryValue); ok {
		n.Value.(*pb.Variant_BinaryValue).BinaryValue.OriginalSize = int32(stringLen)
	}
}

func (n *NamespaceSerializer) dumpEnum(desc string, ordinal int, typeName string) {
	n.dumpType(pb.Variant_VARIANT_ENUM)
	n.Value = &pb.Variant_EnumValue{EnumValue: &pb.Variant_Enumeration{
		StringValue:  desc,
		OrdinalValue: int32(ordinal),
		TypeName:     typeName,
	}}
}

func (n *NamespaceSerializer) dumpBinary(b []byte, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_BINARY)
	originalSize := len(b)
	if len(b) > config.MaxString {
		b = b[:config.MaxString]
	}
	n.Value = &pb.Variant_BinaryValue{
		BinaryValue: &pb.Variant_Binary{
			OriginalSize: int32(originalSize),
			Value:        b,
		},
	}
}

func (n *NamespaceSerializer) dumpInt(i int64) {
	n.dumpType(pb.Variant_VARIANT_LONG)
	n.Value = &pb.Variant_LongValue{LongValue: i}
}

func (n *NamespaceSerializer) dumpBool(b bool) {
	n.dumpType(pb.Variant_VARIANT_INT)
	boolAsInt := int32(0)
	if b {
		boolAsInt = 1
	}
	n.Value = &pb.Variant_IntValue{IntValue: boolAsInt}
}

func (n *NamespaceSerializer) dumpTime(t time.Time, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_TIME)
	n.Value = &pb.Variant_TimeValue{
		TimeValue: &timestamppb.Timestamp{
			Seconds: t.Unix(),
			Nanos:   (int32)(t.Nanosecond()),
		},
	}
}

func (n *NamespaceSerializer) dumpArray(getElem func(i int) (Namespace, bool), arrayLen int, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_LIST)
	if n.getCurrentDepth() >= config.MaxCollectionDepth {
		n.dumpMaxDepth(true)
		return
	}

	listVariant := &pb.Variant_List{
		Type:         "list",
		OriginalSize: int32(arrayLen),
	}

	for i := 0; i < arrayLen; i++ {
		if i >= config.MaxWidth {
			break
		}

		e, ok := getElem(i)
		if !ok {
			break
		}

		listVariant.Values = append(listVariant.Values, n.spawn(e).Variant)
	}
	n.Value = &pb.Variant_ListValue{ListValue: listVariant}
}

func (n *NamespaceSerializer) dumpUnsupported() {
	n.dumpType(pb.Variant_VARIANT_UKNOWN_OBJECT)
}

func (n *NamespaceSerializer) dumpStruct(getField func(i int) (string, Namespace, bool), numOfFields int, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_OBJECT)

	if n.getCurrentDepth()+1 >= config.MaxDepth {
		n.dumpMaxDepth(true)
		return
	}

	n.Attributes = make([]*pb.Variant_NamedValue, 0, numOfFields)
	for i := 0; i < numOfFields; i++ {
		fieldName, fieldValue, ok := getField(i)
		if !ok {
			continue
		}
		n.Attributes = append(n.Attributes, &pb.Variant_NamedValue{Name: fieldName, Value: n.spawn(fieldValue).Variant})
	}
}

func (n *NamespaceSerializer) dumpMap(getKeyValue func(i int) (Namespace, Namespace, bool), mapLen int, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_MAP)
	if n.getCurrentDepth() >= config.MaxCollectionDepth {
		n.dumpMaxDepth(true)
		return
	}

	pairs := make([]*pb.Variant_Pair, 0)
	for i := 0; i < mapLen; i++ {
		if len(pairs) >= config.MaxWidth {
			break
		}

		key, value, ok := getKeyValue(i)
		if !ok {
			continue
		}
		pairs = append(pairs, &pb.Variant_Pair{
			First:  n.spawn(key).Variant,
			Second: n.spawn(value).Variant,
		})
	}

	n.Value = &pb.Variant_MapValue{
		MapValue: &pb.Variant_Map{
			OriginalSize: int32(mapLen),
			Pairs:        pairs,
		},
	}
}

func (n *NamespaceSerializer) dumpNil() {
	n.dumpType(pb.Variant_VARIANT_NONE)
}

func (n *NamespaceSerializer) dumpFunc(functionName string, filename string, lineno int) {
	n.dumpType(pb.Variant_VARIANT_CODE_OBJECT)
	n.Value = &pb.Variant_CodeValue{
		CodeValue: &pb.Variant_CodeObject{
			Name:     functionName,
			Filename: filename,
			Lineno:   uint32(lineno),
		},
	}
}

func (n *NamespaceSerializer) dumpChan(value reflect.Value, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_LIST)

	if n.getCurrentDepth() >= config.MaxCollectionDepth {
		n.dumpMaxDepth(true)
		return
	}

	addr := utils.UnsafePointer(value)
	chanStruct := *(*hchan)(addr)
	elemType := value.Type().Elem()
	bufSize := value.Len() * int(elemType.Size())

	listVariant := &pb.Variant_List{
		Type:         "list",
		OriginalSize: int32(value.Len()),
	}

	for i := 0; i < bufSize; i += int(elemType.Size()) {
		if len(listVariant.Values) >= config.MaxWidth {
			break
		}

		ptr := unsafe.Pointer(uintptr(chanStruct.buf) + uintptr(i))
		val := reflect.NewAt(elemType, ptr).Elem()
		valVariant := &NamespaceSerializer{Variant: &pb.Variant{}, logErrors: n.logErrors, currentDepth: n.currentDepth + 1}
		dumpValue(valVariant, val, config)
		listVariant.Values = append(listVariant.Values, valVariant.Variant)
	}
	n.Value = &pb.Variant_ListValue{ListValue: listVariant}
}

func (n *NamespaceSerializer) dumpRookoutError(r rookoutErrors.RookoutError) {
	n.dumpType(pb.Variant_VARIANT_ERROR)

	stackFramesObj := NewGoObjectNamespace(string(r.Stack()))
	stackFramesObj.ObjectDumpConf = config.TailorObjectDumpConfig(reflect.String, len(r.Stack()))

	n.Value = &pb.Variant_ErrorValue{
		ErrorValue: &pb.Error{
			Message:    r.Error(),
			Type:       r.GetType(),
			Parameters: n.spawn(NewGoObjectNamespace(r.GetArguments())).Variant,
			Traceback:  n.spawn(stackFramesObj).Variant,
		},
	}
}

func (n *NamespaceSerializer) dumpErrorMessage(msg string) {
	n.dumpType(pb.Variant_VARIANT_ERROR)
	n.Value = &pb.Variant_ErrorValue{
		ErrorValue: &pb.Error{
			Message: msg,
		},
	}
}

func (n *NamespaceSerializer) dumpNamespace(getNamedValue func(i int) (string, Namespace), numOfValues int) {
	n.dumpType(pb.Variant_VARIANT_NAMESPACE)

	if n.MaxDepth {
		return
	}

	attributes := make([]*pb.Variant_NamedValue, 0, numOfValues)
	for i := 0; i < numOfValues; i++ {
		name, value := getNamedValue(i)
		attributes = append(attributes, &pb.Variant_NamedValue{
			Name:  name,
			Value: n.spawnAtSameDepth(value).Variant,
		})
	}
	n.Value = &pb.Variant_NamespaceValue{
		NamespaceValue: &pb.Variant_Namespace{
			Attributes: attributes,
		},
	}
}

func (n *NamespaceSerializer) dumpTraceback(getFrame func(i int) (int, string, string), tracebackLen int) {
	n.dumpType(pb.Variant_VARIANT_TRACEBACK)

	tracebackVariant := &pb.Variant_Traceback{}
	for i := 0; i < tracebackLen; i++ {
		lineno, filename, functionName := getFrame(i)
		tracebackVariant.Locations = append(tracebackVariant.Locations, &pb.Variant_CodeObject{
			Lineno:   uint32(lineno),
			Filename: filename,
			Name:     functionName,
			Module:   filename,
		})
	}

	n.Value = &pb.Variant_Traceback_{
		Traceback: tracebackVariant,
	}
}
