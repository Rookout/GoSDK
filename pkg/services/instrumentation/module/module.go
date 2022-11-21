package module

import "unsafe"



//go:linkname firstModuleData runtime.firstmoduledata
var firstModuleData moduledata

//go:linkname funcName runtime.funcname
//goland:noinspection GoUnusedParameter
func funcName(f FuncInfo) string

//go:linkname FindFunc runtime.findfunc
func FindFunc(_ uintptr) FuncInfo

//go:linkname add runtime.add
func add(_ unsafe.Pointer, _ uintptr) unsafe.Pointer

//go:linkname funcMaxSPDelta runtime.funcMaxSPDelta
func funcMaxSPDelta(_ FuncInfo) int32

//go:linkname step runtime.step
//goland:noinspection GoUnusedParameter
func step(p []byte, pc *uintptr, val *int32, first bool) (newp []byte, ok bool)

type funcID uint8









type findfuncbucket struct {
	idx        uint32
	subbuckets [16]byte
}

type ptabEntry struct {
	name nameOff
	typ  TypeOff
}

type nameOff int32

type TypeOff int32

type itab struct {
	inter *interfacetype
	_type *_type
	hash  uint32 
	_     [4]byte
	fun   [1]uintptr 
}

type _type struct {
	size       uintptr
	ptrdata    uintptr 
	hash       uint32
	tflag      tflag
	align      uint8
	fieldAlign uint8
	kind       uint8
	
	
	equal func(unsafe.Pointer, unsafe.Pointer) bool
	
	
	
	gcdata    *byte
	str       nameOff
	ptrToThis TypeOff
}

type tflag uint8

type interfacetype struct {
	typ     _type
	pkgpath name
	mhdr    []imethod
}

type name struct {
	bytes *byte
}

type imethod struct {
	name nameOff
	ityp TypeOff
}

type modulehash struct {
	modulename   string
	linktimehash string
	runtimehash  *string
}

type bitvector struct {
	n        int32 
	bytedata *uint8
}

type FuncInfo struct {
	*_func
	datap *moduledata
}

const minfunc = 16 

const pcbucketsize = 256 * minfunc 

//go:linkname pcdatastart runtime.pcdatastart
func pcdatastart(_ FuncInfo, _ uint32) uint32



type pcvalueCacheEnt struct {
	
	targetpc uintptr
	off      uint32
	
	val int32
}

type pcvalueCache struct {
	entries [2][8]pcvalueCacheEnt
}

//go:linkname pcvalue runtime.pcvalue
func pcvalue(f FuncInfo, off uint32, targetpc uintptr, cache *pcvalueCache, strict bool) (int32, uintptr)

var moduleDatas []*moduledata

func loadModuleDatas() {
	if moduleDatas != nil {
		return
	}

	for moduleData := &firstModuleData; moduleData != nil; moduleData = moduleData.next {
		moduleDatas = append(moduleDatas, moduleData)
	}
}

func Init(stackUsageMap map[uint64][]map[string]int64) {
	loadModuleDatas()
	loadStackUsageMap(stackUsageMap)
}

func FindModuleDataForType(typeAddr uint64) *moduledata {
	for i := range moduleDatas {
		if typeAddr >= uint64(moduleDatas[i].types) && typeAddr < uint64(moduleDatas[i].etypes) {
			return moduleDatas[i]
		}
	}
	return nil
}

func (md *moduledata) GetFirstPC() uint64 {
	return uint64(md.text)
}

func (md *moduledata) GetTypesAddr() uint64 {
	return uint64(md.types)
}
