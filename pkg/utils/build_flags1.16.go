//go:build go1.16 && !go1.18
// +build go1.16,!go1.18

package utils

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"runtime/debug"
)

func GetBuildOpts() (map[string]string, *debug.BuildInfo, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, nil, rookoutErrors.NewReadBuildFlagsError()
	}

	return nil, info, nil
}

func ValidateBuildOpts(opts map[string]string) error {
	return nil
}
