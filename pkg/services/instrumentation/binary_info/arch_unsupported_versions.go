// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

//go:build !amd64 && !arm64
// +build !amd64,!arm64

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
