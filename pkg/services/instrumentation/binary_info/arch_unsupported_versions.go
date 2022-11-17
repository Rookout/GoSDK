//go:build !amd64
// +build !amd64

package binary_info

import (
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/frame"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)

func FixFrameUnwindContext(_ *frame.FrameContext, _ uint64, _ *BinaryInfo) *frame.FrameContext {
	return nil
}

func RegSize(_ uint64) int {
	return 0
}

func RegistersToDwarfRegisters(_ uint64, _ registers.Registers) op.DwarfRegisters {
	return op.DwarfRegisters{}
}
