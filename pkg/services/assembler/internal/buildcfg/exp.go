// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package buildcfg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/goexperiment"
)



type ExperimentFlags struct {
	goexperiment.Flags
	baseline goexperiment.Flags
}










var Experiment ExperimentFlags = func() ExperimentFlags {
	flags, err := ParseGOEXPERIMENT(GOOS, GOARCH, envOr("GOEXPERIMENT", ""))
	if err != nil {
		Error = err
		return ExperimentFlags{}
	}
	return *flags
}()



const DefaultGOEXPERIMENT = ""








var FramePointerEnabled = GOARCH == "amd64" || GOARCH == "arm64"






func ParseGOEXPERIMENT(goos, goarch, goexp string) (*ExperimentFlags, error) {
	
	
	
	
	var regabiSupported, regabiAlwaysOn bool
	switch goarch {
	case "amd64", "arm64", "ppc64le", "ppc64", "riscv64":
		regabiAlwaysOn = true
		regabiSupported = true
	}

	baseline := goexperiment.Flags{
		RegabiWrappers:   regabiSupported,
		RegabiArgs:       regabiSupported,
		CoverageRedesign: true,
	}

	
	flags := &ExperimentFlags{
		Flags:    baseline,
		baseline: baseline,
	}

	
	
	
	if goexp != "" {
		
		names := make(map[string]func(bool))
		rv := reflect.ValueOf(&flags.Flags).Elem()
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rv.Field(i)
			names[strings.ToLower(rt.Field(i).Name)] = field.SetBool
		}

		
		
		
		
		names["regabi"] = func(v bool) {
			flags.RegabiWrappers = v
			flags.RegabiArgs = v
		}

		
		for _, f := range strings.Split(goexp, ",") {
			if f == "" {
				continue
			}
			if f == "none" {
				
				
				
				flags.Flags = goexperiment.Flags{}
				continue
			}
			val := true
			if strings.HasPrefix(f, "no") {
				f, val = f[2:], false
			}
			set, ok := names[f]
			if !ok {
				return nil, fmt.Errorf("unknown GOEXPERIMENT %s", f)
			}
			set(val)
		}
	}

	if regabiAlwaysOn {
		flags.RegabiWrappers = true
		flags.RegabiArgs = true
	}
	
	if !regabiSupported {
		flags.RegabiWrappers = false
		flags.RegabiArgs = false
	}
	
	if flags.RegabiArgs && !flags.RegabiWrappers {
		return nil, fmt.Errorf("GOEXPERIMENT regabiargs requires regabiwrappers")
	}
	return flags, nil
}



func (exp *ExperimentFlags) String() string {
	return strings.Join(expList(&exp.Flags, &exp.baseline, false), ",")
}





func expList(exp, base *goexperiment.Flags, all bool) []string {
	var list []string
	rv := reflect.ValueOf(exp).Elem()
	var rBase reflect.Value
	if base != nil {
		rBase = reflect.ValueOf(base).Elem()
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		name := strings.ToLower(rt.Field(i).Name)
		val := rv.Field(i).Bool()
		baseVal := false
		if base != nil {
			baseVal = rBase.Field(i).Bool()
		}
		if all || val != baseVal {
			if val {
				list = append(list, name)
			} else {
				list = append(list, "no"+name)
			}
		}
	}
	return list
}



func (exp *ExperimentFlags) Enabled() []string {
	return expList(&exp.Flags, nil, false)
}



func (exp *ExperimentFlags) All() []string {
	return expList(&exp.Flags, nil, true)
}
