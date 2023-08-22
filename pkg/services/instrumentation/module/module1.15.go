//go:build go1.15 && !go1.16
// +build go1.15,!go1.16

package module

import (
	"unsafe"
)




type pcHeader struct{}

//go:linkname functab runtime.functab
type functab struct {
	entry   uintptr
	funcoff uintptr
}

//go:linkname textsect runtime.textsect
type textsect struct {
	vaddr    uintptr 
	length   uintptr 
	baseaddr uintptr 
}

//go:linkname _func runtime._func
type _func struct {
	entry   uintptr 
	nameoff int32   

	args        int32  
	deferreturn uint32 

	pcsp      int32
	pcfile    int32
	pcln      int32
	npcdata   int32
	funcID    FuncID  
	_         [2]int8 
	nfuncdata uint8   
}

//go:linkname moduledata runtime.moduledata
type moduledata struct {
	pclntable    []byte
	ftab         []functab
	filetab      []uint32
	findfunctab  uintptr
	minpc, maxpc uintptr

	text, etext           uintptr
	noptrdata, enoptrdata uintptr
	data, edata           uintptr
	bss, ebss             uintptr
	noptrbss, enoptrbss   uintptr
	end, gcdata, gcbss    uintptr
	types, etypes         uintptr

	textsectmap []textsect
	typelinks   []int32 
	itablinks   []*itab

	ptab []ptabEntry

	pluginpath string
	pkghashes  []modulehash

	modulename   string
	modulehashes []modulehash

	hasmain uint8 

	gcdatamask, gcbssmask bitvector

	typemap map[TypeOff]uintptr 

	bad bool 

	next *moduledata
}

//go:linkname funcfile runtime.funcfile
func funcfile(f FuncInfo, fileno int32) string

func getPCTab(m *moduledata) []byte {
	return m.pclntable
}

func (f *FuncInfo) getEntry() uintptr {
	return uintptr(f.entry)
}


func findFuncOffsetInModule(pc uintptr, datap *moduledata) (uintptr, bool) {
	if datap == nil {
		return 0, false
	}
	const nsub = uintptr(len(findfuncbucket{}.subbuckets))

	x := pc - datap.minpc
	b := x / pcbucketsize
	i := x % pcbucketsize / (pcbucketsize / nsub)

	//goland:noinspection GoVetUnsafePointer
	ffb := (*findfuncbucket)(add(unsafe.Pointer(datap.findfunctab), b*unsafe.Sizeof(findfuncbucket{})))
	idx := ffb.idx + uint32(ffb.subbuckets[i])

	
	
	

	if idx >= uint32(len(datap.ftab)) {
		idx = uint32(len(datap.ftab) - 1)
	}
	if pc < datap.ftab[idx].entry {
		
		

		for datap.ftab[idx].entry > pc && idx > 0 {
			idx--
		}
		if idx == 0 {
			
			println("findfunc: bad findfunctab entry idx")
		}
	} else {
		
		for datap.ftab[idx+1].entry <= pc {
			idx++
		}
	}
	funcoff := datap.ftab[idx].funcoff
	if funcoff == ^uintptr(0) {
		
		
		
		
		return 0, false
	}
	return funcoff, true
}

func (md *moduledata) GetTypeMap() map[TypeOff]uintptr {
	return md.typemap
}
