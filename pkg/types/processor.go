package types

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"regexp"
)

type FieldFilterType int

type FieldFilter struct {
	FilterType FieldFilterType
	Pattern    *regexp.Regexp
	Whitelist  bool
}

type Namespace interface {
	CallMethod(name string, args string) (Namespace, rookoutErrors.RookoutError)

	WriteAttribute(name string, value Namespace) rookoutErrors.RookoutError
	ReadAttribute(name string) (Namespace, rookoutErrors.RookoutError)

	ReadKey(key interface{}) (Namespace, rookoutErrors.RookoutError)
	GetObject() interface{}
	ToProtobuf(logErrors bool) *pb.Variant
	ToDict() map[string]interface{}
	ToSimpleDict() interface{}
	Filter(filters []FieldFilter) rookoutErrors.RookoutError
}
