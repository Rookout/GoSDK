// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package objabi

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/abi"
	"strings"
)

var funcIDs = map[string]abi.FuncID{
	"abort":              abi.FuncID_abort,
	"asmcgocall":         abi.FuncID_asmcgocall,
	"asyncPreempt":       abi.FuncID_asyncPreempt,
	"cgocallback":        abi.FuncID_cgocallback,
	"debugCallV2":        abi.FuncID_debugCallV2,
	"gcBgMarkWorker":     abi.FuncID_gcBgMarkWorker,
	"rt0_go":             abi.FuncID_rt0_go,
	"goexit":             abi.FuncID_goexit,
	"gogo":               abi.FuncID_gogo,
	"gopanic":            abi.FuncID_gopanic,
	"handleAsyncEvent":   abi.FuncID_handleAsyncEvent,
	"main":               abi.FuncID_runtime_main,
	"mcall":              abi.FuncID_mcall,
	"morestack":          abi.FuncID_morestack,
	"mstart":             abi.FuncID_mstart,
	"panicwrap":          abi.FuncID_panicwrap,
	"runfinq":            abi.FuncID_runfinq,
	"sigpanic":           abi.FuncID_sigpanic,
	"systemstack_switch": abi.FuncID_systemstack_switch,
	"systemstack":        abi.FuncID_systemstack,

	
	"deferreturn":       abi.FuncIDWrapper,
	"runOpenDeferFrame": abi.FuncIDWrapper,
	"deferCallSave":     abi.FuncIDWrapper,
}



func GetFuncID(name string, isWrapper bool) abi.FuncID {
	if isWrapper {
		return abi.FuncIDWrapper
	}
	if strings.HasPrefix(name, "runtime.") {
		if id, ok := funcIDs[name[len("runtime."):]]; ok {
			return id
		}
	}
	return abi.FuncIDNormal
}
