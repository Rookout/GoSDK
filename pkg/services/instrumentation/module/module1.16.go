//go:build go1.16 && !go1.18
// +build go1.16,!go1.18

package module

import (
	"unsafe"
)



type funcFlag uint8

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

type _func struct {
	entry   uintptr 
	nameoff int32   

	args        int32  
	deferreturn uint32 

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32  
	funcID    funcID  
	_         [2]byte 
	nfuncdata uint8   
}

type pcHeader struct {
	magic          uint32  
	pad1, pad2     uint8   
	minLC          uint8   
	ptrSize        uint8   
	nfunc          int     
	nfiles         uint    
	funcnameOffset uintptr 
	cuOffset       uintptr 
	filetabOffset  uintptr 
	pctabOffset    uintptr 
	pclnOffset     uintptr 
}

//go:linkname moduledata runtime.moduledata
type moduledata struct {
	pcHeader    *pcHeader
	funcnametab []byte
	cutab       []uint32
	filetab     []byte
	pctab       []byte
	
	pclntable    []byte
	ftab         []functab
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

func getPCTab(m *moduledata) []byte {
	return m.pctab
}

func (f *FuncInfo) getEntry() uintptr {
	return f.entry
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
