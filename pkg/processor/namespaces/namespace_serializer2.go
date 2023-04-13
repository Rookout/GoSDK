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

type NamespaceSerializer2 struct {
	*pb.Variant2
	StringCache  map[string]uint32
	logErrors    bool
	currentDepth int
}

func NewNamespaceSerializer2(fromNamespace Namespace, logErrors bool) *NamespaceSerializer2 {
	g := &NamespaceSerializer2{StringCache: make(map[string]uint32), Variant2: &pb.Variant2{}, logErrors: logErrors}
	fromNamespace.Serialize(g)
	return g
}

func (n *NamespaceSerializer2) spawn(fromNamespace Namespace) *NamespaceSerializer2 {
	spawned := &NamespaceSerializer2{StringCache: n.StringCache, Variant2: &pb.Variant2{}, logErrors: n.logErrors, currentDepth: n.currentDepth + 1}

	fromNamespace.Serialize(spawned)
	return spawned
}

func (n *NamespaceSerializer2) spawnAtSameDepth(fromNamespace Namespace) *NamespaceSerializer2 {
	spawned := &NamespaceSerializer2{StringCache: n.StringCache, Variant2: &pb.Variant2{}, logErrors: n.logErrors, currentDepth: n.currentDepth}
	fromNamespace.Serialize(spawned)
	return spawned
}

func (n *NamespaceSerializer2) dumpOriginalType(originalType string) {
	n.OriginalTypeIndexInCache = n.getStringIndexInCache(originalType)
}

func (n *NamespaceSerializer2) dumpTypeMaxDepth(variantType uint32, maxDepthInt uint32) {
	n.VariantTypeMaxDepth = uint32((variantType << 1) | maxDepthInt)
}

func (n *NamespaceSerializer2) dumpType(variantType pb.Variant_Type) {
	maxDepthInt := n.VariantTypeMaxDepth & 1
	n.dumpTypeMaxDepth(uint32(variantType), maxDepthInt)
}

func (n *NamespaceSerializer2) dumpMaxDepth(maxDepth bool) {
	maxDepthInt := 0
	if maxDepth {
		maxDepthInt = 1
	}
	variantType := n.VariantTypeMaxDepth >> 1
	n.dumpTypeMaxDepth(variantType, uint32(maxDepthInt))
}

func (n *NamespaceSerializer2) getCurrentDepth() int {
	return n.currentDepth
}

func (n *NamespaceSerializer2) getStringIndexInCache(s string) uint32 {
	if i, ok := n.StringCache[s]; ok {
		return i
	}

	n.StringCache[s] = uint32(len(n.StringCache))
	return n.StringCache[s]
}

func (n *NamespaceSerializer2) dumpUnsupported() {
	n.dumpType(pb.Variant_VARIANT_UKNOWN_OBJECT)
}

func (n *NamespaceSerializer2) dumpComplex(c complex128) {
	n.dumpType(pb.Variant_VARIANT_COMPLEX)
	n.ComplexValue = &pb.Variant_Complex{Real: real(c), Imaginary: imag(c)}
}

func (n *NamespaceSerializer2) dumpFloat(f float64) {
	n.dumpType(pb.Variant_VARIANT_DOUBLE)
	n.DoubleValue = f
}

func (n *NamespaceSerializer2) dumpString(s string, config config.ObjectDumpConfig) {
	if !utf8.ValidString(s) {
		n.dumpBinary([]byte(s), config)
		return
	}
	n.dumpType(pb.Variant_VARIANT_STRING)
	n.OriginalSize = uint32(len(s))
	if len(s) > config.MaxString {
		s = s[:config.MaxString]
	}
	n.BytesIndexInCache = n.getStringIndexInCache(s)
}

func (n *NamespaceSerializer2) dumpStringLen(stringLen int) {
	n.OriginalSize = uint32(stringLen)
}

func (n *NamespaceSerializer2) dumpBinary(b []byte, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_BINARY)
	n.OriginalSize = uint32(len(b))
	if len(b) > config.MaxString {
		b = b[:config.MaxString]
	}
	n.BytesIndexInCache = n.getStringIndexInCache(string(b))
}

func (n *NamespaceSerializer2) dumpInt(i int64) {
	n.dumpType(pb.Variant_VARIANT_LONG)
	n.LongValue = i
}

func (n *NamespaceSerializer2) dumpBool(b bool) {
	n.dumpType(pb.Variant_VARIANT_LONG)
	if b {
		n.LongValue = 1
	} else {
		n.LongValue = 0
	}
}

func (n *NamespaceSerializer2) dumpTime(t time.Time, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_TIME)
	n.TimeValue = timestamppb.New(t)
}

func (n *NamespaceSerializer2) dumpArray(getElem func(i int) Namespace, arrayLen int, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_LIST)
	n.OriginalSize = uint32(arrayLen)

	if n.getCurrentDepth() >= config.MaxCollectionDepth {
		n.dumpMaxDepth(true)
		return
	}

	for i := 0; i < arrayLen; i++ {
		if i >= config.MaxWidth {
			return
		}

		n.CollectionValues = append(n.CollectionValues, n.spawn(getElem(i)).Variant2)
	}
}

func (n *NamespaceSerializer2) dumpNamespace(getNamedValue func(i int) (string, Namespace), numOfValues int) {
	n.dumpType(pb.Variant_VARIANT_NAMESPACE)

	for i := 0; i < numOfValues; i++ {
		name, value := getNamedValue(i)
		n.AttributeNamesInCache = append(n.AttributeNamesInCache, n.getStringIndexInCache(name))
		n.AttributeValues = append(n.AttributeValues, n.spawnAtSameDepth(value).Variant2)
	}
}

func (n *NamespaceSerializer2) dumpTraceback(getFrame func(i int) (int, string, string), tracebackLen int) {
	n.dumpType(pb.Variant_VARIANT_TRACEBACK)

	for i := 0; i < tracebackLen; i++ {
		lineno, filename, functionName := getFrame(i)
		n.CodeValues = append(n.CodeValues, &pb.Variant_CodeObject{
			Lineno:               uint32(lineno),
			FilenameIndexInCache: n.getStringIndexInCache(filename),
			NameIndexInCache:     n.getStringIndexInCache(functionName),
			ModuleIndexInCache:   n.getStringIndexInCache(filename),
		})
	}
}

func (n *NamespaceSerializer2) dumpStruct(getField func(i int) (string, Namespace), numOfFields int, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_OBJECT)

	if n.getCurrentDepth()+1 >= config.MaxDepth {
		n.dumpMaxDepth(true)
		return
	}

	for i := 0; i < numOfFields; i++ {
		fieldName, fieldValue := getField(i)
		n.AttributeNamesInCache = append(n.AttributeNamesInCache, n.getStringIndexInCache(fieldName))
		n.AttributeValues = append(n.AttributeValues, n.spawn(fieldValue).Variant2)
	}
}

func (n *NamespaceSerializer2) dumpMap(getKeyValue func(i int) (Namespace, Namespace), mapLen int, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_MAP)
	n.OriginalSize = uint32(mapLen)

	if n.getCurrentDepth() >= config.MaxCollectionDepth {
		n.dumpMaxDepth(true)
		return
	}

	for i := 0; i < mapLen; i++ {
		if i >= config.MaxWidth {
			return
		}

		key, value := getKeyValue(i)
		n.CollectionKeys = append(n.CollectionKeys, n.spawn(key).Variant2)
		n.CollectionValues = append(n.CollectionValues, n.spawn(value).Variant2)
	}
}

func (n *NamespaceSerializer2) dumpNil() {
	n.dumpType(pb.Variant_VARIANT_NONE)
}

func (n *NamespaceSerializer2) dumpFunc(functionName string, filename string, lineno int) {
	n.dumpType(pb.Variant_VARIANT_CODE_OBJECT)
	n.CodeValues = append(n.CodeValues, &pb.Variant_CodeObject{
		NameIndexInCache:     n.getStringIndexInCache(functionName),
		ModuleIndexInCache:   n.getStringIndexInCache(filename),
		Lineno:               uint32(lineno),
		FilenameIndexInCache: n.getStringIndexInCache(filename),
	})
}

func (n *NamespaceSerializer2) dumpChan(value reflect.Value, config config.ObjectDumpConfig) {
	n.dumpType(pb.Variant_VARIANT_LIST)

	if n.getCurrentDepth() >= config.MaxCollectionDepth {
		n.dumpMaxDepth(true)
		return
	}

	addr := utils.UnsafePointer(value)
	chanStruct := *(*hchan)(addr)
	elemType := value.Type().Elem()
	bufSize := value.Len() * int(elemType.Size())

	for i := 0; i < bufSize; i += int(elemType.Size()) {
		if len(n.CollectionValues) >= config.MaxWidth {
			return
		}

		ptr := unsafe.Pointer(uintptr(chanStruct.buf) + uintptr(i))
		val := reflect.NewAt(elemType, ptr).Elem()
		valVariant := &NamespaceSerializer2{StringCache: n.StringCache, Variant2: &pb.Variant2{}, logErrors: n.logErrors, currentDepth: n.currentDepth + 1}
		dumpValue(valVariant, val, config)
		n.CollectionValues = append(n.CollectionValues, valVariant.Variant2)
	}
}

func (n *NamespaceSerializer2) dumpRookoutError(r rookoutErrors.RookoutError) {
	n.dumpType(pb.Variant_VARIANT_ERROR)

	parameters := n.spawn(NewGoObjectNamespace(r.GetArguments()))
	traceback := n.spawn(NewGoObjectNamespace(r.StackFrames()))
	n.ErrorValue = &pb.Error2{
		Message:    r.Error(),
		Type:       r.GetType(),
		Parameters: parameters.Variant2,
		Traceback:  traceback.Variant2,
	}
}

func (n *NamespaceSerializer2) dumpErrorMessage(msg string) {
	n.dumpType(pb.Variant_VARIANT_ERROR)
	n.ErrorValue = &pb.Error2{
		Message: msg,
	}
}

func (n *NamespaceSerializer2) dumpEnum(desc string, ordinal int, _ string) {
	n.dumpType(pb.Variant_VARIANT_ENUM)
	n.BytesIndexInCache = n.getStringIndexInCache(desc)
	n.LongValue = int64(ordinal)
}
