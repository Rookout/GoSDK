package binary_info

import (
	"bytes"
	"compress/zlib"
	"debug/dwarf"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/frame"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/line"
	loclist2 "github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/loclist"
	op2 "github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/reader"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/hashicorp/golang-lru/simplelru"
)

type Function struct {
	Name       string
	Entry, End uint64 
	Offset     dwarf.Offset
	cu         *compileUnit

	
	InlinedCalls []InlinedCall
	trampoline   bool
}

const crosscall2SPOffsetBad = 0x8

type BinaryInfo struct {
	
	
	
	sigreturnfn *Function
	
	
	
	
	crosscall2fn      *Function
	debugLocBytes     []byte
	debugLoclistBytes []byte
	PointerSize       int

	debugInfoDirectories []string

	
	Functions []Function
	
	Sources []string
	
	LookupFunc map[string]*Function

	
	SymNames map[uint64]*elf.Symbol

	
	
	Images []*Image

	ElfDynamicSection ElfDynamicSection

	lastModified time.Time 

	closer         io.Closer
	sepDebugCloser io.Closer

	
	
	
	
	
	
	
	
	
	PackageMap map[string][]string

	FrameEntries frame.FrameDescriptionEntries

	types       map[string]dwarfRef
	packageVars []packageVar 

	gStructOffset uint64

	
	
	
	NameOfRuntimeType map[uint64]NameOfRuntimeTypeEntry

	
	consts constantsMap

	
	
	
	inlinedCallLines map[fileLine][]uint64

	Dwarf     *dwarf.Data
	TypeCache sync.Map
}

type NameOfRuntimeTypeEntry struct {
	Typename string
	Kind     int64
}

type fileLine struct {
	file string
	line int
}


type dwarfRef struct {
	imageIndex int
	offset     dwarf.Offset
}


type InlinedCall struct {
	cu            *compileUnit
	LowPC, HighPC uint64 
}




type packageVar struct {
	name   string
	cu     *compileUnit
	offset dwarf.Offset
	addr   uint64
}

type buildIDHeader struct {
	Namesz uint32
	Descsz uint32
	Type   uint32
}


type ElfDynamicSection struct {
	Addr uint64 
	Size uint64 
}


func NewBinaryInfo() *BinaryInfo {
	pointerSize := 4 << (^uintptr(0) >> 63) 
	r := &BinaryInfo{NameOfRuntimeType: make(map[uint64]NameOfRuntimeTypeEntry), PointerSize: pointerSize}
	return r
}


func (bi *BinaryInfo) LoadBinaryInfo(path string, entryPoint uint64, debugInfoDirs []string) error {
	bi.debugInfoDirectories = debugInfoDirs
	

	return bi.AddImage(path, entryPoint)
}

var dwarfTreeCacheSize = 512 





func (bi *BinaryInfo) AddImage(path string, addr uint64) error {
	
	if len(bi.Images) > 0 && !strings.HasPrefix(path, "/") {
		return nil
	}
	for _, image := range bi.Images {
		if image.Path == path && image.addr == addr {
			return nil
		}
	}

	
	image := &Image{Path: path, addr: addr}
	image.dwarfTreeCache, _ = simplelru.NewLRU(dwarfTreeCacheSize, nil)

	
	image.Index = len(bi.Images)
	bi.Images = append(bi.Images, image)
	err := loadBinaryInfo(bi, image, path, addr)
	if err != nil {
		bi.Images[len(bi.Images)-1].loadErr = err
	}
	return err
}

func isSupportedArch(a archID) bool {
	if _, ok := supportedArchs[a]; ok {
		return true
	}
	return false
}



func GetDebugSection(f *File, name string) ([]byte, error) {
	sec := f.Section(getSectionName(name))
	if sec != nil {
		return sec.Data()
	}
	sec = f.Section(getCompressedSectionName(name))
	if sec == nil {
		return nil, fmt.Errorf("could not find .debug_%s section", name)
	}
	b, err := sec.Data()
	if err != nil {
		return nil, err
	}
	return decompressMaybe(b)
}

func decompressMaybe(b []byte) ([]byte, error) {
	if len(b) < 12 || string(b[:4]) != "ZLIB" {
		
		return b, nil
	}

	dlen := binary.BigEndian.Uint64(b[4:12])
	dbuf := make([]byte, dlen)
	r, err := zlib.NewReader(bytes.NewBuffer(b[12:]))
	if err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(r, dbuf); err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}
	return dbuf, nil
}




func (bi *BinaryInfo) parseDebugFrame(image *Image, exe *File, debugInfoBytes []byte) error {
	debugFrameData, err := GetDebugSection(exe, "frame")
	ehFrameSection := getEhFrameSection(exe)
	if ehFrameSection == nil && debugFrameData == nil {
		return fmt.Errorf("could not get .debug_frame section and .eh_frame section: %v", err)
	}
	var ehFrameData []byte
	var ehFrameAddr uint64
	if ehFrameSection != nil {
		ehFrameAddr = ehFrameSection.Addr
		ehFrameData, _ = ehFrameSection.Data()
	}
	byteOrder := frame.DwarfEndian(debugInfoBytes)

	if debugFrameData != nil {
		fe, err := frame.Parse(debugFrameData, byteOrder, image.StaticBase, bi.PointerSize, 0)
		if err != nil {
			return fmt.Errorf("could not parse .debug_frame section: %v", err)
		}
		bi.FrameEntries = bi.FrameEntries.Append(fe)
	}

	if ehFrameData != nil && ehFrameAddr > 0 {
		fe, err := frame.Parse(ehFrameData, byteOrder, image.StaticBase, bi.PointerSize, ehFrameAddr)
		if err != nil {
			if debugFrameData == nil {
				return fmt.Errorf("could not parse .eh_frame section: %v", err)
			}
			return nil
		}
		bi.FrameEntries = bi.FrameEntries.Append(fe)
	}

	return nil
}

func shouldFilterSource(path string) bool {
	if utils.Contains(utils.TrueValues, os.Getenv("ROOKOUT_DONT_FILTER_SOURCES")) {
		return false
	}

	
	if strings.Contains(path, "shouldrunprologue") || strings.Contains(path, "prepforcallback") {
		return false
	}

	return strings.Contains(path, "gorook") || strings.Contains(path, "gosdk")
}

func (bi *BinaryInfo) loadSources(compileUnits []*compileUnit) {
	for _, cu := range compileUnits {
		if cu.lineInfo == nil {
			continue
		}
		for _, fileEntry := range cu.lineInfo.FileNames {
			if shouldFilterSource(fileEntry.Path) {
				continue
			}
			bi.Sources = append(bi.Sources, fileEntry.Path)
		}
	}
	sort.Strings(bi.Sources)
	bi.Sources = uniq(bi.Sources)
}

func (bi *BinaryInfo) loadDebugInfoMaps(image *Image, debugInfoBytes, debugLineBytes []byte) error {
	if bi.types == nil {
		bi.types = make(map[string]dwarfRef)
	}
	if bi.consts == nil {
		bi.consts = make(map[dwarfRef]*constantType)
	}
	if bi.PackageMap == nil {
		bi.PackageMap = make(map[string][]string)
	}
	if bi.inlinedCallLines == nil {
		bi.inlinedCallLines = make(map[fileLine][]uint64)
	}

	image.RuntimeTypeToDIE = make(map[uint64]runtimeTypeDIE)

	ctxt := newLoadDebugInfoMapsContext(bi, image, util.ReadUnitVersions(debugInfoBytes))

	reader := image.Dwarf.Reader()

	for entry, err := reader.Next(); entry != nil; entry, err = reader.Next() {
		if err != nil {
			return errors.New("error reading debug_info")
		}
		switch entry.Tag {
		case dwarf.TagCompileUnit:
			cu := &compileUnit{}
			cu.image = image
			cu.entry = entry
			cu.offset = entry.Offset
			cu.Version = ctxt.offsetToVersion[cu.offset]
			if lang, _ := entry.Val(dwarf.AttrLanguage).(int64); lang == godwarf.DW_LANG_Go {
				cu.IsGo = true
			}
			cu.name, _ = entry.Val(dwarf.AttrName).(string)
			compdir, _ := entry.Val(dwarf.AttrCompDir).(string)
			if compdir != "" {
				cu.name = filepath.Join(compdir, cu.name)
			}

			if shouldFilterSource(cu.name) {
				continue
			}

			cu.ranges, _ = image.Dwarf.Ranges(entry)
			for i := range cu.ranges {
				cu.ranges[i][0] += image.StaticBase
				cu.ranges[i][1] += image.StaticBase
			}
			if len(cu.ranges) >= 1 {
				cu.lowPC = cu.ranges[0][0]
			}
			lineInfoOffset, hasLineInfo := entry.Val(dwarf.AttrStmtList).(int64)
			if hasLineInfo && lineInfoOffset >= 0 && lineInfoOffset < int64(len(debugLineBytes)) {
				cu.lineInfo = line.Parse(compdir, bytes.NewBuffer(debugLineBytes[lineInfoOffset:]), image.debugLineStr, nil, image.StaticBase, runtime.GOOS == "windows", bi.PointerSize)
			}
			cu.producer, _ = entry.Val(dwarf.AttrProducer).(string)
			if cu.IsGo && cu.producer != "" {
				semicolon := strings.Index(cu.producer, ";")
				if semicolon < 0 {
					cu.optimized = GoVersionAfterOrEqual(1, 10)
				} else {
					cu.optimized = !strings.Contains(cu.producer[semicolon:], "-N") || !strings.Contains(cu.producer[semicolon:], "-l")
					cu.producer = cu.producer[:semicolon]
				}
			}
			gopkg, _ := entry.Val(godwarf.AttrGoPackageName).(string)
			if cu.IsGo && gopkg != "" {
				bi.PackageMap[gopkg] = append(bi.PackageMap[gopkg], escapePackagePath(strings.Replace(cu.name, "\\", "/", -1)))
			}
			image.compileUnits = append(image.compileUnits, cu)
			if entry.Children {
				err := bi.loadDebugInfoMapsCompileUnit(ctxt, image, reader, cu)
				if err != nil {
					return err
				}
			}

		case dwarf.TagPartialUnit:
			reader.SkipChildren()

		default:
			
			reader.SkipChildren()
		}
	}

	sort.Sort(compileUnitsByOffset(image.compileUnits))
	sort.Sort(functionsDebugInfoByEntry(bi.Functions))
	sort.Sort(packageVarsByAddr(bi.packageVars))

	bi.LookupFunc = make(map[string]*Function)
	for i := range bi.Functions {
		bi.LookupFunc[bi.Functions[i].Name] = &bi.Functions[i]
	}
	bi.sigreturnfn = bi.LookupFunc["runtime.sigreturn"]
	bi.crosscall2fn = bi.LookupFunc["crosscall2"]

	bi.loadSources(image.compileUnits)
	return nil
}

type loadDebugInfoMapsContext struct {
	ardr                *dwarf.Reader
	abstractOriginTable map[dwarf.Offset]int
	knownPackageVars    map[string]struct{}
	offsetToVersion     map[dwarf.Offset]uint8
}

func newLoadDebugInfoMapsContext(bi *BinaryInfo, image *Image, offsetToVersion map[dwarf.Offset]uint8) *loadDebugInfoMapsContext {
	ctxt := &loadDebugInfoMapsContext{}

	ctxt.ardr = image.Dwarf.Reader()
	ctxt.abstractOriginTable = make(map[dwarf.Offset]int)
	ctxt.offsetToVersion = offsetToVersion

	ctxt.knownPackageVars = map[string]struct{}{}
	for _, v := range bi.packageVars {
		ctxt.knownPackageVars[v.name] = struct{}{}
	}

	return ctxt
}




func escapePackagePath(pkg string) string {
	slash := strings.Index(pkg, "/")
	if slash < 0 {
		slash = 0
	}
	return pkg[:slash] + strings.Replace(pkg[slash:], ".", "%2e", -1)
}

type functionsDebugInfoByEntry []Function

func (v functionsDebugInfoByEntry) Len() int           { return len(v) }
func (v functionsDebugInfoByEntry) Less(i, j int) bool { return v[i].Entry < v[j].Entry }
func (v functionsDebugInfoByEntry) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

type packageVarsByAddr []packageVar

func (v packageVarsByAddr) Len() int               { return len(v) }
func (v packageVarsByAddr) Less(i int, j int) bool { return v[i].addr < v[j].addr }
func (v packageVarsByAddr) Swap(i int, j int)      { v[i], v[j] = v[j], v[i] }


func (bi *BinaryInfo) loadDebugInfoMapsCompileUnit(ctxt *loadDebugInfoMapsContext, image *Image, reader *dwarf.Reader, cu *compileUnit) error {
	hasAttrGoPkgName := GoVersionAfterOrEqual(1, 13)

	depth := 0

	for entry, err := reader.Next(); entry != nil; entry, err = reader.Next() {
		if err != nil {
			return errors.New("error reading debug_info")
		}
		switch entry.Tag {
		case 0:
			if depth == 0 {
				return nil
			} else {
				depth--
			}
		case dwarf.TagImportedUnit:
			err = bi.loadDebugInfoMapsImportedUnit(entry, ctxt, image, cu)
			if err != nil {
				return err
			}
			reader.SkipChildren()

		case dwarf.TagArrayType, dwarf.TagBaseType, dwarf.TagClassType, dwarf.TagStructType, dwarf.TagUnionType, dwarf.TagConstType, dwarf.TagVolatileType, dwarf.TagRestrictType, dwarf.TagEnumerationType, dwarf.TagPointerType, dwarf.TagSubroutineType, dwarf.TagTypedef, dwarf.TagUnspecifiedType:
			if name, ok := entry.Val(dwarf.AttrName).(string); ok {
				if !cu.IsGo {
					name = "C." + name
				}
				if _, exists := bi.types[name]; !exists {
					bi.types[name] = dwarfRef{image.Index, entry.Offset}
				}
			}
			if cu != nil && cu.IsGo && !hasAttrGoPkgName {
				bi.registerTypeToPackageMap(entry)
			}
			image.registerRuntimeTypeToDIE(entry)
			reader.SkipChildren()

		case dwarf.TagVariable:
			if n, ok := entry.Val(dwarf.AttrName).(string); ok {
				var addr uint64
				if loc, ok := entry.Val(dwarf.AttrLocation).([]byte); ok {
					if len(loc) == bi.PointerSize+1 && op2.Opcode(loc[0]) == op2.DW_OP_addr {
						addr, _ = util.ReadUintRaw(bytes.NewReader(loc[1:]), binary.LittleEndian, bi.PointerSize)
					}
				}
				if !cu.IsGo {
					n = "C." + n
				}
				if _, known := ctxt.knownPackageVars[n]; !known {
					bi.packageVars = append(bi.packageVars, packageVar{n, cu, entry.Offset, addr + image.StaticBase})
				}
			}
			reader.SkipChildren()

		case dwarf.TagConstant:
			name, okName := entry.Val(dwarf.AttrName).(string)
			typ, okType := entry.Val(dwarf.AttrType).(dwarf.Offset)
			val, okVal := entry.Val(dwarf.AttrConstValue).(int64)
			if okName && okType && okVal {
				if !cu.IsGo {
					name = "C." + name
				}
				ct := bi.consts[dwarfRef{image.Index, typ}]
				if ct == nil {
					ct = &constantType{}
					bi.consts[dwarfRef{image.Index, typ}] = ct
				}
				ct.values = append(ct.values, constantValue{name: name, fullName: name, value: val})
			}
			reader.SkipChildren()

		case dwarf.TagSubprogram:
			inlined := false
			if inval, ok := entry.Val(dwarf.AttrInline).(int64); ok {
				inlined = inval >= 1
			}

			if inlined {
				err = bi.addAbstractSubprogram(entry, ctxt, reader, cu)
				if err != nil {
					return err
				}
			} else {
				originOffset, hasAbstractOrigin := entry.Val(dwarf.AttrAbstractOrigin).(dwarf.Offset)
				if hasAbstractOrigin {
					err = bi.addConcreteInlinedSubprogram(entry, originOffset, ctxt, reader, cu)
					if err != nil {
						return err
					}
				} else {
					err = bi.addConcreteSubprogram(entry, ctxt, reader, cu)
					if err != nil {
						return err
					}
				}
			}

		default:
			if entry.Children {
				depth++
			}
		}
	}

	return nil
}



func (bi *BinaryInfo) loadDebugInfoMapsImportedUnit(entry *dwarf.Entry, ctxt *loadDebugInfoMapsContext, image *Image, cu *compileUnit) error {
	off, ok := entry.Val(dwarf.AttrImport).(dwarf.Offset)
	if !ok {
		return nil
	}
	reader := image.Dwarf.Reader()
	reader.Seek(off)
	imentry, err := reader.Next()
	if err != nil {
		return nil
	}
	if imentry.Tag != dwarf.TagPartialUnit {
		return nil
	}
	return bi.loadDebugInfoMapsCompileUnit(ctxt, image, reader, cu)
}

func (bi *BinaryInfo) registerTypeToPackageMap(entry *dwarf.Entry) {
	if entry.Tag != dwarf.TagTypedef && entry.Tag != dwarf.TagBaseType && entry.Tag != dwarf.TagClassType && entry.Tag != dwarf.TagStructType {
		return
	}

	typename, ok := entry.Val(dwarf.AttrName).(string)
	if !ok || complexType(typename) {
		return
	}

	dot := strings.LastIndex(typename, ".")
	if dot < 0 {
		return
	}
	path := typename[:dot]
	slash := strings.LastIndex(path, "/")
	if slash < 0 || slash+1 >= len(path) {
		return
	}
	name := path[slash+1:]
	bi.PackageMap[name] = []string{path}
}

func (bi *BinaryInfo) addConcreteInlinedSubprogram(entry *dwarf.Entry, originOffset dwarf.Offset, ctxt *loadDebugInfoMapsContext, reader *dwarf.Reader, cu *compileUnit) error {
	lowpc, highpc, ok := subprogramEntryRange(entry, cu.image)
	if !ok {
		if entry.Children {
			reader.SkipChildren()
		}
		return nil
	}

	originIdx, ok := ctxt.abstractOriginTable[originOffset]
	if !ok {
		if entry.Children {
			reader.SkipChildren()
		}
		return nil
	}

	fn := &bi.Functions[originIdx]
	fn.Offset = entry.Offset
	fn.Entry = lowpc
	fn.End = highpc

	if entry.Children {
		err := bi.loadDebugInfoMapsInlinedCalls(ctxt, reader, cu)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bi *BinaryInfo) loadDebugInfoMapsInlinedCalls(ctxt *loadDebugInfoMapsContext, reader *dwarf.Reader, cu *compileUnit) error {
	for {
		entry, err := reader.Next()
		if err != nil {
			return errors.New("error reading debug_info")
		}
		switch entry.Tag {
		case 0:
			return nil
		case dwarf.TagInlinedSubroutine:
			originOffset, ok := entry.Val(dwarf.AttrAbstractOrigin).(dwarf.Offset)
			if !ok {
				reader.SkipChildren()
				continue
			}

			originIdx, ok := ctxt.abstractOriginTable[originOffset]
			if !ok {
				reader.SkipChildren()
				continue
			}
			fn := &bi.Functions[originIdx]

			lowpc, highpc, ok := subprogramEntryRange(entry, cu.image)
			if !ok {
				reader.SkipChildren()
				continue
			}

			callfileidx, ok1 := entry.Val(dwarf.AttrCallFile).(int64)
			callline, ok2 := entry.Val(dwarf.AttrCallLine).(int64)
			if !ok1 || !ok2 {
				reader.SkipChildren()
				continue
			}
			if cu.lineInfo == nil {
				reader.SkipChildren()
				continue
			}
			if int(callfileidx-1) >= len(cu.lineInfo.FileNames) {
				reader.SkipChildren()
				continue
			}
			callfile := cu.lineInfo.FileNames[callfileidx-1].Path

			fn.InlinedCalls = append(fn.InlinedCalls, InlinedCall{
				cu:     cu,
				LowPC:  lowpc,
				HighPC: highpc,
			})

			fl := fileLine{callfile, int(callline)}
			bi.inlinedCallLines[fl] = append(bi.inlinedCallLines[fl], lowpc)
		}
		reader.SkipChildren()
	}
}

func subprogramEntryRange(entry *dwarf.Entry, image *Image) (lowpc, highpc uint64, ok bool) {
	ok = false
	if ranges, _ := image.Dwarf.Ranges(entry); len(ranges) >= 1 {
		ok = true
		lowpc = ranges[0][0] + image.StaticBase
		highpc = ranges[0][1] + image.StaticBase
	}
	return lowpc, highpc, ok
}

func (bi *BinaryInfo) addConcreteSubprogram(entry *dwarf.Entry, ctxt *loadDebugInfoMapsContext, reader *dwarf.Reader, cu *compileUnit) error {
	lowpc, highpc, ok := subprogramEntryRange(entry, cu.image)
	if !ok {
		if entry.Children {
			reader.SkipChildren()
		}
		return nil
	}

	name, ok := subprogramEntryName(entry, cu)
	if !ok {
		if entry.Children {
			reader.SkipChildren()
		}
		return nil
	}

	fn := Function{
		Name:   name,
		Entry:  lowpc,
		End:    highpc,
		Offset: entry.Offset,
		cu:     cu,
	}
	bi.Functions = append(bi.Functions, fn)

	if entry.Children {
		err := bi.loadDebugInfoMapsInlinedCalls(ctxt, reader, cu)
		if err != nil {
			return err
		}
	}

	return nil
}

func subprogramEntryName(entry *dwarf.Entry, cu *compileUnit) (string, bool) {
	name, ok := entry.Val(dwarf.AttrName).(string)
	if !ok {
		return "", false
	}
	if !cu.IsGo {
		name = "C." + name
	}
	return name, true
}

func (bi *BinaryInfo) addAbstractSubprogram(entry *dwarf.Entry, ctxt *loadDebugInfoMapsContext, reader *dwarf.Reader, cu *compileUnit) error {
	name, ok := subprogramEntryName(entry, cu)
	if !ok {
		if entry.Children {
			reader.SkipChildren()
		}
		return nil
	}

	fn := Function{
		Name:   name,
		Offset: entry.Offset,
		cu:     cu,
	}

	if entry.Children {
		err := bi.loadDebugInfoMapsInlinedCalls(ctxt, reader, cu)
		if err != nil {
			return err
		}
	}

	bi.Functions = append(bi.Functions, fn)
	ctxt.abstractOriginTable[entry.Offset] = len(bi.Functions) - 1
	return nil
}

func (bi *BinaryInfo) getBestMatchingFile(filename string) ([]*compileUnit, string, rookoutErrors.RookoutError) {
	
	var topCu []*compileUnit
	fm := utils.NewFileMatcher()
	for _, image := range bi.Images {
		for _, cu := range image.compileUnits {
			if cu.lineInfo == nil {
				continue
			}
			for _, f := range cu.lineInfo.FileNames {
				matchScore := utils.GetPathMatchingScore(filename, f.Path)
				switch fm.UpdateMatch(matchScore, f.Path) {
				case utils.NewBestMatch:
					logger.Logger().Debugf("NewBestMatch: filepath: %s", f.Path)
					topCu = []*compileUnit{cu}
				case utils.SameBestMatch:
					logger.Logger().Debugf("SameBestMatch: filepath: %s", f.Path)
					topCu = append(topCu, cu)
				}
			}
		}
	}

	if !fm.AnyMatch() {
		
		return nil, "", rookoutErrors.NewFileNotFound(filename)
	}
	if !fm.IsUnique() {
		
		return nil, "", rookoutErrors.NewMultipleFilesFound(filename)
	}
	
	return topCu, fm.GetBestFile(), nil
}



func (bi *BinaryInfo) PCToInlineFunc(pc uint64) *Function {
	fn := bi.PCToFunc(pc)
	dwarfTree, err := fn.cu.image.GetDwarfTree(fn.Offset)
	if err != nil {
		return fn
	}
	entries := reader.InlineStack(dwarfTree, pc)
	if len(entries) == 0 {
		return fn
	}

	fnname, okname := entries[0].Val(dwarf.AttrName).(string)
	if !okname {
		return fn
	}

	return bi.LookupFunc[fnname]
}



func (bi *BinaryInfo) PCToFunc(pc uint64) *Function {
	i := sort.Search(len(bi.Functions), func(i int) bool {
		fn := bi.Functions[i]
		return pc <= fn.Entry || (fn.Entry <= pc && pc < fn.End)
	})
	if i != len(bi.Functions) {
		fn := &bi.Functions[i]
		if fn.Entry <= pc && pc < fn.End {
			return fn
		}
	}
	return nil
}

func (bi *BinaryInfo) FuncToImage(fn *Function) *Image {
	if fn == nil {
		return bi.Images[0]
	}

	return fn.cu.image
}


func (bi *BinaryInfo) PCToLine(pc uint64) (string, int, *Function) {
	fn := bi.PCToFunc(pc)
	if fn == nil {
		return "", 0, nil
	}
	f, ln := fn.cu.lineInfo.PCToLine(fn.Entry, pc)
	return f, ln, fn
}

func (bi *BinaryInfo) LocationExpr(entry godwarf.Entry, attr dwarf.Attr, pc uint64) ([]byte, *LocationExpr, error) {
	
	a := entry.Val(attr)
	if a == nil {
		return nil, nil, fmt.Errorf("no location attribute %s", attr)
	}
	if instr, ok := a.([]byte); ok {
		return instr, &LocationExpr{isBlock: true, instr: instr}, nil
	}
	off, ok := a.(int64)
	if !ok {
		return nil, nil, fmt.Errorf("could not interpret location attribute %s", attr)
	}
	instr := bi.loclistEntry(off, pc)
	if instr == nil {
		return nil, nil, fmt.Errorf("could not find loclist entry at %#x for address %#x", off, pc)
	}
	return instr, &LocationExpr{pc: pc, off: off, instr: instr}, nil
}

type LocationExpr struct {
	isBlock   bool
	isEscaped bool
	off       int64
	pc        uint64
	instr     []byte
}



func (bi *BinaryInfo) loclistEntry(off int64, pc uint64) []byte {
	var base uint64
	image := bi.Images[0]
	cu := bi.findCompileUnit(pc)
	if cu != nil {
		base = cu.lowPC
		image = cu.image
	}
	if image == nil {
		return nil
	}

	var loclist loclist2.Reader = bi.newLoclist2Reader()
	var debugAddr *godwarf.DebugAddr
	loclist5 := bi.newLoclist5Reader()
	if cu != nil && cu.Version >= 5 && loclist5 != nil {
		loclist = loclist5
		if addrBase, ok := cu.entry.Val(dwarf.AttrAddrBase).(int64); ok {
			debugAddr = image.debugAddr.GetSubsection(uint64(addrBase))
		}
	}

	if loclist.Empty() {
		return nil
	}

	e, err := loclist.Find(int(off), image.StaticBase, base, pc, debugAddr)
	if err != nil {
		logger.Logger().Errorf("error reading loclist section: %v", err)
		return nil
	}
	if e != nil {
		return e.Instr
	}

	return nil
}


func (bi *BinaryInfo) findCompileUnit(pc uint64) *compileUnit {
	for _, image := range bi.Images {
		for _, cu := range image.compileUnits {
			if cu.pcInRange(pc) {
				return cu
			}
		}
	}
	return nil
}

func (bi *BinaryInfo) newLoclist2Reader() *loclist2.Dwarf2Reader {
	return loclist2.NewDwarf2Reader(bi.debugLocBytes, bi.PointerSize)
}

func (bi *BinaryInfo) newLoclist5Reader() *loclist2.Dwarf5Reader {
	return loclist2.NewDwarf5Reader(bi.debugLoclistBytes)
}


func (bi *BinaryInfo) PCToImage(pc uint64) *Image {
	fn := bi.PCToFunc(pc)
	return fn.cu.image
}





func (bi *BinaryInfo) Location(entry godwarf.Entry, attr dwarf.Attr, pc uint64, regs op2.DwarfRegisters) (int64, []op2.Piece, *LocationExpr, error) {
	instr, descr, err := bi.LocationExpr(entry, attr, pc)
	if err != nil {
		return 0, nil, nil, err
	}
	addr, pieces, err := op2.ExecuteStackProgram(regs, instr, bi.PointerSize)
	return addr, pieces, descr, err
}


func (bi *BinaryInfo) FindType(name string) (godwarf.Type, error) {
	ref, found := bi.types[name]
	if !found {
		return nil, errors.New("no type entry found, use 'types' for a list of valid types")
	}
	image := bi.Images[ref.imageIndex]
	return godwarf.ReadType(image.Dwarf, ref.imageIndex, ref.offset, &image.TypeCache)
}

func (bi *BinaryInfo) GetConst(typ godwarf.Type) *constantType {
	return bi.consts.Get(typ)
}

func (bi *BinaryInfo) ReadVariableEntry(entry *godwarf.Tree) (name string, typ godwarf.Type, err error) {
	name, ok := entry.Val(dwarf.AttrName).(string)
	if !ok {
		return "", nil, fmt.Errorf("malformed variable DIE (name)")
	}

	typ, err = entry.Type(bi.Dwarf, 0, &bi.TypeCache)
	if err != nil {
		return "", nil, err
	}

	return name, typ, nil
}
