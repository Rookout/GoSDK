package config

import (
	"reflect"
	"sync/atomic"
	"unsafe"
)

var defaults = &ObjectDumpConfigDefaults{}

func UpdateObjectDumpConfigDefaults(config ObjectDumpConfigDefaults) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&defaults)), unsafe.Pointer(&config))
}

func GetObjectDumpConfig(key string) (ObjectDumpConfig, bool) {
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
	return defaults.defaultConfig
}

type ObjectDumpConfig struct {
	MaxDepth           int
	MaxWidth           int
	MaxCollectionDepth int
	MaxString          int
	ShouldTailor       bool
	IsTailored         bool
}

func (o *ObjectDumpConfig) Tailor(kind reflect.Kind) {
	o.IsTailored = true
	o.ShouldTailor = false

	switch kind {
	case reflect.String:
		o.MaxString = defaults.unlimitedConfig.MaxString
		o.MaxDepth = 1
	case reflect.Array, reflect.Slice, reflect.Map:
		o.MaxDepth = defaults.defaultConfig.MaxDepth
		o.MaxWidth = defaults.unlimitedConfig.MaxWidth
		o.MaxCollectionDepth = defaults.defaultConfig.MaxCollectionDepth
		o.MaxString = defaults.defaultConfig.MaxString
	default:
		*o = defaults.tolerantConfig
	}
}

func GetTailoredLimits(obj interface{}) (config ObjectDumpConfig) {
	config.IsTailored = true

	if obj == nil {
		config = defaults.tolerantConfig
		return
	}

	kind := reflect.TypeOf(obj).Kind()
	config.Tailor(kind)
	return
}
