//go:build go1.16 && !go1.18
// +build go1.16,!go1.18

package module

import (
	"os"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type pclntableInfo struct {
	pcDataTables    [][]byte
	funcDataOffsets []unsafe.Pointer
	pcLine          []byte
	pcFile          []byte
	pcsp            []byte
}

func (m *moduleDataPatcher) getPCDataTable(offset uintptr) []byte {
	if offset == 0 {
		return nil
	}
	return m.origModule.pctab[offset:]
}

func (m *moduleDataPatcher) createPCHeader() {
	m.pcHeader.nfunc = len(m.origModule.ftab)
	m.pcHeader.nfiles = uint(len(m.origModule.filetab))
}

func (m *moduleDataPatcher) funcDataOffset(tableIndex uint8) unsafe.Pointer {
	return funcdata(*m.origFunction, tableIndex)
}


func (m *moduleDataPatcher) getModuleData() (moduledata, error) {
	return moduledata{
		pcHeader:    &m.pcHeader,
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
	return functab{funcoff: funcOff, entry: entry}
}

func (m *moduleDataPatcher) buildFunc(pcspOffset uint32, pcfileOffset uint32, pclnOffset uint32) _func {
	return _func{
		entry:   m.newFuncEntryAddress,
		nameoff: m.origFunction.nameoff,

		args:        m.origFunction.args,
		deferreturn: m.getDeferreturn(),

		pcsp:      pcspOffset,
		pcfile:    pcfileOffset,
		pcln:      pclnOffset,
		npcdata:   uint32(len(m.info.pcDataTables)),
		cuOffset:  m.origFunction.cuOffset,
		funcID:    m.origFunction.funcID,
		nfuncdata: m.origFunction.nfuncdata,
	}
}

func (m *moduleDataPatcher) buildPCFile(patcher *PCDataPatcher) error {
	newPCFile, _, err := patcher.CreatePCFile(decodePCDataEntries(m.getPCDataTable(uintptr(m.origFunction.pcfile))))
	if err != nil {
		return err
	}
	newPCFileBytes, err := encodePCDataEntries(newPCFile)

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.origFunction.pcfile)), "Old pcfile")
		dumpPCData(newPCFileBytes, "New pcfile")
	}

	m.info.pcFile = newPCFileBytes
	return nil
}

const ptrSize = 8

func (m *moduleDataPatcher) createPclnTable() rookoutErrors.RookoutError {
	
	if len(m.newPclntable) == 0 {
		m.newPclntable = append(m.newPclntable, 0)
	}

	pcDataOffsets := make([]uint32, len(m.info.pcDataTables))
	for i, table := range m.info.pcDataTables {
		if table == nil {
			continue
		}
		pcDataOffsets[i] = uint32(len(m.newPclntable))
		m.writeBytesToPclnTable(table)
	}

	pcspOffset := uint32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcsp)
	pclnOffset := uint32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcLine)
	pcfileOffset := uint32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcFile)

	f := m.buildFunc(pcspOffset, pcfileOffset, pclnOffset)
	m.funcOffset = uintptr(len(m.newPclntable))
	
	m.writeObjectToPclnTable(unsafe.Pointer(&f), int(unsafe.Offsetof(f.nfuncdata)+unsafe.Sizeof(f.nfuncdata)))

	if len(pcDataOffsets) > 0 {
		m.writeObjectToPclnTable(unsafe.Pointer(&pcDataOffsets[0]), len(pcDataOffsets)*4)
	}

	if len(m.info.funcDataOffsets) > 0 {
		
		if len(pcDataOffsets)%2 == 0 {
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
