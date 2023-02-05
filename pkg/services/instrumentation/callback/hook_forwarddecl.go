package callback

import _ "unsafe"



//go:linkname MoreStack runtime.morestack
func MoreStack()

func ShouldRunPrologue()

func getContext() uintptr
