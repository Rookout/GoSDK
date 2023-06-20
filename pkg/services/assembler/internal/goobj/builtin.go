// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goobj

import "github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"







func NBuiltin() int {
	return len(builtins)
}



func BuiltinName(i int) (string, int) {
	return builtins[i].name, builtins[i].abi
}



func BuiltinIdx(name string, abi int) int {
	i, ok := builtinMap[name]
	if !ok {
		return -1
	}
	if buildcfg.Experiment.RegabiWrappers && builtins[i].abi != abi {
		return -1
	}
	return i
}

//go:generate go run mkbuiltin.go

var builtinMap map[string]int

func init() {
	builtinMap = make(map[string]int, len(builtins))
	for i, b := range builtins {
		builtinMap[b.name] = i
	}
}
