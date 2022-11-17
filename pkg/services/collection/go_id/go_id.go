package go_id

import (
	"github.com/Rookout/GoSDK/pkg/services/go_runtime"
	"unsafe"
)

//go:nosplit
func CurrentGoID() int {
	g := go_runtime.Getg()
	goidInG := g + goidOffsetInG
	goid := *(*int64)(unsafe.Pointer(goidInG)) // goid is int64
	return int(goid)
}
