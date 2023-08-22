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

func TailorObjectDumpConfig(kind reflect.Kind, objLen int) (o ObjectDumpConfig) {
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
		if objLen > defaults.tolerantConfig.MaxWidth || objLen == 0 {
			o.MaxDepth = defaults.defaultConfig.MaxDepth
			o.MaxWidth = defaults.unlimitedConfig.MaxWidth
			o.MaxCollectionDepth = defaults.defaultConfig.MaxCollectionDepth
			o.MaxString = defaults.defaultConfig.MaxString
			return
		}
	}
	return defaults.tolerantConfig
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxObjectDumpConfig(a, b ObjectDumpConfig) ObjectDumpConfig {
	return ObjectDumpConfig{
		MaxDepth:           max(a.MaxDepth, b.MaxDepth),
		MaxCollectionDepth: max(a.MaxCollectionDepth, b.MaxCollectionDepth),
		MaxWidth:           max(a.MaxWidth, b.MaxWidth),
		MaxString:          max(a.MaxString, b.MaxString),
		ShouldTailor:       a.ShouldTailor || b.ShouldTailor,
		IsTailored:         false,
	}
}
