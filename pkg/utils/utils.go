package utils

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
)

const backendCompatibleIsoTimeFormat = "2006-01-02T15:04:05.000000Z"

var TrueValues = []string{"y", "yes", "true", "1"}

func BoolAsInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IntAsBool(i int32) bool {
	return !(0 == i)
}

func CreateByteArray(size int) interface{} {
	array := make([]byte, size)
	for i := range array {
		array[i] = byte(i + 60)
	}
	return array
}

func StringToFloat(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func StringToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

func StringToBool(str string) (bool, error) {
	return strconv.ParseBool(str)
}

func BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

func MapStringToMapInterface(orig *map[string]interface{}) *map[interface{}]interface{} {
	retVal := make(map[interface{}]interface{})
	for key, val := range *orig {
		retVal[key] = val
	}
	return &retVal
}

func InterfaceToMap(i interface{}) *map[interface{}]interface{} {
	if m, ok := i.(map[interface{}]interface{}); ok {
		return &m
	}

	retVal := make(map[interface{}]interface{})

	for _, key := range reflect.ValueOf(i).MapKeys() {
		val := reflect.ValueOf(i).MapIndex(key)

		retVal[key.Interface()] = val.Interface()
	}

	return &retVal
}

func GetFormattedTime() string {
	return fmt.Sprintf(time.Now().Format("2006/01/02 15:04:05"))
}

func ReplaceAll(original string, old string, new string) string {
	return strings.Replace(original, old, new, -1)
}

func GetFloat(o interface{}) (float64, bool) {
	if o == nil {
		return 0, false
	}

	switch o.(type) {
	case float64:
		return o.(float64), true
	case int:
		return float64(o.(int)), true
	case int32:
		return float64(o.(int32)), true
	case int64:
		return float64(o.(int64)), true
	default:
		return 0, false
	}
}

func MergeUrls(baseUrl, postfix string) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, postfix)
	return u.String(), nil
}

func CreateDirectory(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func TimeToBackendCompatibleIsoTimeFormat(t time.Time) string {
	return t.Format(backendCompatibleIsoTimeFormat)
}

func TimeFromBackendCompatibleIsoTimeFormat(t string) time.Time {
	ret, err := time.Parse(backendCompatibleIsoTimeFormat, t)
	if err != nil {
		logrus.WithError(err).Error("unexpected error while converting backend time to go Time")
	}
	return ret
}

func ProtoToBackendCompatibleIsoTimeFormat(t *timestamp.Timestamp) string {
	protoAsTime := time.Unix(t.Seconds, (int64)(t.Nanos))
	return TimeToBackendCompatibleIsoTimeFormat(protoAsTime)
}

func TimeToProtobufTimestamp(t *time.Time) *timestamp.Timestamp {
	if t != nil {
		out, _ := ptypes.TimestampProto(*t)
		return out
	}
	return nil
}


func StrSliceDiff(base, other []string) (diff, equal []string) {
	otherAsMap := make(map[string]struct{}, len(other))
	for _, value := range other {
		otherAsMap[value] = struct{}{}
	}
	for _, value := range base {
		if _, found := otherAsMap[value]; !found {
			diff = append(diff, value)
		} else {
			equal = append(equal, value)
		}
	}
	return diff, equal
}

func MSToNS(milliseconds int64) time.Duration {
	return time.Duration(milliseconds * 1000000)
}

func StringMSToNS(milliseconds string) (time.Duration, error) {
	i, err := strconv.Atoi(milliseconds)
	if err != nil {
		return -1, err
	}

	return MSToNS(int64(i)), nil
}

func FloatMSToNS(milliseconds float64) int64 {
	return int64(milliseconds) * 1000000
}

func NowInNS() int64 {
	return time.Now().UnixNano()
}

func Int64Max(a int64, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func UnsafePointer(value reflect.Value) unsafe.Pointer {
	return unsafe.Pointer(value.Pointer())
}
