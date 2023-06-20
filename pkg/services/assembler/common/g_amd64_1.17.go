//go:build go1.17 && amd64
// +build go1.17,amd64

package common

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"golang.org/x/arch/x86/x86asm"
)

var InitError error
var g = x86asm.R14


func MovGToR12(b *assembler.Builder) *assembler.Instruction {
	return b.Inst(assembler.AMOVQ, x86asm.R12, g)
}
