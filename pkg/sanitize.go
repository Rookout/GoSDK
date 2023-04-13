package pkg

import (
	"reflect"
	"strings"

	"github.com/Rookout/GoSDK/pkg/config"
)

var sanitizeBlacklist = map[string]struct{}{"Labels": {}}

func isBlacklisted(str string) bool {
	_, ok := sanitizeBlacklist[str]
	return ok
}

func Sanitize(obj *config.RookOptions) {
	fields := reflect.TypeOf(*obj)
	value := reflect.ValueOf(obj)
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		if field.Type.Name() == "string" && value.Elem().Field(i).CanSet() && !isBlacklisted(field.Name) {
			value.Elem().Field(i).SetString(sanitizeString(field.Name, value.Elem().Field(i).String()))
		}
	}
}

func sanitizeString(name, value string) string {
	if isBlacklisted(name) {
		return value
	}

	return strings.Trim(value, " \r\n\t")
}
