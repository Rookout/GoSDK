//go:build go1.15 && !go1.22
// +build go1.15,!go1.22

package module

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)





const (
	_PCDATA_UnsafePoint   = 0
	_PCDATA_StackMapIndex = 1
	_PCDATA_InlTreeIndex  = 2
	_PCDATA_ArgLiveIndex  = 3

	_FUNCDATA_ArgsPointerMaps    = 0
	_FUNCDATA_LocalsPointerMaps  = 1
	_FUNCDATA_StackObjects       = 2
	_FUNCDATA_InlTree            = 3
	_FUNCDATA_OpenCodedDeferInfo = 4
	_FUNCDATA_ArgInfo            = 5
	_FUNCDATA_ArgLiveInfo        = 6
	_FUNCDATA_WrapInfo           = 7

	_ArgsSizeUnknown = -0x80000000
)

const (
	
	_PCDATA_UnsafePointSafe   = -1 
	_PCDATA_UnsafePointUnsafe = -2 

	
	
	
	
	
	_PCDATA_Restart1 = -3
	_PCDATA_Restart2 = -4

	
	
	_PCDATA_RestartAtEntry = -5
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
	if len(b) == 0 {
		fmt.Println(prefix, "table missing")
		return
	}
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


func (m *moduleDataPatcher) pcDataOffset(table int) uint32 {
	if table >= int(m.origFunction.npcdata) {
		return 0
	}
	offset := pcdatastart(*m.origFunction, uint32(table))
	if m.isPCDataStartValid(offset) {
		return offset
	}
	return 0
}

func (m *moduleDataPatcher) buildFuncData() {
	for tableIndex := uint8(0); tableIndex < m.origFunction.nfuncdata; tableIndex++ {
		m.info.funcDataOffsets = append(m.info.funcDataOffsets, m.funcDataOffset(tableIndex))
	}
}

func (m *moduleDataPatcher) hasPatchedCode() bool {
	for _, mapping := range m.offsetMappings {
		if mapping.OriginalAddress == BPMarker || mapping.OriginalAddress == PrologueMarker {
			return true
		}
	}
	return false
}


func (m *moduleDataPatcher) buildPCData() error {
	
	numTablesInPatched := int(m.origFunction.npcdata)
	isPatched := m.hasPatchedCode()
	if isPatched && numTablesInPatched <= _PCDATA_UnsafePoint {
		numTablesInPatched = _PCDATA_UnsafePoint + 1 
	}
	m.info.pcDataTables = make([][]byte, numTablesInPatched)
	pcDataPatcher, err := NewPCDataPatcher(m.newFuncEntryAddress, m.offsetMappings, isPatched, instructionSizeBytes)
	if err != nil {
		return err
	}

	for tableIndex := 0; tableIndex < numTablesInPatched; tableIndex++ {
		
		pcDataOffset := m.pcDataOffset(tableIndex)
		newPCDataTable, err := pcDataPatcher.CreatePCData(tableIndex, decodePCDataEntries(m.getPCDataTable(uintptr(pcDataOffset))))
		if err != nil {
			return err
		}
		
		newPCData, err := encodePCDataEntries(newPCDataTable)
		if err != nil {
			return err
		}
		m.info.pcDataTables[tableIndex] = newPCData

		if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
			dumpPCData(m.getPCDataTable(uintptr(pcDataOffset)), fmt.Sprintf("Old pcdata %d", tableIndex))
			dumpPCData(newPCData, fmt.Sprintf("New pcdata %d", tableIndex))
		}
	}
	err = m.buildPCFile(pcDataPatcher)
	if err != nil {
		return err
	}
	err = m.buildPCLine(pcDataPatcher)
	if err != nil {
		return err
	}
	err = m.buildPCSP(pcDataPatcher)
	if err != nil {
		return err
	}
	return nil
}


func (m *moduleDataPatcher) buildPCSP(patcher *PCDataPatcher) error {
	newPCSP, err := patcher.CreatePCSP(decodePCDataEntries(m.getPCDataTable(uintptr(m.origFunction.pcsp))))
	if err != nil {
		return err
	}
	pcspBytes, err := encodePCDataEntries(newPCSP)
	if err != nil {
		return err
	}
	m.info.pcsp = pcspBytes

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.origFunction.pcsp)), "Old PCSP")
		dumpPCData(pcspBytes, "New PCSP")
	}

	return nil
}


func (m *moduleDataPatcher) buildPCLine(patcher *PCDataPatcher) error {

	newPCLine, err := patcher.CreatePCLine(decodePCDataEntries(m.getPCDataTable(uintptr(m.origFunction.pcln))))
	if err != nil {
		return err
	}
	newPCLineBytes, err := encodePCDataEntries(newPCLine)
	m.info.pcLine = newPCLineBytes

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.origFunction.pcln)), "Old pcln")
		dumpPCData(newPCLineBytes, "New pcln")
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

type callbackVerificationInfo struct {
	startAddress    uintptr
	endAddress      uintptr
	origPCDataValue []int32
	origPCSP        int32
	origLine        int32
	origFile        string
	origPC          uintptr
}









func verifySingleCallbackPCDatas(newFunc FuncInfo, cbInfo callbackVerificationInfo) rookoutErrors.RookoutError {
	expectedPCSP, _ := generatePCSP(cbInfo.startAddress, cbInfo.endAddress)
	for i := 0; i < len(expectedPCSP); i++ {
		expectedPCSP[i].Value += cbInfo.origPCSP
		expectedPCSP[i].Offset += cbInfo.startAddress
	}

	for newPC := cbInfo.startAddress; newPC < cbInfo.endAddress; {
		for pcdataIdx, origValue := range cbInfo.origPCDataValue {
			newValue := pcdatavalue1(newFunc, uint32(pcdataIdx), newPC, nil, false)
			if newPC == cbInfo.startAddress || pcdataIdx != _PCDATA_UnsafePoint {
				if origValue != newValue {
					return rookoutErrors.NewPCDataVerificationFailed(uint32(pcdataIdx), origValue, cbInfo.origPC, newValue, newPC)
				}
			} else {
				
				if newValue != _PCDATA_UnsafePointUnsafe {
					return rookoutErrors.NewPCDataAsyncUnsafePointVerificationFailed(newValue, newPC)
				}
			}
		}
		newFile, newLine := funcline1(newFunc, newPC, false)
		if cbInfo.origFile != newFile {
			return rookoutErrors.NewPCFileVerificationFailed(cbInfo.origFile, cbInfo.origPC, newFile, newPC)
		}
		if cbInfo.origLine != newLine {
			return rookoutErrors.NewPCLineVerificationFailed(cbInfo.origLine, cbInfo.origPC, newLine, newPC)
		}

		foundPCSP := false
		for _, pcspEntry := range expectedPCSP {
			if newPC < pcspEntry.Offset {
				newSP, _ := pcvalue(newFunc, uint32(newFunc.pcsp), newPC, nil, false)
				if pcspEntry.Value != newSP {
					return rookoutErrors.NewPCSPInPatchedVerificationFailed(cbInfo.origPCSP, cbInfo.origPC, pcspEntry.Value, newSP, newPC)
				}
				foundPCSP = true
				break
			}
		}
		if !foundPCSP {
			return rookoutErrors.NewPCSPVerificationFailedMissingEntry(cbInfo.origPCSP, cbInfo.origPC, newPC)
		}
		currInstSize, err := instructionSizeBytes(newPC)
		if err != nil {
			return err
		}
		newPC += currInstSize
	}
	return nil
}

func verifyCallbacksPCDatas(nPCData uint32, f1, f2 FuncInfo, addressMappings []AddressMapping) rookoutErrors.RookoutError {
	for mapIdx, mapping := range addressMappings {
		pc1 := mapping.OriginalAddress
		pc2 := mapping.NewAddress
		if _, ok := CallbacksMarkers[pc1]; !ok {
			continue
		}
		cbInfo := callbackVerificationInfo{
			startAddress:    pc2,
			endAddress:      addressMappings[mapIdx+1].NewAddress,
			origPCDataValue: make([]int32, nPCData),
		}
		
		for _, nextMapping := range addressMappings[mapIdx+1:] {
			if _, ok := CallbacksMarkers[nextMapping.OriginalAddress]; !ok {
				pc1 = nextMapping.OriginalAddress
				cbInfo.origPC = pc1
				break
			}
		}

		for i := uint32(0); i < nPCData; i++ {
			cbInfo.origPCDataValue[i] = pcdatavalue1(f1, i, pc1, nil, false)
		}

		sp1, _ := pcvalue(f1, uint32(f1.pcsp), pc1, nil, false)
		file1, line1 := funcline1(f1, pc1, false)

		cbInfo.origPCSP = sp1
		cbInfo.origLine = line1
		cbInfo.origFile = file1
		if err := verifySingleCallbackPCDatas(f2, cbInfo); err != nil {
			return err
		}
	}
	return nil
}

func verifyNonCallbackPCDatas(nPCData uint32, f1, f2 FuncInfo, addressMappings []AddressMapping) rookoutErrors.RookoutError {
	for _, mapping := range addressMappings {
		pc1 := mapping.OriginalAddress
		pc2 := mapping.NewAddress
		if _, ok := CallbacksMarkers[pc1]; ok {
			continue
		}

		for i := uint32(0); i < nPCData; i++ {
			value1 := pcdatavalue1(f1, i, pc1, nil, false)
			value2 := pcdatavalue1(f2, i, pc2, nil, false)
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

func verifyPCDatas(f1, f2 FuncInfo, addressMappings []AddressMapping) rookoutErrors.RookoutError {
	if f1.npcdata != f2.npcdata {
		
		if f1.npcdata > _PCDATA_UnsafePoint || f2.npcdata != _PCDATA_UnsafePoint+1 {
			
			return rookoutErrors.NewDifferentNPCData(uint32(f1.npcdata), uint32(f2.npcdata))
		}

	}
	nPCData := f1.npcdata
	if f2.npcdata > nPCData {
		nPCData = f2.npcdata
	}
	if err := verifyNonCallbackPCDatas(uint32(nPCData), f1, f2, addressMappings); err != nil {
		return err
	}
	if err := verifyCallbacksPCDatas(uint32(nPCData), f1, f2, addressMappings); err != nil {
		return err
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
