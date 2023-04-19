//go:build go1.15 && !go1.21
// +build go1.15,!go1.21

package module

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

var (
	modules     = make(map[string]*moduledata)
	modulesLock sync.Mutex
)


type moduleDataPatcher struct {
	info                 pclntableInfo
	funcName             string
	addressMappings      []AddressMapping
	offsetMappings       []AddressMapping
	origFunction         *FuncInfo
	origFuncEntryAddress uintptr
	newFuncEntryAddress  uintptr
	newFuncEndAddress    uintptr
	origModule           *moduledata
	newPclntable         []byte
	
	funcOffset          uintptr
	ftab                []functab
	findFuncBucketTable *findfuncbucket
	name                string
	pcHeader            pcHeader
	filetab             []uint32
}

type AddressMapping struct {
	NewAddress      uintptr
	OriginalAddress uintptr
}

const BPMarker uintptr = 0xffffffffffffffff
const PrologueMarker uintptr = 0xaaaaaaaaaaaaaaaa

var CallbacksMarkers = map[uintptr]interface{}{BPMarker: nil, PrologueMarker: nil}

func FindFuncMaxSPDelta(addr uint64) int32 {
	if f := FindFunc(uintptr(addr)); f._func != nil {
		return funcMaxSPDelta(f)
	}

	return 0
}

func (m *moduleDataPatcher) writeObjectToPclnTable(value unsafe.Pointer, size int) {
	for i := 0; i < size; i++ {
		p := (*byte)(unsafe.Pointer(uintptr(value) + uintptr(i)))
		m.newPclntable = append(m.newPclntable, *p)
	}
}

func (m *moduleDataPatcher) writeBytesToPclnTable(value []byte) {
	m.newPclntable = append(m.newPclntable, value...)
}


func dumpPCData(b []byte, prefix string) {
	var pc uintptr
	val := int32(-1)
	var ok bool
	b, ok = step(b, &pc, &val, true)
	for {
		if !ok {
			fmt.Println(prefix, "step end (ok)")
			break
		}
		fmt.Printf("\tvalue=%d until offset=0x%08x, \n", val, pc)
		if len(b) <= 0 {
			fmt.Println(prefix, "step end (len)")
			break
		}
		b, ok = step(b, &pc, &val, false)
	}
}


func dumpBuffer(start uintptr, end uintptr, prefix string) {
	//goland:noinspection GoVetUnsafePointer
	bufSlice := (*[1 << 28]byte)(unsafe.Pointer(start))[: end-start : end-start]
	fmt.Printf("Start: %s at 0x%016x\n", prefix, start)

	for _, bb := range bufSlice {
		fmt.Printf("%02x ", bb)
	}
	fmt.Printf("\nEnd: %s\n", prefix)
}



const maxBuckets = 100




var findFuncBuckets [100]findfuncbucket


func (m *moduleDataPatcher) createFindFuncBucket() error {
	
	bucketsCount := ((m.newFuncEndAddress - m.newFuncEntryAddress) / pcbucketsize) + 1
	if bucketsCount > maxBuckets {
		return fmt.Errorf("function is %d/%d bytes long, unable to patch moduledata", m.newFuncEndAddress-m.newFuncEntryAddress, maxBuckets*pcbucketsize)
	}

	m.findFuncBucketTable = &(findFuncBuckets[0])
	return nil
}


func (m *moduleDataPatcher) pcDataOffset(table int) (uint32, bool) {
	offset := pcdatastart(*m.origFunction, uint32(table))
	
	
	if offset == 0 {
		return 0, false
	}
	return offset, m.isPCDataStartValid(offset)
}

func (m *moduleDataPatcher) buildFuncData() {
	for tableIndex := uint8(0); tableIndex < m.origFunction.nfuncdata; tableIndex++ {
		m.info.funcDataOffsets = append(m.info.funcDataOffsets, m.funcDataOffset(tableIndex))
	}
}


func (m *moduleDataPatcher) buildPCData() error {
	m.info.pcDataTables = make([][]byte, m.origFunction.npcdata)
	for tableIndex := 0; tableIndex < int(m.origFunction.npcdata); tableIndex++ {
		
		pcDataOffset, ok := m.pcDataOffset(tableIndex)
		if !ok {
			logger.Logger().Debugf("pcDataOffset of table %d is %d, skipping", tableIndex, pcDataOffset)
			continue
		}
		
		newPCData, err := updatePCDataOffsets(m.getPCDataTable(uintptr(pcDataOffset)), m.offsetMappings, nil)
		if err != nil {
			return err
		}
		m.info.pcDataTables[tableIndex] = newPCData

		if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
			dumpPCData(m.getPCDataTable(uintptr(pcDataOffset)), fmt.Sprintf("Old pcdata %d", tableIndex))
			dumpPCData(newPCData, fmt.Sprintf("New pcdata %d", tableIndex))
		}
	}
	err := m.buildPCFile()
	if err != nil {
		return err
	}
	err = m.buildPCLine()
	if err != nil {
		return err
	}
	err = m.buildPCSP()
	if err != nil {
		return err
	}
	return nil
}


func (m *moduleDataPatcher) buildPCSP() error {
	newPCSP, err := updatePCDataOffsets(m.getPCDataTable(uintptr(m.origFunction.pcsp)), m.offsetMappings, func(start uintptr, end uintptr) ([]PCDataEntry, error) {
		return generatePCSP(start+m.newFuncEntryAddress, end+m.newFuncEntryAddress)
	})
	if err != nil {
		return err
	}
	m.info.pcsp = newPCSP

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.origFunction.pcsp)), "Old PCSP")
		dumpPCData(newPCSP, "New PCSP")
	}

	return nil
}


func (m *moduleDataPatcher) buildPCLine() error {
	newPCLine, err := updatePCDataOffsets(m.getPCDataTable(uintptr(m.origFunction.pcln)), m.offsetMappings, nil)
	if err != nil {
		return err
	}
	m.info.pcLine = newPCLine

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.origFunction.pcln)), "Old pcln")
		dumpPCData(newPCLine, "New pcln")
	}

	return nil
}


func (m *moduleDataPatcher) createFuncTable() error {
	
	m.ftab = append(m.ftab, m.newFuncTab(uintptr(len(m.newPclntable)), m.newFuncEntryAddress))
	m.ftab = append(m.ftab, m.newFuncTab(m.funcOffset, m.newFuncEntryAddress))
	
	m.ftab = append(m.ftab, m.newFuncTab(uintptr(len(m.newPclntable)), m.newFuncEndAddress))
	return nil
}

func (m *moduleDataPatcher) getDeferreturn() uint32 {
	if newDeferReturn, ok := findNewAddressByOriginalAddress(uintptr(m.origFunction.deferreturn), m.offsetMappings); ok {
		return uint32(newDeferReturn)
	}
	return 0
}

func addModule(newModule *moduledata) {
	modulesLock.Lock()
	
	modules[newModule.modulename] = newModule

	for datap := &firstModuleData; ; {
		if datap.next == nil {
			datap.next = newModule
			break
		}
		datap = datap.next
	}
	modulesLock.Unlock()
}

func verifyPCDatas(f1, f2 FuncInfo, addressMappings []AddressMapping) rookoutErrors.RookoutError {
	if f1.npcdata != f2.npcdata {
		return rookoutErrors.NewDifferentNPCData(uint32(f1.npcdata), uint32(f2.npcdata))
	}

	prevMapping := addressMappings[0]
	for _, mapping := range addressMappings {
		
		_, ok := CallbacksMarkers[prevMapping.OriginalAddress]
		prevMapping = mapping
		if ok {
			continue
		}
		if _, ok := CallbacksMarkers[mapping.OriginalAddress]; ok {
			continue
		}

		pc1 := mapping.OriginalAddress - 1
		pc2 := mapping.NewAddress - 1

		for i := uint32(0); i < uint32(f1.npcdata); i++ {
			value1 := pcdatavalue1(f1, uint32(i), pc1, nil, false)
			value2 := pcdatavalue1(f2, uint32(i), pc2, nil, false)
			if value1 != value2 {
				return rookoutErrors.NewPCDataVerificationFailed(i, value1, pc1, value2, pc2)
			}
		}

		sp1, _ := pcvalue(f1, uint32(f1.pcsp), pc1, nil, false)
		sp2, _ := pcvalue(f2, uint32(f2.pcsp), pc2, nil, false)
		if sp1 != sp2 {
			return rookoutErrors.NewPCSPVerificationFailed(sp1, pc1, sp2, pc2)
		}

		file1, line1 := funcline1(f1, pc1, false)
		file2, line2 := funcline1(f2, pc2, false)
		if file1 != file2 {
			return rookoutErrors.NewPCFileVerificationFailed(file1, pc1, file2, pc2)
		}
		if line1 != line2 {
			return rookoutErrors.NewPCLineVerificationFailed(line1, pc1, line2, pc2)
		}
	}

	return nil
}

func verifyFuncDatas(f1, f2 FuncInfo) rookoutErrors.RookoutError {
	if f1.nfuncdata != f2.nfuncdata {
		return rookoutErrors.NewDifferentNFuncData(f1.nfuncdata, f2.nfuncdata)
	}

	for i := 0; i < int(f1.nfuncdata); i++ {
		value1 := funcdata(f1, uint8(i))
		value2 := funcdata(f2, uint8(i))
		if value1 != value2 {
			return rookoutErrors.NewFuncDataVerificationFailed(i, uintptr(value1), uintptr(value2))
		}
	}

	return nil
}

func (m *moduleDataPatcher) verifyModule(module *moduledata) (err rookoutErrors.RookoutError) {
	defer func() {
		if r := recover(); r != nil {
			err = rookoutErrors.NewModuleVerificationFailed(r)
		}
	}()

	newFuncInfo := FuncInfo{
		_func: (*_func)(unsafe.Pointer(&m.newPclntable[m.funcOffset])),
		datap: module,
	}
	err = verifyPCDatas(*m.origFunction, newFuncInfo, m.addressMappings)
	if err != nil {
		return err
	}
	err = verifyFuncDatas(*m.origFunction, newFuncInfo)
	if err != nil {
		return err
	}

	return nil
}




func PatchModuleData(addressMappings []AddressMapping, offsetMappings []AddressMapping, stateID int) error {
	function := FindFunc(uintptr(addressMappings[1].OriginalAddress))
	funcName := FuncName(function)
	moduleName := fmt.Sprintf("Rookout-%s[%x]-%d", funcName, function.getEntry(), stateID)
	if _, ok := modules[moduleName]; ok {
		return nil
	}

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		fmt.Printf("Address mapping\n")
		for _, p := range addressMappings {
			fmt.Printf("0x%016x:0x%016x\n", p.OriginalAddress, p.NewAddress)
		}
		fmt.Printf("Address offsets\n")
		for _, p := range offsetMappings {
			fmt.Printf("0x%016x:0x%016x\n", p.OriginalAddress, p.NewAddress)
		}
	}

	data := &moduleDataPatcher{origFunction: &function, name: moduleName, funcName: funcName}
	data.addressMappings, data.offsetMappings = addressMappings, offsetMappings
	data.origFuncEntryAddress = function.getEntry()
	data.newFuncEntryAddress = data.addressMappings[0].NewAddress
	data.newFuncEndAddress = data.addressMappings[len(data.addressMappings)-1].NewAddress
	data.origModule = function.datap

	
	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpBuffer(data.newFuncEntryAddress, data.newFuncEndAddress, "New function")
		oldFuncEnd := data.origFuncEntryAddress + data.offsetMappings[len(data.offsetMappings)-1].OriginalAddress
		dumpBuffer(data.origFuncEntryAddress, oldFuncEnd, "Original function")
	}

	err := data.buildPCData()
	if err != nil {
		return err
	}
	data.buildFuncData()
	if err = data.createPclnTable(); err != nil {
		return err
	}
	data.createPCHeader()
	if err = data.createFindFuncBucket(); err != nil {
		return err
	}
	if err = data.createFuncTable(); err != nil {
		return err
	}

	module, err := data.getModuleData()
	if err != nil {
		return err
	}

	err = data.verifyModule(&module)
	if err != nil {
		return err
	}

	addModule(&module)
	return nil
}
