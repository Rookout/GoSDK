package callback

import _ "unsafe"



func PrepForCallback()

//go:linkname MoreStack runtime.morestack
func MoreStack()

func ShouldRunPrologue()
