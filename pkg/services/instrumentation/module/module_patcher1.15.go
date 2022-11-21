//go:build go1.15 && !go1.16
// +build go1.15,!go1.16

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
}

func (m *moduleDataPatcher) getPCDataTable(offset uintptr) []byte {
	return m.newPclntable[offset:]
}


func (m *moduleDataPatcher) getModuleData() (moduledata, error) {
	if !(m.state.pcDataPatched && m.state.pcLinePatched && m.state.funcAddressPatched && m.state.pcspPatched &&
		m.state.funcTableCreated && m.state.findFuncBucketCreated && m.state.pcFilePatched) {
		return moduledata{}, errors.New("must fully patch module before creating module data")
	}
	return moduledata{
		filetab:     m.origModule.filetab, 
		ftab:        m.ftab,
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

		modulename: m.name, 

		hasmain: 0, 

		gcdatamask: m.origModule.gcdatamask,
		gcbssmask:  m.origModule.gcbssmask,

		typemap: m.origModule.typemap,

		bad: false, 

		next: nil, 
	}, nil
}

func (m *moduleDataPatcher) createPCHeader() error {
	return nil
}

func (m *moduleDataPatcher) isPCDataStartValid(pcDataStart uint32) bool {
	return int(pcDataStart) < len(m.origModule.pclntable)
}

func (m *moduleDataPatcher) newFuncTab(funcOff uintptr, entry uintptr) functab {
	return functab{funcoff: funcOff, entry: entry}
}


func (m *moduleDataPatcher) patchFuncAddress() error {
	if m.state.funcAddressPatched {
		return errors.New("Attempted to patch the func address twice")
	}

	
	funcOffsetEntryPointer := unsafe.Pointer(&m.newPclntable[m.funcOffset])
	patchUInt64WithPointer(funcOffsetEntryPointer, uint64(m.newFuncEntryAddress))

	m.state.funcAddressPatched = true
	return nil
}

func validateModuleFtab(module *moduledata, newFuncEntry uintptr) error {
	moduleName := module.modulename
	if len(module.ftab) != expectedFtabSize {
		return fmt.Errorf("expected exactly %d functions in the module %s ftab. The first and last are dummy values. Got %d instead", expectedFtabSize, moduleName, len(module.ftab))
	}
	if module.ftab[0].entry != module.minpc {
		return fmt.Errorf("The first dummy function should have the same pc as the module %s minpc. Got %d expected %d", moduleName, module.ftab[0].entry, module.minpc)
	}
	if module.ftab[len(module.ftab)-1].entry != module.maxpc {
		return fmt.Errorf("The last dummy function should have the same pc as the module %s max. Got %d expected %d", moduleName, module.ftab[len(module.ftab)-1].entry, module.maxpc)
	}
	patchedFuncTab := module.ftab[patchedIdx]
	if patchedFuncTab.entry != newFuncEntry {
		return fmt.Errorf("Bad patched function entry address in module %s. Expected %d, got %d", moduleName, newFuncEntry, patchedFuncTab.entry)
	}
	patchedOffset := int(patchedFuncTab.funcoff)
	if patchedOffset >= len(module.pclntable) {
		return fmt.Errorf("Patched function offset (%d) outside of module %s pclntable (len=%d)", patchedOffset, moduleName, len(module.pclntable))
	}
	
	return nil
}
