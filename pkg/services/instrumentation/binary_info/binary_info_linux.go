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

//go:build linux
// +build linux

package binary_info

import (
	"debug/elf"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/utils"
)

type File = elf.File
type archID = elf.Machine


var supportedArchs = map[elf.Machine]interface{}{
	elf.EM_X86_64:  nil,
	elf.EM_AARCH64: nil,
	elf.EM_386:     nil,
}

const crosscall2SPOffset = 0x58

func loadBinaryInfo(bi *BinaryInfo, image *Image, path string, addr uint64) error {
	exe, err := os.OpenFile(path, 0, os.ModePerm)
	if err != nil {
		return err
	}
	image.closer = exe
	elfFile, err := elf.NewFile(exe)
	if err != nil {
		return err
	}
	if !isSupportedArch(elfFile.Machine) {
		return errors.New("unsupported linux arch")
	}

	if image.Index == 0 {
		
		
		
		
		
		if addr != 0 {
			image.StaticBase = addr - elfFile.Entry
		} else if elfFile.Type == elf.ET_DYN {
			return errors.New("could not determine base address of a PIE")
		}
		if dynsec := elfFile.Section(".dynamic"); dynsec != nil {
			bi.ElfDynamicSection.Addr = dynsec.Addr + image.StaticBase
			bi.ElfDynamicSection.Size = dynsec.Size
		}
	} else {
		image.StaticBase = addr
	}

	dwarfFile := elfFile

	var debugInfoBytes []byte
	image.Dwarf, err = elfFile.DWARF()
	if err != nil {
		var sepFile *os.File
		var serr error
		sepFile, dwarfFile, serr = bi.openSeparateDebugInfo(image, elfFile, bi.debugInfoDirectories)
		if serr != nil {
			return serr
		}
		image.sepDebugCloser = sepFile
		image.Dwarf, err = dwarfFile.DWARF()
		if err != nil {
			return err
		}
	}

	debugInfoBytes, err = GetDebugSection(dwarfFile, "info")
	if err != nil {
		return err
	}

	debugLineBytes, err := GetDebugSection(dwarfFile, "line")
	if err != nil {
		return err
	}
	bi.debugLocBytes, _ = GetDebugSection(dwarfFile, "loc")
	bi.debugLoclistBytes, _ = GetDebugSection(dwarfFile, "loclists")
	debugAddrBytes, _ := GetDebugSection(dwarfFile, "addr")
	image.debugAddr = godwarf.ParseAddr(debugAddrBytes)
	debugLineStrBytes, _ := GetDebugSection(dwarfFile, "line_str")
	image.debugLineStr = debugLineStrBytes

	wg := &sync.WaitGroup{}
	wg.Add(2)
	utils.CreateGoroutine(func() {
		defer wg.Done()
		err = bi.parseDebugFrame(image, dwarfFile, debugInfoBytes)
		if err != nil {
			logger.Logger().WithError(err).Error("Failed to parse debug frame")
		}
	})
	utils.CreateGoroutine(func() {
		defer wg.Done()
		err = bi.loadDebugInfoMaps(image, debugInfoBytes, debugLineBytes)
		if err != nil {
			logger.Logger().WithError(err).Error("Failed to load debug info maps")
		}
	})
	wg.Wait()
	return nil
}

func getSectionName(section string) string {
	return ".debug_" + section
}

func getCompressedSectionName(section string) string {
	return ".zdebug_" + section
}









func (bi *BinaryInfo) openSeparateDebugInfo(image *Image, exe *File, debugInfoDirectories []string) (*os.File, *File, error) {
	var debugFilePath string
	for _, dir := range debugInfoDirectories {
		var potentialDebugFilePath string
		if strings.Contains(dir, "build-id") {
			desc1, desc2, err := parseBuildID(exe)
			if err != nil {
				continue
			}
			potentialDebugFilePath = fmt.Sprintf("%s/%s/%s.debug", dir, desc1, desc2)
		} else if strings.HasPrefix(image.Path, "/proc") {
			path, err := filepath.EvalSymlinks(image.Path)
			if err == nil {
				potentialDebugFilePath = fmt.Sprintf("%s/%s.debug", dir, filepath.Base(path))
			}
		} else {
			potentialDebugFilePath = fmt.Sprintf("%s/%s.debug", dir, filepath.Base(image.Path))
		}
		_, err := os.Stat(potentialDebugFilePath)
		if err == nil {
			debugFilePath = potentialDebugFilePath
			break
		}
	}
	if debugFilePath == "" {
		return nil, nil, errors.New("no debug info found")
	}
	sepFile, err := os.OpenFile(debugFilePath, 0, os.ModePerm)
	if err != nil {
		return nil, nil, errors.New("can't open separate debug file: " + err.Error())
	}

	elfFile, err := elf.NewFile(sepFile)
	if err != nil {
		sepFile.Close()
		return nil, nil, fmt.Errorf("can't open separate debug file %q: %v", debugFilePath, err.Error())
	}

	if !isSupportedArch(elfFile.Machine) {
		sepFile.Close()
		return nil, nil, fmt.Errorf("can't open separate debug file %q", debugFilePath)
	}

	return sepFile, elfFile, nil
}

func parseBuildID(exe *elf.File) (string, string, error) {
	buildid := exe.Section(".note.gnu.build-id")
	if buildid == nil {
		return "", "", errors.New("no build id")
	}

	br := buildid.Open()
	bh := new(buildIDHeader)
	if err := binary.Read(br, binary.LittleEndian, bh); err != nil {
		return "", "", errors.New("can't read build-id header: " + err.Error())
	}

	name := make([]byte, bh.Namesz)
	if err := binary.Read(br, binary.LittleEndian, name); err != nil {
		return "", "", errors.New("can't read build-id name: " + err.Error())
	}

	if strings.TrimSpace(string(name)) != "GNU\x00" {
		return "", "", errors.New("invalid build-id signature")
	}

	descBinary := make([]byte, bh.Descsz)
	if err := binary.Read(br, binary.LittleEndian, descBinary); err != nil {
		return "", "", errors.New("can't read build-id desc: " + err.Error())
	}
	desc := hex.EncodeToString(descBinary)
	return desc[:2], desc[2:], nil
}

func getEhFrameSection(f *elf.File) *elf.Section {
	return f.Section(".eh_frame")
}
