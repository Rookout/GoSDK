//go:build go1.15 && !go1.20
// +build go1.15,!go1.20

package module

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/logger"

	"github.com/go-errors/errors"
)

var (
	modules     = make(map[string]*moduledata)
	modulesLock sync.Mutex
)

type AddressMapping struct {
	NewAddress      uintptr
	OriginalAddress uintptr
}

type stackUsageInfo struct {
	offset    uintptr 
	valueDiff int32
}

type PCSPNativeInfo struct {
	BpOpcodesSizeInBytes          int
	BpStackUsage                  int32
	PrologueAfterUsingStackOffset int
	PrologueStackUsage            int32
}

var stackUsageMap map[uintptr][]stackUsageInfo

func loadStackUsageMap(unparsedStackUsageMap map[uint64][]map[string]int64) {
	valueDiffKey := "valueDiff"
	offsetKey := "offset"
	stackUsageMap = make(map[uintptr][]stackUsageInfo)
	for marker, m := range unparsedStackUsageMap {
		for _, info := range m {
			stackUsageMap[uintptr(marker)] = append(stackUsageMap[uintptr(marker)], stackUsageInfo{
				offset:    uintptr(info[offsetKey]),
				valueDiff: int32(info[valueDiffKey]),
			})
		}
	}
}

const BPMarker uintptr = 0xffffffffffffffff
const PrologueMarker uintptr = 0xaaaaaaaaaaaaaaaa

var callbacksMarkers = [...]uintptr{BPMarker, PrologueMarker}

func FindFuncMaxSPDelta(addr uint64) int32 {
	if f := FindFunc(uintptr(addr)); f._func != nil {
		return funcMaxSPDelta(f)
	}

	return 0
}

func patchUInt32WithPointer(ptr unsafe.Pointer, value uint32) {
	*(*uint32)(ptr) = value
}

func patchUInt64WithPointer(ptr unsafe.Pointer, value uint64) {
	*(*uint64)(ptr) = value
}



func cArrayToUint64Slice(arrayPointer unsafe.Pointer, count int) []uint64 {
	return (*[1 << 28]uint64)(arrayPointer)[:count:count]
}




func bufferToAddressMapping(addressMappingsBufferPointer unsafe.Pointer, origFuncAddress uintptr) (addressMappings []AddressMapping,
	offsetMappings []AddressMapping) {
	addressMappingsCount := *(*uint64)(addressMappingsBufferPointer)
	
	addressMappingsBufferPointer = unsafe.Pointer(uintptr(addressMappingsBufferPointer) + unsafe.Sizeof(uint64(0)))
	addressMappingsSlice := cArrayToUint64Slice(addressMappingsBufferPointer, int(addressMappingsCount)*2)

	newFuncAddress := uintptr(addressMappingsSlice[0])

	for i := 0; i < len(addressMappingsSlice); i += 2 {
		newAddress := uintptr(addressMappingsSlice[i])
		origAddress := uintptr(addressMappingsSlice[i+1])
		newOffset := newAddress - newFuncAddress
		origOffset := origAddress - origFuncAddress

		
		
		for _, callbackMarker := range callbacksMarkers {
			if origAddress == callbackMarker {
				origOffset = origAddress
			}
		}

		addressMappings = append(addressMappings, AddressMapping{newAddress, origAddress})
		offsetMappings = append(offsetMappings, AddressMapping{newOffset, origOffset})
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
	return addressMappings, offsetMappings
}


func (m *moduleDataPatcher) addBufferToPclntable(buffer *[]byte) uint32 {
	newOffset := uint32(len(m.newPclntable))
	m.newPclntable = append(m.newPclntable, *buffer...)
	return newOffset
}



func (m *moduleDataPatcher) patchPClntableEntryUInt32(newEntry *[]byte, offsetEntry uintptr) uint32 {
	newOffset := m.addBufferToPclntable(newEntry)
	patchUInt32WithPointer(unsafe.Pointer(&m.newPclntable[offsetEntry]), newOffset)
	return newOffset
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
	if m.state.findFuncBucketCreated {
		return errors.New("Attempted to create findfuncbucket twice.")
	}

	
	bucketsCount := ((m.newFuncEndAddress - m.newFuncEntryAddress) / pcbucketsize) + 1
	if bucketsCount > maxBuckets {
		return fmt.Errorf("function is %d/%d bytes long, unable to patch moduledata", m.newFuncEndAddress-m.newFuncEntryAddress, maxBuckets*pcbucketsize)
	}

	m.findFuncBucketTable = &(findFuncBuckets[0])
	m.state.findFuncBucketCreated = true
	return nil
}


func (m *moduleDataPatcher) pcDataOffset(table int) (uint32, bool) {
	offset := pcdatastart(*m.function, uint32(table))
	
	
	if offset == 0 {
		return 0, false
	}
	return offset, m.isPCDataStartValid(offset)
}


func (m *moduleDataPatcher) pcDataOffsetOffset(table int) uintptr {
	f := _func{}
	offset := m.funcOffset + unsafe.Offsetof(f.nfuncdata) + unsafe.Sizeof(f.nfuncdata) + uintptr(table)*4
	return offset
}

type pcDataTableInfo struct {
	offset       uint32
	data         []byte
	lastPCDataPC uintptr
}

type allPCDataPatchInfo struct {
	pcdataInfo []*pcDataTableInfo
	pcFileInfo *pcDataTableInfo
	pcLineInfo *pcDataTableInfo
	pcSPInfo   *pcDataTableInfo
}


func (m *moduleDataPatcher) patchPCData() (*allPCDataPatchInfo, error) {
	if m.state.pcDataPatched {
		return nil, errors.New("Attempted to patch the pcData twice.")
	}
	info := &allPCDataPatchInfo{pcdataInfo: make([]*pcDataTableInfo, m.function.npcdata)}
	for tableIndex := 0; tableIndex < int(m.function.npcdata); tableIndex++ {
		info.pcdataInfo[tableIndex] = nil
		
		pcDataOffset, ok := m.pcDataOffset(tableIndex)
		if !ok {
			logger.Logger().Debugf("pcDataOffset of table %d is %d, skipping", tableIndex, pcDataOffset)
			continue
		}
		
		pcDataOffsetOffset := m.pcDataOffsetOffset(tableIndex)
		
		newPCData, lastNewValidPCOffset, err := updatePCDataOffsets(m.getPCDataTable(uintptr(pcDataOffset)), m.offsetMappings, nil)
		if err != nil {
			return nil, err
		}
		lastNewValidPC := lastNewValidPCOffset + m.newFuncEntryAddress
		
		newPCDataOffset := m.addBufferToPclntable(&newPCData)
		//goland:noinspection GoVetUnsafePointer
		info.pcdataInfo[tableIndex] = &pcDataTableInfo{
			offset:       newPCDataOffset,
			data:         newPCData,
			lastPCDataPC: lastNewValidPC,
		}
		
		patchUInt32WithPointer(unsafe.Pointer(&m.newPclntable[pcDataOffsetOffset]), newPCDataOffset)

		if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
			dumpPCData(m.getPCDataTable(uintptr(pcDataOffset)), fmt.Sprintf("Old pcdata %d", tableIndex))
			dumpPCData(m.newPclntable[newPCDataOffset:], fmt.Sprintf("New pcdata %d", tableIndex))
		}
	}
	var err error
	info.pcFileInfo, err = m.patchPCFile()
	if err != nil {
		return nil, err
	}
	info.pcLineInfo, err = m.patchPCLine()
	if err != nil {
		return nil, err
	}
	info.pcSPInfo, err = m.patchPCSP()
	if err != nil {
		return nil, err
	}
	m.state.pcDataPatched = true
	return info, nil
}


func (m *moduleDataPatcher) patchPCSP() (*pcDataTableInfo, error) {
	if m.state.pcspPatched {
		return nil, errors.New("Attempted to patch the pcsp twice")
	}

	newPCSP, lastNewValidPCOffset, err := updatePCDataOffsets(m.getPCDataTable(uintptr(m.function.pcsp)), m.offsetMappings, stackUsageMap)
	if err != nil {
		return nil, err
	}
	lastNewValidPC := lastNewValidPCOffset + m.newFuncEntryAddress
	f := m.function._func
	pcspEntryOffset := m.funcOffset + unsafe.Offsetof(f.pcsp)
	pcspOffset := m.patchPClntableEntryUInt32(&newPCSP, pcspEntryOffset)

	m.state.pcspPatched = true

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.function.pcsp)), "Old PCSP")
		dumpPCData(newPCSP, "New PCSP")
	}

	return &pcDataTableInfo{
		offset:       pcspOffset,
		data:         newPCSP,
		lastPCDataPC: lastNewValidPC,
	}, nil
}


func (m *moduleDataPatcher) patchPCFile() (*pcDataTableInfo, error) {
	if m.state.pcFilePatched {
		return nil, errors.New("Attempted to patch the pcFile twice.")
	}

	newPCFile, lastNewValidPCOffset, err := updatePCDataOffsets(m.getPCDataTable(uintptr(m.function.pcfile)), m.offsetMappings, nil)
	if err != nil {
		return nil, err
	}
	lastNewValidPC := lastNewValidPCOffset + m.newFuncEntryAddress

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.function.pcfile)), "Old pcfile")
		dumpPCData(newPCFile, "New pcfile")
	}

	f := m.function._func
	pcFileEntryOffset := m.funcOffset + unsafe.Offsetof(f.pcfile)
	pcFileOffset := m.patchPClntableEntryUInt32(&newPCFile, pcFileEntryOffset)

	m.state.pcFilePatched = true
	return &pcDataTableInfo{
		offset:       pcFileOffset,
		data:         newPCFile,
		lastPCDataPC: lastNewValidPC,
	}, nil
}


func (m *moduleDataPatcher) patchPCLine() (*pcDataTableInfo, error) {
	if m.state.pcLinePatched {
		return nil, errors.New("Attempted to patch the pcLine twice.")
	}

	newPCLine, lastNewValidPCOffset, err := updatePCDataOffsets(m.getPCDataTable(uintptr(m.function.pcln)), m.offsetMappings, nil)
	if err != nil {
		return nil, err
	}
	lastNewValidPC := lastNewValidPCOffset + m.newFuncEntryAddress

	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpPCData(m.getPCDataTable(uintptr(m.function.pcln)), "Old pcln")
		dumpPCData(newPCLine, "New pcln")
	}

	f := m.function._func
	pcLineEntryOffset := m.funcOffset + unsafe.Offsetof(f.pcln)
	pcLineOffset := m.patchPClntableEntryUInt32(&newPCLine, pcLineEntryOffset)

	m.state.pcLinePatched = true
	return &pcDataTableInfo{
		offset:       pcLineOffset,
		data:         newPCLine,
		lastPCDataPC: lastNewValidPC,
	}, nil
}


func (m *moduleDataPatcher) createFuncTable() error {
	if m.state.funcTableCreated {
		return errors.New("Attempted to create ftab twice.")
	}

	
	m.ftab = append(m.ftab, m.newFuncTab(uintptr(len(m.newPclntable)), m.newFuncEntryAddress))
	m.ftab = append(m.ftab, m.newFuncTab(m.funcOffset, m.newFuncEntryAddress))
	
	m.ftab = append(m.ftab, m.newFuncTab(uintptr(len(m.newPclntable)), m.newFuncEndAddress)) 

	m.state.funcTableCreated = true
	return nil
}

func (m *moduleDataPatcher) patchDeferReturn() error {
	f := (*_func)(unsafe.Pointer(&m.newPclntable[m.funcOffset]))
	if newDeferReturn, ok := findNewAddressByOriginalAddress(uintptr(f.deferreturn), m.offsetMappings); ok {
		f.deferreturn = uint32(newDeferReturn)
	}
	return nil
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




func PatchModuleData(addr uint64, rawAddressMapping unsafe.Pointer, stateId int) error {
	function := FindFunc(uintptr(addr)) 
	moduleName := "Rookout-" + funcName(function) + "-" + strconv.Itoa(stateId)
	if _, ok := modules[moduleName]; ok {
		return nil
	}

	data := &moduleDataPatcher{function: &function, name: moduleName}
	data.addressMappings, data.offsetMappings = bufferToAddressMapping(rawAddressMapping, function.getEntry())
	data.origFuncEntryAddress = function.getEntry()
	data.newFuncEntryAddress = data.addressMappings[0].NewAddress
	data.newFuncEndAddress = data.addressMappings[len(data.addressMappings)-1].NewAddress
	data.origModule = function.datap
	data.newPclntable = append(data.newPclntable, data.origModule.pclntable...)
	
	if _, ok := os.LookupEnv("ROOKOUT_DEV_DEBUG"); ok {
		dumpBuffer(data.newFuncEntryAddress, data.newFuncEndAddress, "New function")
		oldFuncEnd := data.origFuncEntryAddress + data.offsetMappings[len(data.offsetMappings)-1].OriginalAddress
		dumpBuffer(data.origFuncEntryAddress, oldFuncEnd, "Original function")
	}

	if funcOffset, exists := findFuncOffsetInModule(data.origFuncEntryAddress, data.origModule); exists {
		data.funcOffset = funcOffset
	} else {
		return errors.New("couldn't find func offset in module data")
	}
	if err := data.patchFuncAddress(); err != nil {
		return err
	}
	allPCDataInfo, err := data.patchPCData()
	if err != nil {
		return err
	}
	if err = data.patchDeferReturn(); err != nil {
		return err
	}
	if err = data.createFuncTable(); err != nil {
		return err
	}
	if err = data.createPCHeader(); err != nil {
		return err
	}
	if err = data.createFindFuncBucket(); err != nil {
		return err
	}

	module, err := data.getModuleData()
	if err != nil {
		return err
	}
	if err = validateModuleFuncAndPCDataTables(&module, data.newFuncEntryAddress, data.newFuncEndAddress, allPCDataInfo); err != nil {
		return err
	}
	addModule(&module)
	return nil
}

func validateModuleFuncAndPCDataTables(module *moduledata, newFuncEntry, newFuncEnd uintptr, pcdataInfo *allPCDataPatchInfo) error {
	if err := validateModuleBoundaries(module, newFuncEntry, newFuncEnd); err != nil {
		return err
	}
	if err := validateModuleFtab(module, newFuncEntry); err != nil {
		return err
	}
	if err := validatePatchedFunctionInfo(module, newFuncEntry, pcdataInfo); err != nil {
		return err
	}
	return nil
}

func validateModuleBoundaries(module *moduledata, newFuncEntry, newFuncEnd uintptr) error {
	if module.minpc != newFuncEntry || module.maxpc != newFuncEnd {
		return fmt.Errorf("bad boundaries of module %s. Expected [%d, %d), Got [%d, %d)", module.modulename, newFuncEntry, newFuncEnd, module.minpc, module.maxpc)
	}
	
	return nil
}

const patchedIdx = 1
const expectedFtabSize = 3 

func validatePatchedFunctionInfo(module *moduledata, newFuncEntry uintptr, info *allPCDataPatchInfo) error {
	funcOffset := module.ftab[patchedIdx].funcoff
	moduleName := module.modulename
	fInfo := FuncInfo{
		_func: (*_func)(unsafe.Pointer(&module.pclntable[funcOffset])),
		datap: module,
	}
	if fInfo.getEntry() != newFuncEntry {
		return fmt.Errorf("got bad function entry for patched function in module %s. Expected %d, got %d", moduleName, newFuncEntry, fInfo.getEntry())
	}
	if int(fInfo.npcdata) != len(info.pcdataInfo) {
		return fmt.Errorf("got different npcdata for patched function in module %s. Expected %d, got %d", moduleName, len(info.pcdataInfo), int(fInfo.npcdata))
	}
	for table := 0; table < len(info.pcdataInfo); table++ {
		tableOff := pcdatastart(fInfo, uint32(table))
		if info.pcdataInfo[table] == nil && tableOff > 0 {
			return fmt.Errorf("got unexpected pcdata table %d for patched function in module %s. The table should not exist", table, moduleName)
		}
		if info.pcdataInfo[table] != nil && tableOff == 0 {
			return fmt.Errorf("got missing pcdata table %d for patched function in module %s", table, moduleName)
		}
		if info.pcdataInfo[table] == nil {
			continue 
		}
		if err := validatePCDataTable(module, info.pcdataInfo[table], fInfo, tableOff, fmt.Sprintf("pc data table %d", table)); err != nil {
			return err
		}
	}
	if err := validatePCDataTable(module, info.pcFileInfo, fInfo, uint32(fInfo.pcfile), "pc file"); err != nil {
		return err
	}
	if err := validatePCDataTable(module, info.pcLineInfo, fInfo, uint32(fInfo.pcln), "pc line"); err != nil {
		return err
	}
	if err := validatePCDataTable(module, info.pcSPInfo, fInfo, uint32(fInfo.pcsp), "pc sp"); err != nil {
		return err
	}
	return nil
}




func validatePCDataTableOffsetAndData(module *moduledata, expectedInfo *pcDataTableInfo, tableName string, tableOffset uint32) error {
	moduleName := module.modulename
	if expectedInfo.offset != tableOffset {
		return fmt.Errorf("got bad offset for %s for patched function in module %s. Expected %d, got %d", tableName, moduleName, expectedInfo.offset, tableOffset)
	}
	pcTab := getPCTab(module)[tableOffset:]
	if len(pcTab) < len(expectedInfo.data) {
		return fmt.Errorf("%s for patched function in module %s is to short. Need at least %d bytes, got %d", tableName, moduleName, len(expectedInfo.data), len(pcTab))
	}
	for i := 0; i < len(expectedInfo.data); i++ {
		if expectedInfo.data[i] != pcTab[i] {
			return fmt.Errorf("got bad byte #%d in %s for patched function in module %s. Expected %d, got %d", i, tableName, moduleName, expectedInfo.data[i], pcTab[i])
		}
	}
	return nil
}



func validatePCDataTableFormat(fInfo FuncInfo, expectedInfo *pcDataTableInfo, moduleName, tableName string) error {
	val, pc := pcvalue(fInfo, expectedInfo.offset, expectedInfo.lastPCDataPC, nil, false)
	if val == -1 && pc == 0 {
		return fmt.Errorf("got bad format for table %s in module %s", tableName, moduleName)
	}
	return nil
}

func validatePCDataTable(module *moduledata, expectedInfo *pcDataTableInfo, fInfo FuncInfo, tableOffset uint32, tableName string) error {
	if err := validatePCDataTableOffsetAndData(module, expectedInfo, tableName, tableOffset); err != nil {
		return err
	}
	
	if err := validatePCDataTableFormat(fInfo, expectedInfo, module.modulename, tableName); err != nil {
		return err
	}
	return nil
}
