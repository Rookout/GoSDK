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

//go:build windows
// +build windows

package binary_info

import (
	"debug/pe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

var supportedArchs = map[archID]bool{}

const crosscall2SPOffset = 0x118

type File = pe.File
type archID = string
type section struct {
	pe.Section
	Addr uint64
}

func loadBinaryInfo(_ *BinaryInfo, _ *Image, _ string, _ uint64) error {
	return rookoutErrors.NewUnsupportedPlatform()
}

func getSectionName(_ string) string {
	return ""
}

func getCompressedSectionName(_ string) string {
	return ""
}

func getEhFrameSection(_ *pe.File) *section {
	return nil
}
