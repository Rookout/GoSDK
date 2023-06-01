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

package line

import (
	"bytes"
	"encoding/binary"
	"path"
	"strings"
	"sync"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
)


type DebugLinePrologue struct {
	UnitLength     uint32
	Version        uint16
	Length         uint32
	MinInstrLength uint8
	MaxOpPerInstr  uint8
	InitialIsStmt  uint8
	LineBase       int8
	LineRange      uint8
	OpcodeBase     uint8
	StdOpLengths   []uint8
}


type DebugLineInfo struct {
	Prologue     *DebugLinePrologue
	IncludeDirs  []string
	FileNames    []*FileEntry
	Instructions []byte
	Lookup       map[string]*FileEntry

	Logf func(string, ...interface{})

	stateMachineCacheLock *sync.RWMutex
	
	stateMachineCache map[uint64]*StateMachine

	lastMachineCacheLock *sync.RWMutex
	
	lastMachineCache map[uint64]*StateMachine

	
	debugLineStr []byte

	
	staticBase uint64

	
	normalizeBackslash bool
	ptrSize            int
	endSeqIsValid      bool
}


type FileEntry struct {
	Path        string
	DirIdx      uint64
	LastModTime uint64
	Length      uint64
}

type DebugLines []*DebugLineInfo


func ParseAll(data []byte, debugLineStr []byte, logfn func(string, ...interface{}), staticBase uint64, normalizeBackslash bool, ptrSize int) DebugLines {
	var (
		lines = make(DebugLines, 0)
		buf   = bytes.NewBuffer(data)
	)

	
	for buf.Len() > 0 {
		lines = append(lines, Parse("", buf, debugLineStr, logfn, staticBase, normalizeBackslash, ptrSize))
	}

	return lines
}



func Parse(compdir string, buf *bytes.Buffer, debugLineStr []byte, logfn func(string, ...interface{}), staticBase uint64, normalizeBackslash bool, ptrSize int) *DebugLineInfo {
	dbl := new(DebugLineInfo)
	dbl.Logf = logfn
	if logfn == nil {
		dbl.Logf = func(string, ...interface{}) {}
	}
	dbl.staticBase = staticBase
	dbl.ptrSize = ptrSize
	dbl.Lookup = make(map[string]*FileEntry)
	dbl.IncludeDirs = append(dbl.IncludeDirs, compdir)

	dbl.stateMachineCacheLock = &sync.RWMutex{}
	dbl.stateMachineCacheLock.Lock()
	dbl.stateMachineCache = make(map[uint64]*StateMachine)
	dbl.stateMachineCacheLock.Unlock()
	dbl.lastMachineCacheLock = &sync.RWMutex{}
	dbl.lastMachineCacheLock.Lock()
	dbl.lastMachineCache = make(map[uint64]*StateMachine)
	dbl.lastMachineCacheLock.Unlock()
	dbl.normalizeBackslash = normalizeBackslash
	dbl.debugLineStr = debugLineStr

	parseDebugLinePrologue(dbl, buf)
	if dbl.Prologue.Version >= 5 {
		if !parseIncludeDirs5(dbl, buf) {
			return nil
		}
		if !parseFileEntries5(dbl, buf) {
			return nil
		}
	} else {
		if !parseIncludeDirs2(dbl, buf) {
			return nil
		}
		if !parseFileEntries2(dbl, buf) {
			return nil
		}
	}

	
	
	
	
	dbl.Instructions = buf.Next(int(dbl.Prologue.UnitLength - dbl.Prologue.Length - 6))

	return dbl
}

func parseDebugLinePrologue(dbl *DebugLineInfo, buf *bytes.Buffer) {
	p := new(DebugLinePrologue)

	p.UnitLength = binary.LittleEndian.Uint32(buf.Next(4))
	p.Version = binary.LittleEndian.Uint16(buf.Next(2))
	if p.Version >= 5 {
		dbl.ptrSize = int(buf.Next(1)[0])  
		dbl.ptrSize += int(buf.Next(1)[0]) 
	}

	p.Length = binary.LittleEndian.Uint32(buf.Next(4))
	p.MinInstrLength = uint8(buf.Next(1)[0])
	if p.Version >= 4 {
		p.MaxOpPerInstr = uint8(buf.Next(1)[0])
	} else {
		p.MaxOpPerInstr = 1
	}
	p.InitialIsStmt = uint8(buf.Next(1)[0])
	p.LineBase = int8(buf.Next(1)[0])
	p.LineRange = uint8(buf.Next(1)[0])
	p.OpcodeBase = uint8(buf.Next(1)[0])

	p.StdOpLengths = make([]uint8, p.OpcodeBase-1)
	binary.Read(buf, binary.LittleEndian, &p.StdOpLengths)

	dbl.Prologue = p
}


func parseIncludeDirs2(info *DebugLineInfo, buf *bytes.Buffer) bool {
	for {
		str, err := util.ParseString(buf)
		if err != nil {
			if info.Logf != nil {
				info.Logf("error reading string: %v", err)
			}
			return false
		}
		if str == "" {
			break
		}

		info.IncludeDirs = append(info.IncludeDirs, str)
	}
	return true
}


func parseIncludeDirs5(info *DebugLineInfo, buf *bytes.Buffer) bool {
	dirEntryFormReader := readEntryFormat(buf, info.Logf)
	if dirEntryFormReader == nil {
		return false
	}
	dirCount, _ := util.DecodeULEB128(buf)
	info.IncludeDirs = make([]string, 0, dirCount)
	for i := uint64(0); i < dirCount; i++ {
		dirEntryFormReader.reset()
		for dirEntryFormReader.next(buf) {
			switch dirEntryFormReader.contentType {
			case _DW_LNCT_path:
				switch dirEntryFormReader.formCode {
				case _DW_FORM_string:
					info.IncludeDirs = append(info.IncludeDirs, dirEntryFormReader.str)
				case _DW_FORM_line_strp:
					buf := bytes.NewBuffer(info.debugLineStr[dirEntryFormReader.u64:])
					dir, _ := util.ParseString(buf)
					info.IncludeDirs = append(info.IncludeDirs, dir)
				default:
					info.Logf("unsupported string form %#x", dirEntryFormReader.formCode)
				}
			case _DW_LNCT_directory_index:
			case _DW_LNCT_timestamp:
			case _DW_LNCT_size:
			case _DW_LNCT_MD5:
			}
		}
		if dirEntryFormReader.err != nil {
			if info.Logf != nil {
				info.Logf("error reading directory entries table: %v", dirEntryFormReader.err)
			}
			return false
		}
	}
	return true
}


func parseFileEntries2(info *DebugLineInfo, buf *bytes.Buffer) bool {
	for {
		entry := readFileEntry(info, buf, true)
		if entry == nil {
			return false
		}
		if entry.Path == "" {
			break
		}

		info.FileNames = append(info.FileNames, entry)
		info.Lookup[entry.Path] = entry
	}
	return true
}

func readFileEntry(info *DebugLineInfo, buf *bytes.Buffer, exitOnEmptyPath bool) *FileEntry {
	entry := new(FileEntry)

	var err error
	entry.Path, err = util.ParseString(buf)
	if err != nil {
		if info.Logf != nil {
			info.Logf("error reading file entry: %v", err)
		}
		return nil
	}
	if entry.Path == "" && exitOnEmptyPath {
		return entry
	}

	if info.normalizeBackslash {
		entry.Path = strings.ReplaceAll(entry.Path, "\\", "/")
	}

	entry.DirIdx, _ = util.DecodeULEB128(buf)
	entry.LastModTime, _ = util.DecodeULEB128(buf)
	entry.Length, _ = util.DecodeULEB128(buf)
	if !pathIsAbs(entry.Path) {
		if entry.DirIdx < uint64(len(info.IncludeDirs)) {
			entry.Path = path.Join(info.IncludeDirs[entry.DirIdx], entry.Path)
		}
	}

	return entry
}







func pathIsAbs(s string) bool {
	if len(s) >= 1 && s[0] == '/' {
		return true
	}
	if len(s) >= 2 && s[1] == ':' && (('a' <= s[0] && s[0] <= 'z') || ('A' <= s[0] && s[0] <= 'Z')) {
		return true
	}
	return false
}


func parseFileEntries5(info *DebugLineInfo, buf *bytes.Buffer) bool {
	fileEntryFormReader := readEntryFormat(buf, info.Logf)
	if fileEntryFormReader == nil {
		return false
	}
	fileCount, _ := util.DecodeULEB128(buf)
	info.FileNames = make([]*FileEntry, 0, fileCount)
	for i := 0; i < int(fileCount); i++ {
		fileEntryFormReader.reset()
		for fileEntryFormReader.next(buf) {
			entry := new(FileEntry)
			var p string
			var diridx int
			diridx = -1

			switch fileEntryFormReader.contentType {
			case _DW_LNCT_path:
				switch fileEntryFormReader.formCode {
				case _DW_FORM_string:
					p = fileEntryFormReader.str
				case _DW_FORM_line_strp:
					buf := bytes.NewBuffer(info.debugLineStr[fileEntryFormReader.u64:])
					p, _ = util.ParseString(buf)
				default:
					info.Logf("unsupported string form %#x", fileEntryFormReader.formCode)
				}
			case _DW_LNCT_directory_index:
				diridx = int(fileEntryFormReader.u64)
			case _DW_LNCT_timestamp:
				entry.LastModTime = fileEntryFormReader.u64
			case _DW_LNCT_size:
				entry.Length = fileEntryFormReader.u64
			case _DW_LNCT_MD5:
				
			}

			if info.normalizeBackslash {
				p = strings.ReplaceAll(p, "\\", "/")
			}

			if diridx >= 0 && !pathIsAbs(p) && diridx < len(info.IncludeDirs) {
				p = path.Join(info.IncludeDirs[diridx], p)
			}
			entry.Path = p
			info.FileNames = append(info.FileNames, entry)
			info.Lookup[entry.Path] = entry
		}
		if fileEntryFormReader.err != nil {
			if info.Logf != nil {
				info.Logf("error reading file entries table: %v", fileEntryFormReader.err)
			}
			return false
		}
	}
	return true
}
