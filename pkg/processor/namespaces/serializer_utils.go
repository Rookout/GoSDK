package namespaces

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/module"
)



type hchan struct {
	qcount   uint           
	dataqsiz uint           
	buf      unsafe.Pointer 
}

type Namespace interface {
	CallMethod(name string, args string) (Namespace, rookoutErrors.RookoutError)

	WriteAttribute(name string, value Namespace) rookoutErrors.RookoutError
	ReadAttribute(name string) (Namespace, rookoutErrors.RookoutError)

	ReadKey(key interface{}) (Namespace, rookoutErrors.RookoutError)
	GetObject() interface{}
	Serialize(serializer Serializer)
}

type Serializer interface {
	getCurrentDepth() int

	dumpOriginalType(originalType string)
	dumpType(t pb.Variant_Type)
	dumpMaxDepth(maxDepth bool)

	dumpTime(t time.Time, config config.ObjectDumpConfig)
	dumpNamespace(getNamedValue func(i int) (string, Namespace), numOfValues int)
	dumpTraceback(getFrame func(i int) (int, string, string), tracebackLen int)
	dumpFunc(functionName string, filename string, lineno int)
	dumpArray(getElem func(i int) (Namespace, bool), arrayLen int, config config.ObjectDumpConfig)
	dumpBinary(b []byte, config config.ObjectDumpConfig)
	dumpNil()
	dumpUnsupported()
	dumpRookoutError(err rookoutErrors.RookoutError)
	dumpErrorMessage(msg string)
	dumpEnum(desc string, ordinal int, typeName string)
	dumpMap(getKeyValue func(i int) (Namespace, Namespace, bool), mapLen int, config config.ObjectDumpConfig)
	dumpStruct(getField func(i int) (string, Namespace, bool), numOfFields int, config config.ObjectDumpConfig)
	dumpInt(i int64)
	dumpFloat(f float64)
	dumpComplex(c complex128)
	dumpBool(b bool)
	dumpString(s string, config config.ObjectDumpConfig)
	dumpStringLen(stringLen int)
	dumpChan(value reflect.Value, config config.ObjectDumpConfig)
}

func dumpValue(s Serializer, value reflect.Value, config config.ObjectDumpConfig) {
	defer func() {
		defer func() {
			recover()
		}()

		
		s.dumpOriginalType(value.Type().String())
	}()
	s.dumpMaxDepth(s.getCurrentDepth() >= config.MaxDepth)

	if !value.IsValid() {
		s.dumpNil()
		return
	}

	switch value.Kind() {
	case reflect.Chan, reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Func:
		if value.IsNil() {
			dumpUint(s, uint64(value.Pointer()), config)
			return
		}
	}

	switch value.Type() {
	case reflect.TypeOf(time.Time{}):
		dumpTimeValue(s, value, config)
		return
	case reflect.TypeOf(fmt.Errorf("")), reflect.TypeOf(&rookoutErrors.RookoutErrorImpl{}):
		dumpErrorValue(s, value)
		return
	case reflect.TypeOf([]byte{}):
		dumpBinaryValue(s, value, config)
		return
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.dumpInt(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		dumpUint(s, value.Uint(), config)
	case reflect.UnsafePointer:
		dumpUint(s, value.Addr().Uint(), config)
	case reflect.Float32, reflect.Float64:
		s.dumpFloat(value.Float())
	case reflect.Complex64, reflect.Complex128:
		s.dumpComplex(value.Complex())
	case reflect.String:
		s.dumpString(value.String(), config)
	case reflect.Bool:
		s.dumpBool(value.Bool())
	case reflect.Struct:
		dumpStructValue(s, value, config)
	case reflect.Array, reflect.Slice:
		dumpArrayValue(s, value, config)
	case reflect.Map:
		dumpMapValue(s, value, config)
	case reflect.Ptr:
		dumpValue(s, value.Elem(), config)
	case reflect.Interface:
		dumpValue(s, value.Elem(), config)
	case reflect.Chan:
		s.dumpChan(value, config)
	case reflect.Func:
		dumpFuncValue(s, value)
	case reflect.Invalid:
		s.dumpErrorMessage(fmt.Sprintf("Failed to dump type: %s", value.Type().String()))
	default:
		
		logger.Logger().Fatalf("Unknown kind: %s, type: %s", value.Kind().String(), value.Type().String())
	}
}

func dumpInterface(s Serializer, obj interface{}, config config.ObjectDumpConfig) {
	if nil == obj {
		s.dumpNil()
		return
	}

	dumpValue(s, reflect.ValueOf(obj), config)
}

func dumpUint(s Serializer, u uint64, config config.ObjectDumpConfig) {
	if u < math.MaxInt64 {
		s.dumpInt(int64(u))
	} else {
		s.dumpString(strconv.FormatUint(u, 10), config)
		s.dumpType(pb.Variant_VARIANT_LARGE_INT)
	}
}

var unixTimeMethod = func() reflect.Value {
	f, _ := reflect.TypeOf(time.Time{}).MethodByName("Unix")
	return f.Func
}()
var nsecTimeMethod = func() reflect.Value {
	f, _ := reflect.TypeOf(time.Time{}).MethodByName("Nanosecond")
	return f.Func
}()

func getTime(value reflect.Value) (t time.Time) {
	defer func() {
		
		if recover() != nil {
			unix := unixTimeMethod.Call([]reflect.Value{value})[0].Int()
			nsec := nsecTimeMethod.Call([]reflect.Value{value})[0].Int()
			t = time.Unix(unix, nsec)
		}
	}()

	i := value.Interface()
	if t, ok := i.(time.Time); ok {
		return t
	}
	if t, ok := i.(*time.Time); ok {
		return *t
	}
	return time.Time{}
}

func dumpTimeValue(s Serializer, value reflect.Value, config config.ObjectDumpConfig) {
	s.dumpTime(getTime(value), config)
}

func dumpArrayValue(s Serializer, value reflect.Value, config config.ObjectDumpConfig) {
	getElem := func(i int) (n Namespace, ok bool) {
		defer func() {
			if recover() != nil {
				n = NewValueNamespace(value.Index(i))
			}
		}()
		v := reflect.NewAt(value.Type().Elem(), unsafe.Pointer(value.Index(i).Addr().Pointer()))
		return NewGoObjectNamespace(v.Elem().Interface()), true
	}
	s.dumpArray(getElem, value.Len(), config)
}

func dumpBinaryValue(s Serializer, value reflect.Value, config config.ObjectDumpConfig) {
	b := make([]byte, value.Len())
	for i := range b {
		if i > config.MaxString {
			break
		}

		b[i] = *(*byte)(unsafe.Pointer(value.Pointer() + uintptr(i)))
	}
	s.dumpBinary(b, config)
}

func dumpStructValue(s Serializer, value reflect.Value, config config.ObjectDumpConfig) {
	numOfFields := value.NumField()
	getField := func(i int) (string, Namespace, bool) {
		fieldName := value.Type().Field(i).Name
		fieldValue := NewValueNamespace(value.Field(i))
		return fieldName, fieldValue, true
	}
	s.dumpStruct(getField, numOfFields, config)
}

func dumpMapValue(s Serializer, value reflect.Value, config config.ObjectDumpConfig) {
	mapKeys := value.MapKeys()
	getKeyValue := func(i int) (Namespace, Namespace, bool) {
		key := NewGoObjectNamespace(mapKeys[i].Interface())
		value := NewGoObjectNamespace(value.MapIndex(mapKeys[i]).Interface())
		return key, value, true
	}

	s.dumpMap(getKeyValue, value.Len(), config)
}

func dumpFuncValue(s Serializer, value reflect.Value) {
	funcAddr := value.Pointer()
	funcInfo := module.FindFunc(funcAddr)
	functionName := module.FuncName(funcInfo)

	
	s.dumpFunc(functionName, "", 0)
}

func dumpError(s Serializer, err error) {
	if r, ok := err.(rookoutErrors.RookoutError); ok {
		s.dumpRookoutError(r)
		return
	}
	s.dumpErrorMessage(err.Error())
}

func dumpErrorValue(s Serializer, value reflect.Value) {
	if value.Type() == reflect.TypeOf(&rookoutErrors.RookoutErrorImpl{}) {
		value = value.Elem()
	}

	if value.Type() == reflect.TypeOf(rookoutErrors.RookoutErrorImpl{}) {
		err := rookoutErrors.RookoutErrorImpl{
			ExternalError: value.FieldByName("ExternalError").Interface().(error),
			Type:          value.FieldByName("Type").Interface().(string),
			Arguments:     value.FieldByName("Arguments").Interface().(map[string]interface{}),
		}
		dumpError(s, err)
		return
	}

	errorFunc := value.MethodByName("Error")
	if !errorFunc.IsValid() {
		s.dumpNil()
		return
	}
	msg := errorFunc.Call([]reflect.Value{})[0]
	s.dumpErrorMessage(msg.String())
}
