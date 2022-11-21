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
