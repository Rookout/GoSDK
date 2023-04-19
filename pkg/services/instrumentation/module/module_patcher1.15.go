//go:build go1.15 && !go1.16
// +build go1.15,!go1.16

package module

import (
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type pclntableInfo struct {
	pcDataTables    [][]byte
	funcDataOffsets []unsafe.Pointer
	pcLine          []byte
	pcFile          []byte
	pcsp            []byte
	files           map[int32][]byte
}

func (m *moduleDataPatcher) getPCDataTable(offset uintptr) []byte {
	return m.origModule.pclntable[offset:]
}


func (m *moduleDataPatcher) getModuleData() (moduledata, error) {
	return moduledata{
		filetab:     m.filetab, 
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

const ptrSize = 8

func (m *moduleDataPatcher) createPCHeader() error {
	return nil
}

func (m *moduleDataPatcher) funcDataOffset(tableIndex uint8) unsafe.Pointer {
	return funcdata(*m.origFunction, tableIndex)
}

func (m *moduleDataPatcher) buildFunc(funcnameOffset int32, pcspOffset int32, pcfileOffset int32, pclnOffset int32) _func {
	return _func{
		entry:   m.newFuncEntryAddress,
		nameoff: funcnameOffset,

		args:        m.origFunction.args,
		deferreturn: m.getDeferreturn(),

		pcsp:      pcspOffset,
		pcfile:    pcfileOffset,
		pcln:      pclnOffset,
		npcdata:   int32(len(m.info.pcDataTables)),
		funcID:    m.origFunction.funcID,
		nfuncdata: m.origFunction.nfuncdata,
	}
}

func (m *moduleDataPatcher) isPCDataStartValid(pcDataStart uint32) bool {
	return int(pcDataStart) < len(m.origModule.pclntable)
}

func (m *moduleDataPatcher) newFuncTab(funcOff uintptr, entry uintptr) functab {
	return functab{funcoff: funcOff, entry: entry}
}

func (m *moduleDataPatcher) buildPCFile() error {
	oldFilenosToNew := make(map[int32]int32)
	pcDataEntries := decodePCDataEntries(m.getPCDataTable(uintptr(m.origFunction.pcfile)))
	i := int32(0)
	for _, entry := range pcDataEntries {
		if _, ok := oldFilenosToNew[entry.Value]; !ok {
			oldFilenosToNew[entry.Value] = i
			i++
		}
	}

	m.info.files = make(map[int32][]byte)
	for oldFileno, newFileno := range oldFilenosToNew {
		m.info.files[newFileno] = []byte(funcfile(*m.origFunction, oldFileno))
	}

	
	if err := updatePCDataEntries(pcDataEntries, m.offsetMappings, false); err != nil {
		return err
	}
	
	for i := range pcDataEntries {
		pcDataEntries[i].Value = oldFilenosToNew[pcDataEntries[i].Value]
	}
	pcFile, err := encodePCDataEntries(pcDataEntries)
	if err != nil {
		return err
	}
	m.info.pcFile = pcFile
	return nil
}

func (m *moduleDataPatcher) createPclnTable() rookoutErrors.RookoutError {
	m.filetab = make([]uint32, len(m.info.files))
	for fileno, filename := range m.info.files {
		m.filetab[fileno] = uint32(len(m.newPclntable))
		m.writeBytesToPclnTable(append(filename, 0))
	}

	
	if len(m.newPclntable) == 0 {
		m.newPclntable = append(m.newPclntable, 0)
	}

	funcnameOffset := int32(len(m.newPclntable))
	m.writeBytesToPclnTable(append([]byte(m.funcName), 0))

	pcDataOffsets := make([]uint32, len(m.info.pcDataTables))
	for i, table := range m.info.pcDataTables {
		if table == nil {
			continue
		}
		pcDataOffsets[i] = uint32(len(m.newPclntable))
		m.writeBytesToPclnTable(table)
	}

	pcspOffset := int32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcsp)
	pclnOffset := int32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcLine)
	pcfileOffset := int32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcFile)

	f := m.buildFunc(funcnameOffset, pcspOffset, pcfileOffset, pclnOffset)
	m.funcOffset = uintptr(len(m.newPclntable))
	
	m.writeObjectToPclnTable(unsafe.Pointer(&f), int(unsafe.Offsetof(f.nfuncdata)+unsafe.Sizeof(f.nfuncdata)))

	if len(pcDataOffsets) > 0 {
		m.writeObjectToPclnTable(unsafe.Pointer(&pcDataOffsets[0]), len(pcDataOffsets)*4)
	}

	if len(m.info.funcDataOffsets) > 0 {
		
		if len(pcDataOffsets)%2 != 0 {
			m.writeBytesToPclnTable(make([]byte, 4))
		}
		m.writeObjectToPclnTable(unsafe.Pointer(&m.info.funcDataOffsets[0]), len(m.info.funcDataOffsets)*ptrSize)
	}

	return m.alignPclntable()
}


func (m *moduleDataPatcher) alignPclntable() rookoutErrors.RookoutError {
	if uintptr(unsafe.Pointer(&m.newPclntable[m.funcOffset]))&4 == 0 {
		return nil
	}

	pclntable := m.newPclntable
	m.newPclntable = make([]byte, uintptr(len(pclntable))+4)
	alignment := uintptr(unsafe.Pointer(&m.newPclntable[m.funcOffset])) & 4
	for i := uintptr(0); i < m.funcOffset; i++ {
		m.newPclntable[i] = pclntable[i]
	}
	for i := m.funcOffset; i < uintptr(len(pclntable)); i++ {
		m.newPclntable[i+alignment] = pclntable[i]
	}
	m.funcOffset += alignment

	if uintptr(unsafe.Pointer(&m.newPclntable[m.funcOffset]))&4 != 0 {
		return rookoutErrors.NewFailedToAlignFunc(
			uintptr(unsafe.Pointer(&m.newPclntable[m.funcOffset])),
			uintptr(unsafe.Pointer(&m.newPclntable[0])),
			m.funcOffset)
	}

	return nil
}
