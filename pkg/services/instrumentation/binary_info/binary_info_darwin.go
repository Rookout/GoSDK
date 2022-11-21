//go:build darwin
// +build darwin

package binary_info

import (
	"debug/macho"
	"errors"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/frame"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/utils"
	"sync"
)

type archID = macho.Cpu
type File = macho.File

var supportedArchs = map[macho.Cpu]interface{}{
	macho.CpuAmd64: nil,
	macho.CpuArm64: nil,
}

const crosscall2SPOffset = 0x58



func loadBinaryInfo(bi *BinaryInfo, image *Image, path string, entryPoint uint64) error {
	exe, err := macho.Open(path)
	if err != nil {
		return err
	}
	if entryPoint != 0 {
		
		
		
		
		image.StaticBase = entryPoint - 0x100000000
	}
	image.closer = exe
	if !isSupportedArch(exe.Cpu) {
		return errors.New("unsupported darwin arch")
	}
	image.Dwarf, err = exe.DWARF()
	if err != nil {
		return err
	}
	debugInfoBytes, err := GetDebugSection(exe, "info")
	if err != nil {
		return err
	}

	debugLineBytes, err := GetDebugSection(exe, "line")
	if err != nil {
		return err
	}
	bi.debugLocBytes, _ = GetDebugSection(exe, "loc")
	bi.debugLoclistBytes, _ = GetDebugSection(exe, "loclists")
	debugAddrBytes, _ := GetDebugSection(exe, "addr")
	image.debugAddr = godwarf.ParseAddr(debugAddrBytes)
	debugLineStrBytes, _ := GetDebugSection(exe, "line_str")
	image.debugLineStr = debugLineStrBytes

	wg := &sync.WaitGroup{}
	wg.Add(2)
	utils.CreateGoroutine(func() {
		defer wg.Done()
		err = bi.parseDebugFrame(image, exe, debugInfoBytes)
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
	bi.macOSDebugFrameBugWorkaround()
	return nil
}

func getSectionName(section string) string {
	return "__debug_" + section
}

func getCompressedSectionName(section string) string {
	return "__zdebug_" + section
}

func getEhFrameSection(f *macho.File) *macho.Section {
	return f.Section("__eh_frame")
}





















func (bi *BinaryInfo) macOSDebugFrameBugWorkaround() {
	if len(bi.Images) > 1 {
		
		
		
		return
	}

	
	var fn *Function
	for i := range bi.Functions {
		if bi.Functions[i].cu.IsGo && bi.Functions[i].Entry > 0 {
			fn = &bi.Functions[i]
			break
		}
	}
	if fn == nil {
		
		return
	}

	if fde, _ := bi.FrameEntries.FDEForPC(fn.Entry); fde != nil {
		
		
		return
	}

	
	var fde *frame.FrameDescriptionEntry
	for i := range bi.FrameEntries {
		if bi.FrameEntries[i].CIE.CIE_id == ^uint32(0) {
			fde = bi.FrameEntries[i]
			break
		}
	}

	if fde == nil {
		
		return
	}

	fnsize := fn.End - fn.Entry

	if fde.End()-fde.Begin() != fnsize || fde.Begin() > fn.Entry {
		
		return
	}

	delta := fn.Entry - fde.Begin()

	

	for i := range bi.FrameEntries {
		if bi.FrameEntries[i].CIE.CIE_id == ^uint32(0) {
			bi.FrameEntries[i].Translate(delta)
		}
	}
}
