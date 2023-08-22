//go:build go1.21 && !go1.22
// +build go1.21,!go1.22

package module

import (
	"os"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type pclntableInfo struct {
	pcDataTables    [][]byte
	funcDataOffsets []uint32
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

func (m *moduleDataPatcher) funcDataOffset(table uint8) uint32 {
	p := uintptr(unsafe.Pointer(&m.origFunction.nfuncdata)) + unsafe.Sizeof(m.origFunction.nfuncdata) + uintptr(m.origFunction.npcdata)*4 + uintptr(table)*4
	return *(*uint32)(unsafe.Pointer(p))
}

func (m *moduleDataPatcher) buildFunc(pcspOffset uint32, pcfileOffset uint32, pclnOffset uint32) _func {
	return _func{
		entryoff: 0,
		nameoff:  m.origFunction.nameoff,

		args:        m.origFunction.args,
		deferreturn: m.getDeferreturn(),

		pcsp:      pcspOffset,
		pcfile:    pcfileOffset,
		pcln:      pclnOffset,
		npcdata:   uint32(len(m.info.pcDataTables)),
		cuOffset:  m.origFunction.cuOffset,
		funcID:    m.origFunction.funcID,
		flag:      m.origFunction.flag,
		nfuncdata: uint8(len(m.info.funcDataOffsets)),
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


func (m *moduleDataPatcher) createPCHeader() {
	m.pcHeader.nfunc = len(m.origModule.ftab)
	m.pcHeader.nfiles = (uint)(len(m.origModule.filetab))
	m.pcHeader.textStart = m.newFuncEntryAddress
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
		gofunc:      m.origModule.gofunc,
		rodata:      m.origModule.noptrdata,

		modulename: m.name,

		hasmain: 0, 

		gcdatamask: m.origModule.gcdatamask,
		gcbssmask:  m.origModule.gcbssmask,

		typemap:   m.origModule.typemap,
		inittasks: m.origModule.inittasks,

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

func (m *moduleDataPatcher) createPclnTable() rookoutErrors.RookoutError {
	
	m.newPclntable = make([]byte, 1)

	pcDataOffsets := make([]uint32, len(m.info.pcDataTables))
	for i, table := range m.info.pcDataTables {
		if table == nil {
			continue
		}
		pcDataOffsets[i] = uint32(len(m.newPclntable))
		m.writeObjectToPclnTable(unsafe.Pointer(&table[0]), len(table))
	}

	pcspOffset := uint32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcsp)
	pcfileOffset := uint32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcFile)
	pclnOffset := uint32(len(m.newPclntable))
	m.writeBytesToPclnTable(m.info.pcLine)

	f := m.buildFunc(pcspOffset, pcfileOffset, pclnOffset)
	m.funcOffset = uintptr(len(m.newPclntable))
	
	m.writeObjectToPclnTable(unsafe.Pointer(&f), int(unsafe.Offsetof(f.nfuncdata)+unsafe.Sizeof(f.nfuncdata)))

	if len(pcDataOffsets) > 0 {
		m.writeObjectToPclnTable(unsafe.Pointer(&pcDataOffsets[0]), len(pcDataOffsets)*4)
	}

	if len(m.info.funcDataOffsets) > 0 {
		m.writeObjectToPclnTable(unsafe.Pointer(&m.info.funcDataOffsets[0]), len(m.info.funcDataOffsets)*4)
	}

	return nil
}
