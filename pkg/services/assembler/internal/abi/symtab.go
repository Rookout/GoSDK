// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package abi


type FuncFlag uint8

const (
	
	
	
	
	
	
	FuncFlagTopFrame FuncFlag = 1 << iota

	
	
	
	
	
	
	
	FuncFlagSPWrite

	
	FuncFlagAsm
)





type FuncID uint8

const (
	
	

	FuncIDNormal FuncID = iota 
	FuncID_abort
	FuncID_asmcgocall
	FuncID_asyncPreempt
	FuncID_cgocallback
	FuncID_debugCallV2
	FuncID_gcBgMarkWorker
	FuncID_goexit
	FuncID_gogo
	FuncID_gopanic
	FuncID_handleAsyncEvent
	FuncID_mcall
	FuncID_morestack
	FuncID_mstart
	FuncID_panicwrap
	FuncID_rt0_go
	FuncID_runfinq
	FuncID_runtime_main
	FuncID_sigpanic
	FuncID_systemstack
	FuncID_systemstack_switch
	FuncIDWrapper 
)





const ArgsSizeUnknown = -0x80000000




const (
	PCDATA_UnsafePoint   = 0
	PCDATA_StackMapIndex = 1
	PCDATA_InlTreeIndex  = 2
	PCDATA_ArgLiveIndex  = 3

	FUNCDATA_ArgsPointerMaps    = 0
	FUNCDATA_LocalsPointerMaps  = 1
	FUNCDATA_StackObjects       = 2
	FUNCDATA_InlTree            = 3
	FUNCDATA_OpenCodedDeferInfo = 4
	FUNCDATA_ArgInfo            = 5
	FUNCDATA_ArgLiveInfo        = 6
	FUNCDATA_WrapInfo           = 7
)


const (
	UnsafePointSafe   = -1 
	UnsafePointUnsafe = -2 

	
	
	
	
	
	UnsafePointRestart1 = -3
	UnsafePointRestart2 = -4

	
	UnsafePointRestartAtEntry = -5
)
