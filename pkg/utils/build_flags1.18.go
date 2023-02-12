//go:build go1.18
// +build go1.18

package utils

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	gcflags = "-gcflags"

	ldflags = "-ldflags"
)

type packageType int

const (
	packageTypeNone packageType = iota
	packageTypeAll
	packageTypeElse
)

func packageNameToType(name string) packageType {
	switch strings.TrimSpace(name) {
	case "":
		return packageTypeNone
	case "all":
		return packageTypeAll
	default:
		return packageTypeElse
	}
}

type flagMatcher struct {
	flag         string                      
	validPackage map[packageType]interface{} 
	defaultOk    bool                        
	expectedVal  bool                        
}

type flagRes struct {
	flag      string
	pType     packageType
	isDefault bool
	val       bool
}

var (
	legalFlags = map[string]map[string]flagMatcher{
		gcflags: {
			"-N":                  {"-N", map[packageType]interface{}{packageTypeAll: nil}, true, true},                  
			"-dwarflocationlists": {"-dwarflocationlists", map[packageType]interface{}{packageTypeAll: nil}, true, true}, 
		},
	}

	illegalFlags = map[string]map[string]flagMatcher{
		ldflags: {
			"-s": {"-s", map[packageType]interface{}{packageTypeAll: nil}, false, false}, 
			"-w": {"-w", map[packageType]interface{}{packageTypeAll: nil}, false, false}, 
		},
	}

	parseFlagsCategories = map[string]interface{}{
		gcflags: nil,
		ldflags: nil,
	}
)

func parseFlag(prevPackageType packageType, packageName *string, flag string, val *string) *flagRes {
	isFlag := func(f string) bool {
		return strings.HasPrefix(f, "-")
	}
	isPackage := func(p string) bool {
		return !isFlag(p)
	}
	currentPackageType := prevPackageType
	if packageName != nil {
		if !isPackage(*packageName) {
			return nil
		}
		currentPackageType = packageNameToType(*packageName)
	}
	if !isFlag(flag) {
		return nil
	}
	if val == nil {
		return &flagRes{flag: flag, pType: currentPackageType, isDefault: true, val: false}
	}

	boolVal, err := strconv.ParseBool(*val)
	if err != nil {
		return nil
	}
	return &flagRes{flag: flag, pType: currentPackageType, isDefault: false, val: boolVal}
}

func flagStr2FlagRes(flag string, prevPackageType packageType) (*flagRes, error) {
	splitted := strings.Split(flag, "=")
	var parsed *flagRes = nil
	switch len(splitted) {
	case 1: 
		parsed = parseFlag(prevPackageType, nil, splitted[0], nil)
	case 2: 
		parsed = parseFlag(prevPackageType, &splitted[0], splitted[1], nil)
		if parsed == nil {
			parsed = parseFlag(prevPackageType, nil, splitted[0], &splitted[1])
		}
	case 3: 
		parsed = parseFlag(prevPackageType, &splitted[0], splitted[1], &splitted[2])
	}
	if parsed == nil {
		return nil, fmt.Errorf("flag must be of the form [<PACKAGE>=]-<FLAG>[=<BOOL_VAL>] given %s", flag)
	}
	return parsed, nil
}

func parseFlagsLine(line string) ([]flagRes, error) {
	prevPackageType := packageTypeNone
	vals := strings.Fields(line)
	res := make([]flagRes, 0, len(vals))
	for _, v := range vals {
		parsed, err := flagStr2FlagRes(v, prevPackageType)
		if err != nil {
			return nil, err
		}
		res = append(res, *parsed)
		prevPackageType = parsed.pType
	}
	return res, nil
}

func (m flagMatcher) isFlagMatching(flag flagRes) bool {
	if _, ok := m.validPackage[flag.pType]; !ok {
		return false
	}
	if flag.isDefault {
		return m.defaultOk
	}
	return flag.val == m.expectedVal
}

func validateTrueVals(opts map[string][]flagRes) error {
	for requiredFlag, flagSpec := range legalFlags {
		flags, found := opts[requiredFlag]
		if !found {
			return fmt.Errorf("didn't provide required flag %s", requiredFlag)
		}
		matched := false
		for _, flag := range flags {
			if matcher, ok := flagSpec[flag.flag]; ok {
				if matcher.isFlagMatching(flag) {
					matched = true
					break
				}
			}
		}
		if !matched {
			return fmt.Errorf("didn't provide any of the required flags for %s", requiredFlag)
		}
	}
	return nil
}

func validateFalseVals(opts map[string][]flagRes) error {
	for optionalFlag, flagSpec := range illegalFlags {
		flags, found := opts[optionalFlag]
		if !found {
			continue
		}
		for _, flag := range flags {
			if matcher, ok := flagSpec[flag.flag]; ok {
				if !matcher.isFlagMatching(flag) {
					return fmt.Errorf("provided illegal value %s for %s", flag.flag, optionalFlag)
				}
			}
		}
	}
	return nil
}

func GetBuildOpts() (map[string]string, *debug.BuildInfo, error) {
	opts := make(map[string]string)
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, nil, rookoutErrors.NewReadBuildFlagsError()
	}
	for _, setting := range info.Settings {
		opts[setting.Key] = setting.Value
	}
	return opts, info, nil
}

func ValidateBuildOpts(opts map[string]string) error {
	parsed := make(map[string][]flagRes)
	for optName := range opts {
		if _, ok := parseFlagsCategories[optName]; ok {
			parsedFlags, err := parseFlagsLine(opts[optName])
			if err != nil {
				return err
			}
			parsed[optName] = parsedFlags
		}
	}

	if err := validateTrueVals(parsed); err != nil {
		return rookoutErrors.NewValidateBuildFlagsError(err)
	}
	if err := validateFalseVals(parsed); err != nil {
		return rookoutErrors.NewValidateBuildFlagsError(err)
	}
	return nil
}
