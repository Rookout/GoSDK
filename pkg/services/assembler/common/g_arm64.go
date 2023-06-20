//go:build arm64
// +build arm64

package common

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"golang.org/x/arch/arm64/arm64asm"
)

var InitError rookoutErrors.RookoutError
var g = arm64asm.X28

func MovGToX20(b *assembler.Builder) *assembler.Instruction {
	return b.Inst(assembler.AMOVD, arm64asm.X20, g)
}
