package utils

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"reflect"
	"strings"
)

var IntTypes = []string{
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
}

var StringTypes = []string{
	"string",
}

var FloatTypes = []string{
	"float32", "float64",
}

var ComplexTypes = []string{
	"complex64", "complex128",
}

var basicTypes = []string{"bool",

	"byte", 
	"rune", 

	"func", "*runtime.Func",

	"array", "slice", 

	"struct", 

	"nil"} 

var allPrimitives = map[string]bool{}

func appendPrimitiveMap(list []string) {
	for _, v := range list {
		allPrimitives[v] = true
	}
}

func init() {
	appendPrimitiveMap(IntTypes)
	appendPrimitiveMap(StringTypes)
	appendPrimitiveMap(FloatTypes)
	appendPrimitiveMap(ComplexTypes)
	appendPrimitiveMap(basicTypes)
}

var knownType = []string{
	"<*runtime.Func Value>",
	"<*list.List Value>",
	"Time",
	"Func",
}

func GetTypeString(i interface{}) string {
	if nil == i {
		return "nil"
	}
	return reflect.TypeOf(i).Kind().String()
}

func GetValueString(i interface{}) string {
	if nil == i {
		return "nil"
	}
	return reflect.ValueOf(i).String()
}

func GetObjectType(obj interface{}) string {
	t := GetTypeString(obj)
	if t == "ptr" {
		t = GetValueString(obj)
	}
	return t
}

func IsPrimitiveType(obj interface{}) bool {
	objectType := GetObjectType(obj)

	_, ok := allPrimitives[objectType]
	return ok
}

func IsKnownType(obj interface{}) bool {
	x := GetObjectType(obj)
	if "struct" == x {
		x = reflect.TypeOf(obj).Name()
	}
	if _, ok := obj.(error); ok {
		return true
	}
	return Contains(knownType, x)
}

func GetType(i interface{}) reflect.Type {
	return reflect.ValueOf(i).Type()
}

func SameType(i1 interface{}, i2 interface{}) bool {
	return GetType(i1) == GetType(i2)
}

func ToString(i interface{}) (string, rookoutErrors.RookoutError) {
	if str, ok := i.(string); true == ok {
		return str, nil
	}
	return "", rookoutErrors.NewBadTypeException("interface is not string", i)
}

func ToBool(i interface{}) (bool, rookoutErrors.RookoutError) {
	if str, ok := i.(bool); true == ok {
		return str, nil
	}
	return false, rookoutErrors.NewBadTypeException("interface is not bool", i)
}

func ToFloat64(i interface{}) (float64, rookoutErrors.RookoutError) {
	if f, ok := i.(float64); true == ok {
		return f, nil
	}
	return 0, rookoutErrors.NewBadTypeException("interface is not float64", i)
}

func ToInt(i interface{}) (int, rookoutErrors.RookoutError) {
	v, ok := i.(int)

	if false == ok {
		
		f, ok := i.(float64)
		if false == ok {
			return 0, rookoutErrors.NewBadTypeException("interface is not float64", i)
		}
		return int(f), nil
	}

	return int(v), nil
}

func GetValueAsInterface(reflectedObject reflect.Value) interface{} {
	switch reflectedObject.Type().Kind() {
	case reflect.String:
		return reflectedObject.String()
	case reflect.Int, reflect.Int64:
		return reflectedObject.Int()
	case reflect.Float32, reflect.Float64:
		return reflectedObject.Float()
	case reflect.Bool:
		return reflectedObject.Bool()
	case reflect.Complex64, reflect.Complex128:
		return reflectedObject.Complex()
	case reflect.Interface:
		return reflectedObject.Interface()
	default:
		return nil
	}
}

func GetStructAttribute(i interface{}, attrib string) (interface{}, rookoutErrors.RookoutError) {
	reflectedValue := reflect.ValueOf(i)

	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem()
	}
	typeOfT := reflectedValue.Type()

	switch reflectedValue.Kind() {
	case reflect.Struct:
		field := reflectedValue.FieldByName(attrib)
		switch field.Kind() {
		case reflect.String:
			return field.String(), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int(), nil
		}

		for i := 0; i < reflectedValue.NumField(); i++ {
			f := reflectedValue.Field(i)

			if attrib == strings.ToLower(typeOfT.Field(i).Name) {
				return f.Interface(), nil
			}
		}

	}
	return nil, rookoutErrors.NewRookAttributeNotFoundException(attrib)
}
