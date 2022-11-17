package namespaces

import (
	"container/list"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/config"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"go/constant"
	"math"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
	"unicode/utf8"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/sirupsen/logrus"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
)


var JSMaxSafe = int64(math.Pow(2, 53))

func int64ToSafeJsNumber(num int64) interface{} {
	if num < JSMaxSafe && num > -1*JSMaxSafe {
		
		return num
	}
	return fmt.Sprintf("%d", num)
}

func dumpPrimitive(obj interface{}, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) rookoutErrors.RookoutError {
	if nil == obj {
		v.VariantType = pb.Variant_VARIANT_NONE
		return nil
	}

	kind := reflect.TypeOf(obj).Kind()
	switch {
	case (kind == reflect.Int) && (obj.(int) < intMax) && (obj.(int) > intMin):
		v.VariantType = pb.Variant_VARIANT_INT
		v.Value = &pb.Variant_IntValue{IntValue: int32(obj.(int))}
	case (kind == reflect.Int32) && (obj.(int32) < intMax) && (obj.(int32) > intMin):
		v.VariantType = pb.Variant_VARIANT_INT
		v.Value = &pb.Variant_IntValue{IntValue: obj.(int32)}
	case kind == reflect.Bool:
		
		v.VariantType = pb.Variant_VARIANT_INT
		v.Value = &pb.Variant_IntValue{IntValue: int32(utils.BoolAsInt(obj.(bool)))}
	case (kind == reflect.Int64) && (obj.(int64) < longMax) && (obj.(int64) > longMin):
		v.VariantType = pb.Variant_VARIANT_LONG
		v.Value = &pb.Variant_LongValue{LongValue: obj.(int64)}
	case kind == reflect.Float32:
		v.VariantType = pb.Variant_VARIANT_DOUBLE
		v.Value = &pb.Variant_DoubleValue{DoubleValue: float64(obj.(float32))}
	case kind == reflect.Float64:
		v.VariantType = pb.Variant_VARIANT_DOUBLE
		v.Value = &pb.Variant_DoubleValue{DoubleValue: obj.(float64)}
	case kind == reflect.Complex64:
		v.VariantType = pb.Variant_VARIANT_COMPLEX
		v.Value = &pb.Variant_ComplexValue{ComplexValue: &pb.Variant_Complex{Real: float64(real(obj.(complex64))), Imaginary: float64(imag(obj.(complex64)))}}
	case kind == reflect.Complex128:
		v.VariantType = pb.Variant_VARIANT_COMPLEX
		v.Value = &pb.Variant_ComplexValue{ComplexValue: &pb.Variant_Complex{Real: float64(real(obj.(complex128))), Imaginary: float64(imag(obj.(complex128)))}}
	case kind == reflect.String:
		v.VariantType = pb.Variant_VARIANT_STRING

		
		objAsString := fmt.Sprintf("%v", obj)

		originalSize := len(objAsString)
		if config.MaxString < originalSize {
			v.Value = &pb.Variant_StringValue{
				StringValue: &pb.Variant_String{
					OriginalSize: int32(originalSize),
					Value:        objAsString[0:config.MaxString]}}
		} else {
			v.Value = &pb.Variant_StringValue{
				StringValue: &pb.Variant_String{
					OriginalSize: int32(originalSize),
					Value:        objAsString}}
		}
	case kind == reflect.Slice:
		val := reflect.ValueOf(obj)
		originalSize := val.Len()

		if config.MaxString < originalSize {
			originalSize = config.MaxString
		}

		if _, ok := obj.([]byte); ok {
			newRange := val.Bytes()
			newRange = newRange[:originalSize]

			v.VariantType = pb.Variant_VARIANT_BINARY
			v.Value = &pb.Variant_BinaryValue{
				BinaryValue: &pb.Variant_Binary{
					OriginalSize: int32(val.Len()),
					Value:        newRange,
				}}
		} else {
			l := list.New()
			for i := 0; i < originalSize; i++ {
				l.PushBack(val.Index(i).Interface())
			}

			err := dumpList(l, v, currentDepth+1, config, logErrors)
			if err != nil {
				return err
			}
		}

	case kind == reflect.Array:

		val := reflect.ValueOf(obj)
		originalSize := val.Len()

		if config.MaxString < originalSize {
			originalSize = config.MaxString
		}

		typeString := ""
		if 0 != originalSize {
			typeString = utils.GetTypeString(val.Index(0).Interface())
		}

		if "uint8" == typeString {
			newRange := make([]byte, originalSize)
			for i := 0; i < originalSize; i++ {
				newRange[i] = byte(val.Index(i).Uint())
			}

			v.VariantType = pb.Variant_VARIANT_BINARY
			v.Value = &pb.Variant_BinaryValue{
				BinaryValue: &pb.Variant_Binary{
					OriginalSize: int32(val.Len()),
					Value:        newRange,
				}}
		} else {
			l := list.New()
			for i := 0; i < originalSize; i++ {
				l.PushBack(val.Index(i).Interface())
			}

			err := dumpList(l, v, currentDepth+1, config, logErrors)
			if err != nil {
				return err
			}
		}

	case kind == reflect.Struct:
		if "Time" == reflect.TypeOf(obj).Name() {
			dumpTime(obj, v)
			break
		}

		v.VariantType = pb.Variant_VARIANT_OBJECT
		v.Value = &pb.Variant_MessageValue{}
	case kind == reflect.Func:
		v.VariantType = pb.Variant_VARIANT_CODE_OBJECT

		val := reflect.ValueOf(obj)
		f := runtime.FuncForPC(val.Pointer())

		file, line := f.FileLine(val.Pointer())

		functionName, err := utils.GetFunctionName(f.Name())
		if nil != err {
			return err
		}

		fileName := filepath.Base(file)

		v.Value = &pb.Variant_CodeValue{
			CodeValue: &pb.Variant_CodeObject{
				Name:     functionName,
				Filename: fileName,
				Lineno:   uint32(line),
			},
		}
	default:
		if logErrors {
			logrus.Error("Failed to dump primitive")
		}

		return rookoutErrors.NewNotImplemented()
	}

	return nil
}

func dumpAttributes(obj interface{}, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) rookoutErrors.RookoutError {
	reflectedValue := reflect.ValueOf(obj)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}

	numberOfFields := reflectedValue.NumField()
	for i := 0; i < numberOfFields; i++ {
		newAttribute := pb.Variant_NamedValue{}
		newAttribute.Name = reflectedValue.Type().Field(i).Name
		newAttribute.Value = &pb.Variant{}
		if err := dumpGoObject(utils.GetValueAsInterface(reflectedValue.Field(i)), newAttribute.Value, currentDepth+1, config, logErrors); nil != err {
			return err
		}

		v.Attributes = append(v.Attributes, &newAttribute)
	}
	return nil
}

func convertFloatValue(v *variable.Variable) float64 {
	switch v.FloatSpecial {
	case variable.FloatIsPosInf:
		return math.Inf(1)
	case variable.FloatIsNegInf:
		return math.Inf(-1)
	case variable.FloatIsNaN:
		return math.NaN()
	}
	f, _ := constant.Float64Val(v.Value)
	return f
}


func dumpVariable(v *variable.Variable) *pb.Variant {
	variant := &pb.Variant{}
	variant.OriginalType = PrettyTypeName(v.DwarfType) 
	if v.Unreadable != nil {
		variant.VariantType = pb.Variant_VARIANT_ERROR
		variant.Value = &pb.Variant_ErrorValue{
			ErrorValue: &pb.Error{
				Message: fmt.Sprintf("unreadable var: %s", v.Unreadable.Error()),
				Type:    "error",
			},
		}

		return variant
	}
	if v.OnlyAddr {
		variant.VariantType = pb.Variant_VARIANT_MAX_DEPTH
		variant.MaxDepth = true
		return variant
	}

	if cd := v.ConstDescr(); cd != "" {
		variant.VariantType = pb.Variant_VARIANT_ENUM
		enumValue := &pb.Variant_EnumValue{EnumValue: &pb.Variant_Enumeration{
			StringValue:  cd,
			OrdinalValue: 0,
			TypeName:     variant.OriginalType,
		}}

		switch v.Kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, _ := constant.Int64Val(v.Value)
			enumValue.EnumValue.OrdinalValue = int32(n)
		}

		variant.Value = enumValue
		return variant
	}

	switch v.Kind {
	case reflect.Complex64, reflect.Complex128:
		r, _ := constant.Float64Val(constant.Real(v.Value))

		i, _ := constant.Float64Val(constant.Imag(v.Value))
		variant.VariantType = pb.Variant_VARIANT_COMPLEX
		if v.Value != nil {
			variant.Value = &pb.Variant_ComplexValue{ComplexValue: &pb.Variant_Complex{Real: r, Imaginary: i}}
		}
		return variant
	case reflect.Float32, reflect.Float64:
		variant.VariantType = pb.Variant_VARIANT_DOUBLE
		variant.Value = &pb.Variant_DoubleValue{DoubleValue: convertFloatValue(v)}
		return variant
	case reflect.String:
		variant.VariantType = pb.Variant_VARIANT_STRING
		if v.Value != nil {
			s := constant.StringVal(v.Value)
			if utf8.ValidString(s) {
				variant.Value = &pb.Variant_StringValue{StringValue: &pb.Variant_String{Value: s, OriginalSize: int32(v.Len)}}
			} else {
				variant.VariantType = pb.Variant_VARIANT_BINARY
				variant.Value = &pb.Variant_BinaryValue{BinaryValue: &pb.Variant_Binary{Value: []byte(s), OriginalSize: int32(v.Len)}}
			}
		}
		return variant
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, _ := constant.Int64Val(v.Value)
		variant.VariantType = pb.Variant_VARIANT_LONG
		variant.Value = &pb.Variant_LongValue{LongValue: n}
		return variant
	case reflect.Bool:
		variant.VariantType = pb.Variant_VARIANT_INT
		boolVal := constant.BoolVal(v.Value)
		boolAsInt := int32(0)
		if boolVal {
			boolAsInt = 1
		}
		variant.Value = &pb.Variant_IntValue{IntValue: boolAsInt}
		return variant
	case reflect.Map:
		variant.VariantType = pb.Variant_VARIANT_MAP
		pairs := make([]*pb.Variant_Pair, 0, len(v.Children)/2)
		
		for i := 0; i < len(v.Children); i += 2 {
			key := v.Children[i]
			value := v.Children[i+1]
			pair := &pb.Variant_Pair{
				First:  dumpVariable(key),
				Second: dumpVariable(value),
			}
			pairs = append(pairs, pair)
		}
		variant.Value = &pb.Variant_MapValue{
			MapValue: &pb.Variant_Map{
				OriginalSize: int32(v.Len),
				Pairs:        pairs}}
		return variant
	case reflect.Slice, reflect.Array:
		variant.VariantType = pb.Variant_VARIANT_LIST
		if v.Base == 0 && len(v.Children) == 0 {
			
			return variant
		}

		listVariant := &pb.Variant_List{
			Type:         "list",
			OriginalSize: int32(v.Len)}

		for _, child := range v.Children {
			childVariant := dumpVariable(child)
			listVariant.Values = append(listVariant.Values, childVariant)
		}
		variant.Value = &pb.Variant_ListValue{ListValue: listVariant}
		return variant
	case reflect.Ptr:
		if len(v.Children) == 0 || v.Children[0].Addr == 0 {
			variant.VariantType = pb.Variant_VARIANT_NONE
			return variant
		}

		if v.Children[0].OnlyAddr {
			variant.VariantType = pb.Variant_VARIANT_MAX_DEPTH
			variant.MaxDepth = true
			return variant
		}

		
		child := dumpVariable(v.Children[0])
		child.OriginalType = variant.OriginalType
		return child
	case reflect.UnsafePointer:
		if len(v.Children) == 0 {
			variant.VariantType = pb.Variant_VARIANT_NONE
			return variant
		}
		variant.VariantType = pb.Variant_VARIANT_STRING
		str := fmt.Sprintf("unsafe.Pointer(%#x)", v.Children[0].Addr)
		variant.Value = &pb.Variant_StringValue{StringValue: &pb.Variant_String{
			OriginalSize: int32(len(str)),
			Value:        str,
		}}
		return variant
	case reflect.Func:
		addr := v.Base
		if addr == 0 {
			addr = v.Addr
		}
		variant.VariantType = pb.Variant_VARIANT_CODE_OBJECT
		variant.Value = &pb.Variant_CodeValue{CodeValue: &pb.Variant_CodeObject{
			Name:     v.FunctionName,
			Filename: v.FileName,
			Lineno:   uint32(v.Line),
		}}
		return variant
	case reflect.Interface:
		if v.Addr == 0 || len(v.Children) == 0 {
			
			
			variant.VariantType = pb.Variant_VARIANT_NONE
			return variant
		}
		data := v.Children[0]

		if data.OnlyAddr {
			variant.VariantType = pb.Variant_VARIANT_MAX_DEPTH
			variant.MaxDepth = true
			return variant
		}

		
		child := dumpVariable(data)
		child.OriginalType = fmt.Sprintf("%s (%s)", variant.OriginalType, child.OriginalType)
		fallthrough 
	case reflect.Struct:
		variant.VariantType = pb.Variant_VARIANT_OBJECT
		variant.Attributes = make([]*pb.Variant_NamedValue, 0)
		for _, child := range v.Children {
			val := dumpVariable(child)
			variant.Attributes = append(variant.Attributes, &pb.Variant_NamedValue{Name: child.Name, Value: val})
		}
		return variant
	case reflect.Chan:
		variant.VariantType = pb.Variant_VARIANT_LIST
		if len(v.Children) == 0 {
			
			return variant
		}

		listVariant := &pb.Variant_List{
			Type:         "list",
			OriginalSize: int32(v.Len)}

		for _, child := range v.Children {
			childVariant := dumpVariable(child)
			listVariant.Values = append(listVariant.Values, childVariant)
		}
		variant.Value = &pb.Variant_ListValue{ListValue: listVariant}

		return variant
	}

	variant.VariantType = pb.Variant_VARIANT_NOT_SUPPORTED
	logger.Logger().Tracef("unsupported variant type: %v (%s)", v.Kind, variant.OriginalType)

	return variant
}

func dumpRookoutError(r rookoutErrors.RookoutError, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) rookoutErrors.RookoutError {
	v.VariantType = pb.Variant_VARIANT_ERROR

	errorVariant := &pb.Error{
		Message: r.Error(),
		Type:    r.GetType(),
	}
	parameters := &pb.Variant{}
	err := dumpGoObject(r.GetArguments(), parameters, currentDepth+1, config, logErrors)
	if err != nil {
		return err
	}
	traceback := &pb.Variant{}
	err = dumpGoObject(r.StackFrames(), traceback, currentDepth+1, config, logErrors)
	if err != nil {
		return err
	}
	errorVariant.Parameters = parameters
	errorVariant.Traceback = traceback

	v.Value = &pb.Variant_ErrorValue{
		ErrorValue: errorVariant,
	}
	return nil
}

func dumpGoObject(obj interface{}, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) (err rookoutErrors.RookoutError) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				e = fmt.Errorf("panic: %v", r)
			}

			err = rookoutErrors.NewRuntimeError(e.Error())
			v = GetErrorVariant(err, logErrors)
		}
	}()

	if currentDepth == config.MaxDepth {
		v.VariantType = pb.Variant_VARIANT_MAX_DEPTH
	} else {
		if err = dumpBaseObject(obj, v, currentDepth, config, logErrors); nil != err {
			return
		}

		if utils.IsPrimitiveType(obj) {
			if err = dumpPrimitive(obj, v, currentDepth, config, logErrors); nil != err {
				return
			}
		} else {
			if v.OriginalType == "map" {
				obj = *utils.InterfaceToMap(obj)
			}
			switch typedObj := obj.(type) {
			case rookoutErrors.RookoutError:
				err = dumpRookoutError(typedObj, v, currentDepth, config, logErrors)
			case *list.List:
				err = dumpList(typedObj, v, currentDepth, config, logErrors)
			case map[interface{}]interface{}:
				err = dumpMap(typedObj, v, currentDepth, config, logErrors)
			case error:
				v.VariantType = pb.Variant_VARIANT_ERROR
			default:
				v.VariantType = pb.Variant_VARIANT_UKNOWN_OBJECT
			}
		}
	}
	return
}

func dumpTime(obj interface{}, v *pb.Variant) {
	v.VariantType = pb.Variant_VARIANT_TIME
	v.Value = &pb.Variant_TimeValue{
		TimeValue: &google_protobuf.Timestamp{
			Seconds: obj.(time.Time).Unix(),
			Nanos:   (int32)(obj.(time.Time).Nanosecond())}}
}

func dumpMap(m map[interface{}]interface{}, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) rookoutErrors.RookoutError {
	v.VariantType = pb.Variant_VARIANT_MAP

	i := 0
	pairs := make([]*pb.Variant_Pair, 0)
	for k, val := range m {
		if i >= config.MaxWidth {
			break
		}

		keyVariant := pb.Variant{}
		if err := dumpGoObject(k, &keyVariant, currentDepth+1, config, logErrors); nil != err {
			return err
		}

		valueVariant := pb.Variant{}
		if err := dumpGoObject(val, &valueVariant, currentDepth+1, config, logErrors); nil != err {
			return err
		}

		pairs = append(pairs, &pb.Variant_Pair{
			First:  &keyVariant,
			Second: &valueVariant,
		})

		i++
	}

	v.Value = &pb.Variant_MapValue{
		MapValue: &pb.Variant_Map{
			OriginalSize: int32(len(m)),
			Pairs:        pairs}}

	return nil
}

func dumpList(l *list.List, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) rookoutErrors.RookoutError {
	v.VariantType = pb.Variant_VARIANT_LIST

	listVariant := &pb.Variant_List{
		Type:         "list",
		OriginalSize: (int32)(l.Len())}

	
	if currentDepth < config.MaxCollectionDepth {
		i := 0

		for e := l.Front(); e != nil; e = e.Next() {
			if i >= config.MaxWidth {
				break
			}

			enumeratedVariant := pb.Variant{}
			if err := dumpGoObject(e.Value, &enumeratedVariant, currentDepth+1, config, logErrors); nil != err {
				return err
			}

			listVariant.Values = append(listVariant.Values, &enumeratedVariant)

			i++
		}
	}

	v.Value = &pb.Variant_ListValue{
		ListValue: listVariant}

	return nil
}

func dumpBaseObject(obj interface{}, v *pb.Variant, currentDepth int, config config.ObjectDumpConfig, logErrors bool) rookoutErrors.RookoutError {
	v.OriginalType = utils.GetTypeString(obj)

	switch v.OriginalType {
	case "struct":
		v.OriginalType = reflect.TypeOf(obj).Name()

		if !utils.IsKnownType(obj) {
			if err := dumpAttributes(obj, v, currentDepth, config, logErrors); nil != err {
				return err
			}
		}

	case "ptr":
		v.OriginalType = utils.GetValueString(obj)
		if !utils.IsKnownType(obj) {
			if err := dumpAttributes(obj, v, currentDepth, config, logErrors); nil != err {
				return err
			}
		}
	}

	return nil
}

func getAttributesDict(attributes map[string]types.Namespace) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range attributes {
		result[k] = v.ToDict()
	}

	return result
}

func getAttributesSimpleDict(Attributes map[string]types.Namespace) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range Attributes {
		result[k] = v.ToSimpleDict()
	}

	return result
}
func codeObjectToDict(
	attributes map[string]types.Namespace,
	originalType string,
	Name string,
	Module string,
	Filename string,
	Lineno uint32,
	maxDepth bool) map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "CodeObjectNamespace",
		"@common_type":   "code",
		"@original_type": originalType,
		"@attributes":    getAttributesDict(attributes),
		"@max_depth":     maxDepth,
		"@value": map[string]interface{}{
			"name":     Name,
			"module":   Module,
			"filename": Filename,
			"lineno":   Lineno,
		},
	}
}

func codeObjectToSimpleDict(Name string,
	Module string) interface{} {
	return fmt.Sprintf("%s @ %s", Name, Module)
}

func dynamicObjectToDict() map[string]interface{} {
	return map[string]interface{}{
		"@namespace":   "DynamicObjectNamespace",
		"@common_type": "dynamic",
	}
}

func dynamicObjectToSimpleDict() interface{} {
	return "<Dynamic>"
}

func maxDepthToDict() map[string]interface{} {
	return map[string]interface{}{
		"@namespace":   "MaxDepthNamespace",
		"@common_type": "Namespace",
	}
}

func maxDepthToSimpleDict() interface{} {
	return "<MaxDepthLimit>"
}

func UnknownObjectToDict(attributes map[string]types.Namespace, originalType string, maxDepth bool) map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "UnknownNamespace",
		"@common_type":   "UnknownObject",
		"@original_type": originalType,
		"@max_depth":     maxDepth,
		"@attributes":    getAttributesDict(attributes),
	}
}

func UnknownObjectToSimpleDict() interface{} {
	return "<Unknown>"
}

func UserObjectToDict(attributes map[string]types.Namespace, originalType string, maxDepth bool) map[string]interface{} {
	return map[string]interface{}{
		"@namespace":     "UserObjectNamespace",
		"@common_type":   "UserObject",
		"@original_type": originalType,
		"@max_depth":     maxDepth,
		"@attributes":    getAttributesDict(attributes),
	}
}

func UserObjectToSimpleDict(attributes map[string]types.Namespace) interface{} {
	return getAttributesSimpleDict(attributes)
}

func GetErrorVariant(rookoutError rookoutErrors.RookoutError, logErrors bool) *pb.Variant {
	v := &pb.Variant{}
	v.VariantType = pb.Variant_VARIANT_ERROR

	v.Value = &pb.Variant_ErrorValue{
		ErrorValue: &pb.Error{
			Message:   rookoutError.Error(),
			Type:      "error",
			Traceback: NewGoObjectNamespace(rookoutError.Stack()).ToProtobuf(logErrors),
		},
	}

	return v
}


func recoverFromPanic(recoverMessage interface{}, v *pb.Variant, logErrors bool) {
	if recoverMessage != nil {
		e, ok := recoverMessage.(error)
		if !ok {
			e = fmt.Errorf("panic: %v", recoverMessage)
		}

		err := rookoutErrors.NewRuntimeError(e.Error())
		v = GetErrorVariant(err, false)

		if logErrors {
			logger.Logger().Errorf("%s\n%s", e.Error(), string(debug.Stack()))
		}
	}
}

func PrettyTypeName(typ godwarf.Type) string {
	if typ == nil {
		return ""
	}
	if typ.Common().Name != "" {
		return typ.Common().Name
	}
	r := typ.String()
	if r == "*void" {
		return "unsafe.Pointer"
	}
	return r
}
