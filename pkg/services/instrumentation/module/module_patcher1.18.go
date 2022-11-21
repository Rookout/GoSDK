//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

package module

import (
	"fmt"
	"unsafe"

	"github.com/go-errors/errors"
)



type moduleDataPatcherState struct {
	funcAddressPatched    bool
	pcspPatched           bool
	pcFilePatched         bool
	pcLinePatched         bool
	pcDataPatched         bool
	findFuncBucketCreated bool
	funcTableCreated      bool
	pcHeaderCreated       bool
}


type moduleDataPatcher struct {
	state                moduleDataPatcherState
	addressMappings      []AddressMapping
	offsetMappings       []AddressMapping
	function             *FuncInfo
	origFuncEntryAddress uintptr
	newFuncEntryAddress  uintptr
	newFuncEndAddress    uintptr
	origModule           *moduledata
	newPclntable         []byte
	
	funcOffset          uintptr
	ftab                []functab
	findFuncBucketTable *findfuncbucket
	name                string
	pcHeader            *pcHeader
}

func (m *moduleDataPatcher) getPCDataTable(offset uintptr) []byte {
	return m.origModule.pctab[offset:]
}


func (m *moduleDataPatcher) createPCHeader() error {
	if m.state.pcHeaderCreated {
		return errors.New("Attempted to create the PCHeader twice.")
	}

	m.pcHeader = (*pcHeader)(unsafe.Pointer(&(m.newPclntable[0])))
	m.pcHeader.nfunc = len(m.origModule.ftab)
	m.pcHeader.nfiles = (uint)(len(m.origModule.filetab))
	m.pcHeader.textStart = m.newFuncEntryAddress

	m.state.pcHeaderCreated = true
	return nil
}


func (m *moduleDataPatcher) getModuleData() (moduledata, error) {
	if !(m.state.pcDataPatched && m.state.pcLinePatched && m.state.funcAddressPatched && m.state.pcspPatched &&
		m.state.funcTableCreated && m.state.pcHeaderCreated && m.state.findFuncBucketCreated && m.state.pcFilePatched) {
		return moduledata{}, errors.New("must fully patch module before creating module data")
	}
	return moduledata{
		pcHeader:    m.pcHeader,
		funcnametab: m.origModule.funcnametab, 
		filetab:     m.origModule.filetab,     
		cutab:       m.origModule.cutab,       
		ftab:        m.ftab,
		pctab:       m.newPclntable,                                 
		pclntable:   m.newPclntable,                                 
		findfunctab: uintptr(unsafe.Pointer(m.findFuncBucketTable)), 
		minpc:       m.newFuncEntryAddress,                          
		maxpc:       m.newFuncEndAddress,
		text:        m.newFuncEntryAddress,   
		etext:       m.newFuncEndAddress,     
		noptrdata:   m.origModule.noptrdata,  
		enoptrdata:  m.origModule.enoptrdata, 
		data:        m.origModule.data,       
		edata:       m.origModule.edata,      
		bss:         m.origModule.bss,        
		ebss:        m.origModule.ebss,       
		noptrbss:    m.origModule.noptrbss,   
		enoptrbss:   m.origModule.enoptrbss,  
		end:         m.origModule.end,
		gcdata:      m.origModule.gcdata,
		gcbss:       m.origModule.gcbss,
		types:       m.origModule.types,  
		etypes:      m.origModule.etypes, 
		gofunc:      m.origModule.gofunc,
		rodata:      m.origModule.noptrdata,

		modulename: m.name, 

		hasmain: 0, 

		gcdatamask: m.origModule.gcdatamask,
		gcbssmask:  m.origModule.gcbssmask,

		typemap: m.origModule.typemap,

		bad: false, 

		next: nil, 
	}, nil
}

func (m *moduleDataPatcher) isPCDataStartValid(pcDataStart uint32) bool {
	return int(pcDataStart) < len(m.origModule.pctab)
}

func (m *moduleDataPatcher) newFuncTab(funcOff uintptr, entry uintptr) functab {
	entryOff := entry - m.newFuncEntryAddress
	return functab{funcoff: uint32(funcOff), entryoff: uint32(entryOff)}
}


func (m *moduleDataPatcher) patchFuncAddress() error {
	if m.state.funcAddressPatched {
		return errors.New("Attempted to patch the func address twice")
	}

	
	funcOffsetEntryPointer := unsafe.Pointer(&m.newPclntable[m.funcOffset])
	patchUInt32WithPointer(funcOffsetEntryPointer, 0)

	m.state.funcAddressPatched = true
	return nil
}

func validateModuleFtab(module *moduledata, _ uintptr) error {
	moduleName := module.modulename
	if len(module.ftab) != expectedFtabSize {
		return fmt.Errorf("expected exactly %d functions in the module %s ftab. The first and last are dummy values. Got %d instead", expectedFtabSize, moduleName, len(module.ftab))
	}
	if module.ftab[0].entryoff != 0 {
		return fmt.Errorf("expected entryoff of ftab[0] in %s to be 0, got %d", moduleName, module.ftab[0].entryoff)
	}
	if module.ftab[len(module.ftab)-1].entryoff != uint32(module.maxpc-module.minpc) {
		return fmt.Errorf("expected entryoff of ftab[-1] in %s to be %d, got %d", moduleName, uint32(module.maxpc-module.minpc), module.ftab[len(module.ftab)-1])
	}
	patchedFuncTab := module.ftab[patchedIdx]
	if patchedFuncTab.entryoff != 0 {
		return fmt.Errorf("expected entryoff of patched function tab in %s to be 0, got %d", moduleName, patchedFuncTab.entryoff)
	}
	patchedOffset := int(patchedFuncTab.funcoff)
	if patchedOffset >= len(module.pclntable) {
		return fmt.Errorf("patched function offset (%d) outside of module %s pclntable (len=%d)", patchedOffset, moduleName, len(module.pclntable))
	}
	
	return nil
}
