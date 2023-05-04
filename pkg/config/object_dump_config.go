package config

import (
	"reflect"
)

func GetObjectDumpConfig(key string) (ObjectDumpConfig, bool) {
	defaults := config.Load().(DynamicConfiguration).ObjectDumpConfigDefaults
	switch key {
	case "strict":
		return defaults.strictConfig, true
	case "default":
		return defaults.defaultConfig, true
	case "tolerant":
		return defaults.tolerantConfig, true
	default:
		return ObjectDumpConfig{}, false
	}
}

func GetDefaultDumpConfig() ObjectDumpConfig {
	c := config.Load().(DynamicConfiguration)
	return c.ObjectDumpConfigDefaults.defaultConfig
}

type ObjectDumpConfig struct {
	MaxDepth           int
	MaxWidth           int
	MaxCollectionDepth int
	MaxString          int
	ShouldTailor       bool
	IsTailored         bool
}

func (o *ObjectDumpConfig) Tailor(kind reflect.Kind, objLen int) {
	defer func() {
		o.IsTailored = true
		o.ShouldTailor = false
	}()

	defaults := config.Load().(DynamicConfiguration).ObjectDumpConfigDefaults
	switch kind {
	case reflect.String:
		o.MaxString = defaults.unlimitedConfig.MaxString
		o.MaxDepth = 1
		return
	case reflect.Array, reflect.Slice, reflect.Map:
		if objLen > defaults.tolerantConfig.MaxWidth {
			o.MaxDepth = defaults.defaultConfig.MaxDepth
			o.MaxWidth = defaults.unlimitedConfig.MaxWidth
			o.MaxCollectionDepth = defaults.defaultConfig.MaxCollectionDepth
			o.MaxString = defaults.defaultConfig.MaxString
			return
		}
	}
	*o = defaults.tolerantConfig
}

func GetTailoredLimits(obj interface{}) ObjectDumpConfig {
	c := ObjectDumpConfig{
		IsTailored: true,
	}

	defaults := config.Load().(DynamicConfiguration).ObjectDumpConfigDefaults
	if obj == nil {
		return defaults.tolerantConfig
	}

	value := reflect.ValueOf(obj)
	objLen := 0
	if value.Kind() == reflect.Array ||
		value.Kind() == reflect.Slice ||
		value.Kind() == reflect.Map {
		objLen = value.Len()
	}
	c.Tailor(value.Kind(), objLen)
	return c
}
