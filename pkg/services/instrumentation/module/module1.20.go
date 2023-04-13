//go:build go1.20 && !go1.21
// +build go1.20,!go1.21

package module

import (
	"runtime"
	"unsafe"
)



type funcFlag uint8

type functab struct {
	entryoff uint32 
	funcoff  uint32
}

type textsect struct {
	vaddr    uintptr 
	end      uintptr 
	baseaddr uintptr 
}

type _func struct {
	entryOff uint32 
	nameOff  int32  

	args        int32  
	deferreturn uint32 

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32 
	startLine int32  
	funcID    funcID 
	flag      funcFlag
	_         [1]byte 
	nfuncdata uint8   

	
	
	

	
	
	
	
	
	
	
	

	
	
	
	
	
	
	
	
}

type pcHeader struct {
	magic          uint32  
	pad1, pad2     uint8   
	minLC          uint8   
	ptrSize        uint8   
	nfunc          int     
	nfiles         uint    
	textStart      uintptr 
	funcnameOffset uintptr 
	cuOffset       uintptr 
	filetabOffset  uintptr 
	pctabOffset    uintptr 
	pclnOffset     uintptr 
}

type moduledata struct {
	pcHeader     *pcHeader
	funcnametab  []byte
	cutab        []uint32
	filetab      []byte
	pctab        []byte
	pclntable    []byte
	ftab         []functab
	findfunctab  uintptr
	minpc, maxpc uintptr

	text, etext           uintptr
	noptrdata, enoptrdata uintptr
	data, edata           uintptr
	bss, ebss             uintptr
	noptrbss, enoptrbss   uintptr
	covctrs, ecovctrs     uintptr
	end, gcdata, gcbss    uintptr
	types, etypes         uintptr
	rodata                uintptr
	gofunc                uintptr // go.func.*

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

	typemap map[TypeOff]*_type 

	bad bool 

	next *moduledata
}

func getPCTab(m *moduledata) []byte {
	return m.pctab
}

func (f *FuncInfo) getEntry() uintptr {
	entry, _ := f.datap.textAddr(f.entryOff)
	return entry
}


func findFuncOffsetInModule(pc uintptr, datap *moduledata) (uintptr, bool) {
	if datap == nil {
		return 0, false
	}
	const nsub = uintptr(len(findfuncbucket{}.subbuckets))

	pcOff, ok := datap.textOff(pc)
	if !ok {
		return 0, false
	}

	x := uintptr(pcOff) + datap.text - datap.minpc
	b := x / pcbucketsize
	i := x % pcbucketsize / (pcbucketsize / nsub)

	ffb := (*findfuncbucket)(add(unsafe.Pointer(datap.findfunctab), b*unsafe.Sizeof(findfuncbucket{})))
	idx := ffb.idx + uint32(ffb.subbuckets[i])

	
	for datap.ftab[idx+1].entryoff <= pcOff {
		idx++
	}

	funcoff := uintptr(datap.ftab[idx].funcoff)
	if funcoff == ^uintptr(0) {
		
		
		
		
		return 0, false
	}

	return funcoff, true
}


func (md *moduledata) textOff(pc uintptr) (uint32, bool) {
	res := uint32(pc - md.text)
	if len(md.textsectmap) > 1 {
		for i, sect := range md.textsectmap {
			if sect.baseaddr > pc {
				
				return 0, false
			}
			end := sect.baseaddr + (sect.end - sect.vaddr)
			
			if i == len(md.textsectmap) {
				end++
			}
			if pc < end {
				res = uint32(pc - sect.baseaddr + sect.vaddr)
				break
			}
		}
	}
	return res, true
}


func (md *moduledata) textAddr(off32 uint32) (uintptr, bool) {
	off := uintptr(off32)
	res := md.text + off
	if len(md.textsectmap) > 1 {
		for i, sect := range md.textsectmap {
			
			if off >= sect.vaddr && off < sect.end || (i == len(md.textsectmap)-1 && off == sect.end) {
				res = sect.baseaddr + off - sect.vaddr
				break
			}
		}
		if res > md.etext && runtime.GOARCH != "wasm" { 
			println("runtime: textAddr", hex(res), "out of range", hex(md.text), "-", hex(md.etext))
			return 0, false
		}
	}
	return res, true
}

type hex uint64

func (md *moduledata) GetTypeMap() map[TypeOff]uintptr {
	typemap := make(map[TypeOff]uintptr)
	for k, v := range md.typemap {
		typemap[k] = uintptr(unsafe.Pointer(v))
	}
	return typemap
}
