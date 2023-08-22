//go:build go1.15 && !go1.22
// +build go1.15,!go1.22

package module

import (
	_ "unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"

	"github.com/go-errors/errors"
)

type PCDataEntry struct {
	Offset uintptr
	Value  int32
}


func decodePCDataEntries(p []byte) (pcDataEntries []PCDataEntry) {
	if p == nil {
		return pcDataEntries
	}
	var pc uintptr
	val := int32(-1)
	p, ok := step(p, &pc, &val, true)
	for {
		if !ok {
			return pcDataEntries
		}
		pcDataEntries = append(pcDataEntries, PCDataEntry{Offset: pc, Value: val})
		if len(p) <= 0 {
			return pcDataEntries
		}
		p, ok = step(p, &pc, &val, false)
	}
}



func findNewAddressByOriginalAddress(originalAddress uintptr, addressMappings []AddressMapping) (uintptr, bool) {
	for _, mapping := range addressMappings {
		if originalAddress == mapping.OriginalAddress {
			return mapping.NewAddress, true
		}
	}
	return 0, false
}



func updatePCDataEntries(pcDataEntries []PCDataEntry, offsetMappings []AddressMapping) {
	for i := 0; i < len(pcDataEntries); i++ {
		newOffset, found := findNewAddressByOriginalAddress(pcDataEntries[i].Offset, offsetMappings)
		if found {
			pcDataEntries[i].Offset = newOffset
		}
	}
}



func getEntryForOffset(offset uintptr, pcDataEntries []PCDataEntry) (int, *PCDataEntry) {
	prevPCOffset := uintptr(0)
	for index, pcDataEntry := range pcDataEntries {
		if prevPCOffset <= offset && pcDataEntry.Offset > offset {
			return index, &pcDataEntry
		}
		prevPCOffset = pcDataEntry.Offset
	}

	return -1, nil
}


func UvarintToBytes(x uint64) (buf []byte) {
	for x >= 0x80 {
		buf = append(buf, byte(x)|0x80)
		x >>= 7
	}
	buf = append(buf, byte(x))
	return buf
}


func writeUvarintToBytes(bytes []byte, x uint64) ([]byte, error) {
	encodedVariant := UvarintToBytes(x)
	bytes = append(bytes, encodedVariant...)
	return bytes, nil
}

func encode(v int32) uint32 {
	return uint32(v<<1) ^ uint32(v>>31)
}


func writePCDataEntry(p []byte, value int32, offset int32) ([]byte, error) {
	p, err := writeUvarintToBytes(p, uint64(encode(value)))
	if err != nil {
		return nil, err
	}
	p, err = writeUvarintToBytes(p, uint64(offset/pcQuantum))
	if err != nil {
		return nil, err
	}

	return p, nil
}

func removeDuplicateValues(pcDataEntries []PCDataEntry) (noDups []PCDataEntry) {
	if len(pcDataEntries) == 0 {
		return nil
	}
	for i := range pcDataEntries {
		if i == len(pcDataEntries)-1 {
			noDups = append(noDups, pcDataEntries[i])
			break
		}

		if pcDataEntries[i].Value == pcDataEntries[i+1].Value {
			continue
		}
		noDups = append(noDups, pcDataEntries[i])
	}

	return noDups
}


func encodePCDataEntries(pcDataEntries []PCDataEntry) (encoded []byte, err error) {
	if len(pcDataEntries) == 0 {
		return nil, nil
	}
	encoded = make([]byte, 0, len(pcDataEntries)*20)
	prevOffset := int32(0)
	prevValue := int32(-1)

	for _, newPair := range pcDataEntries {
		valueDelta := newPair.Value - prevValue
		offsetDelta := int32(newPair.Offset) - prevOffset
		encoded, err = writePCDataEntry(encoded, valueDelta, offsetDelta)
		if err != nil {
			return nil, err
		}
		prevOffset = int32(newPair.Offset)
		prevValue = newPair.Value
	}
	
	encoded = append(encoded, 0)
	return encoded, nil
}









func addCallbackEntry(pcDataEntries []PCDataEntry, callbackOffsetStart, callbackOffsetEnd uintptr, pcDataGenerator func(uintptr, uintptr, int32) ([]PCDataEntry, error)) ([]PCDataEntry, error) {
	callbackEntryIndex, entryForCallback := getEntryForOffset(callbackOffsetStart, pcDataEntries)
	if callbackEntryIndex == -1 {
		return nil, errors.New("No PCData entry in table after breakpoint")
	}

	newPCDataEntries := pcDataEntries
	callbackPCDataEntries, err := pcDataGenerator(callbackOffsetStart, callbackOffsetEnd, entryForCallback.Value)
	if err != nil {
		return nil, err
	}
	fromCallbackUntilEnd := append(callbackPCDataEntries, newPCDataEntries[callbackEntryIndex:]...)
	newPCDataEntries = append(newPCDataEntries[:callbackEntryIndex], fromCallbackUntilEnd...)
	return newPCDataEntries, nil
}

func addOffsetAndValueToEntries(offset uintptr, value int32, entries []PCDataEntry) []PCDataEntry {
	var callbackPCDataEntries []PCDataEntry

	for _, callbackPCInfo := range entries {
		callbackPCDataEntries = append(callbackPCDataEntries, PCDataEntry{offset + callbackPCInfo.Offset, value + callbackPCInfo.Value})
	}

	return callbackPCDataEntries
}


func addCallbacksEntries(pcDataEntries []PCDataEntry, offsetMappings []AddressMapping, pcDataGenerator func(uintptr, uintptr, int32) ([]PCDataEntry, error)) ([]PCDataEntry, error) {
	for mapIndex, mapping := range offsetMappings {
		if _, ok := CallbacksMarkers[mapping.OriginalAddress]; ok {
			callbackOffsetStart := mapping.NewAddress
			callbackOffsetEnd := offsetMappings[mapIndex+1].NewAddress
			newPCDataEntries, err := addCallbackEntry(pcDataEntries, callbackOffsetStart, callbackOffsetEnd, pcDataGenerator)
			if err != nil {
				return nil, err
			}
			pcDataEntries = newPCDataEntries
		}
	}

	return pcDataEntries, nil
}

type PCDataPatcher struct {
	newFuncEntry   uintptr
	offsetMappings []AddressMapping
	isPatched      bool
	instSizeReader func(pc uintptr) (uintptr, rookoutErrors.RookoutError)
}








func verifyOffsetMappings(offsetMappings []AddressMapping) rookoutErrors.RookoutError {
	numLastMappingsToCheck := 2
	if len(offsetMappings) < numLastMappingsToCheck {
		return rookoutErrors.NewIllegalAddressMappings()
	}
	lastMappings := offsetMappings[len(offsetMappings)-numLastMappingsToCheck:]
	for _, m := range lastMappings {
		if _, ok := CallbacksMarkers[m.OriginalAddress]; ok {
			return rookoutErrors.NewIllegalAddressMappings()
		}
	}

	return nil
}

func NewPCDataPatcher(newFuncEntry uintptr, offsetMappings []AddressMapping, isPatched bool, instSizeReader func(uintptr) (uintptr, rookoutErrors.RookoutError)) (*PCDataPatcher, error) {
	if err := verifyOffsetMappings(offsetMappings); err != nil {
		return nil, err
	}
	return &PCDataPatcher{
		newFuncEntry:   newFuncEntry,
		offsetMappings: offsetMappings,
		isPatched:      isPatched,
		instSizeReader: instSizeReader,
	}, nil
}

func (p *PCDataPatcher) updateOffsets(table []PCDataEntry) error {
	updatePCDataEntries(table, p.offsetMappings)

	
	
	
	
	
	
	
	
	
	
	var newMappings []AddressMapping
	for i := 0; i < len(p.offsetMappings); i++ {
		currentMapping := p.offsetMappings[i]
		if _, ok := CallbacksMarkers[currentMapping.OriginalAddress]; ok {
			i++ 
			
			for ; i < len(p.offsetMappings); i++ {
				if _, ok := CallbacksMarkers[p.offsetMappings[i].OriginalAddress]; !ok {
					break
				}
			}
			if i < len(p.offsetMappings) {
				newMappings = append(newMappings, AddressMapping{OriginalAddress: p.offsetMappings[i].NewAddress, NewAddress: currentMapping.NewAddress})
			} else {
				
				return rookoutErrors.NewIllegalAddressMappings()
			}
		}
	}
	updatePCDataEntries(table, newMappings)

	return nil
}

func (p *PCDataPatcher) createSanitized(oldTable []PCDataEntry, builder func([]PCDataEntry) ([]PCDataEntry, error)) ([]PCDataEntry, error) {
	var oldTableCopy []PCDataEntry = nil
	if len(oldTable) > 0 {
		oldTableCopy = make([]PCDataEntry, len(oldTable))
		copy(oldTableCopy, oldTable)
	}

	newTable, err := builder(oldTableCopy)
	if err != nil {
		return nil, err
	}
	return removeDuplicateValues(newTable), nil
}




func (p *PCDataPatcher) CreatePCData(tableIndex int, oldTable []PCDataEntry) ([]PCDataEntry, error) {
	return p.createSanitized(oldTable,
		func(table []PCDataEntry) ([]PCDataEntry, error) {
			err := p.updateOffsets(table)
			if err != nil {
				return nil, err
			}

			if p.isPatched && tableIndex == _PCDATA_UnsafePoint {
				table, err = p.fixAsyncUnsafePointPCData(table)
			}

			return table, err
		})
}


func (p *PCDataPatcher) CreatePCLine(oldTable []PCDataEntry) ([]PCDataEntry, error) {
	return p.createSanitized(oldTable,
		func(table []PCDataEntry) ([]PCDataEntry, error) {
			err := p.updateOffsets(table)
			return table, err
		})
}



func (p *PCDataPatcher) CreatePCSP(oldTable []PCDataEntry) ([]PCDataEntry, error) {
	return p.createSanitized(oldTable,
		func(table []PCDataEntry) ([]PCDataEntry, error) {
			err := p.updateOffsets(table)
			if err != nil {
				return nil, err
			}
			if p.isPatched {
				pcspGenerator := func(callbackOffsetStart, callbackOffsetEnd uintptr, callbackOffsetValue int32) ([]PCDataEntry, error) {
					
					patchedPCSP, err := generatePCSP(callbackOffsetStart+p.newFuncEntry, callbackOffsetEnd+p.newFuncEntry)
					if err != nil {
						return nil, err
					}
					
					return addOffsetAndValueToEntries(callbackOffsetStart, callbackOffsetValue, patchedPCSP), nil
				}
				table, err = addCallbacksEntries(table, p.offsetMappings, pcspGenerator)
				if err != nil {
					return nil, err
				}
			}
			return table, nil
		})
}








func (p *PCDataPatcher) fixAsyncUnsafePointPCData(entries []PCDataEntry) ([]PCDataEntry, error) {
	if len(entries) == 0 {
		entries = []PCDataEntry{{Offset: p.offsetMappings[len(p.offsetMappings)-1].NewAddress, Value: _PCDATA_UnsafePointSafe}}
	}
	
	pcdataAsyncUnsafeGenerator := func(callbackOffsetStart, callbackOffsetEnd uintptr, callbackOffsetValue int32) ([]PCDataEntry, error) {
		firstCallbackInstructionPC := callbackOffsetStart + p.newFuncEntry
		firstCallbackInstructionSize, err := p.instSizeReader(firstCallbackInstructionPC)
		if err != nil {
			return nil, err
		}
		callbackEntries := []PCDataEntry{
			{
				
				Offset: callbackOffsetStart + firstCallbackInstructionSize,
				Value:  callbackOffsetValue,
			},
			{
				
				Offset: callbackOffsetEnd,
				Value:  _PCDATA_UnsafePointUnsafe,
			},
		}
		return callbackEntries, nil
	}
	return addCallbacksEntries(entries, p.offsetMappings, pcdataAsyncUnsafeGenerator)
}
