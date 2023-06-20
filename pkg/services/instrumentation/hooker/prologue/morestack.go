package prologue

import (
	_ "unsafe"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
)

var morestackAddrs = make(map[uintptr]struct{})
var morestackAddr uintptr

//go:linkname morestack runtime.morestack
func morestack()

//go:linkname morestack_noctxt runtime.morestack_noctxt
func morestack_noctxt()

func addMorestackFunc(binaryInfo *binary_info.BinaryInfo, morestack func()) {
	addr, err := binaryInfo.GetUnwrappedFuncPointer(morestack)
	if err != nil {
		logger.Logger().Warningf("Error while trying to get unwrapped morestack pointer: %v", err)
		return
	}

	morestackAddrs[addr] = struct{}{}
}

func Init(binaryInfo *binary_info.BinaryInfo) rookoutErrors.RookoutError {
	var err rookoutErrors.RookoutError
	morestackAddr, err = binaryInfo.GetUnwrappedFuncPointer(morestack)
	if err != nil {
		return err
	}
	addMorestackFunc(binaryInfo, morestack)
	addMorestackFunc(binaryInfo, morestack_noctxt)
	return nil
}
