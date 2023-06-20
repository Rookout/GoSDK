//go:build go1.15 && !go1.17 && amd64
// +build go1.15,!go1.17,amd64

package common

import (
	"reflect"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/disassembler"
	"github.com/Rookout/GoSDK/pkg/utils"
	"golang.org/x/arch/x86/x86asm"
)

func movGToR12()

var InitError error
var movGToR12Bytes []byte


func init() {
	
	entry := reflect.ValueOf(movGToR12).Pointer()
	movGInst, err := disassembler.DecodeOne(entry)
	if err != nil {
		InitError = err
		return
	}
	
	retInst, err := disassembler.DecodeOne(entry + uintptr(movGInst.Len))
	if err != nil {
		InitError = err
		return
	}
	if retInst.Op != x86asm.RET {
		InitError = rookoutErrors.NewUnexpectedInstruction(movGInst, retInst)
		return
	}

	movGToR12Bytes = utils.MakeSliceFromPointer(entry, movGInst.Len)
}


func MovGToR12(b *assembler.Builder) *assembler.Instruction {
	return b.Bytes(movGToR12Bytes)
}
