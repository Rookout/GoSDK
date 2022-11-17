//go:build go1.15 && !go1.20
// +build go1.15,!go1.20

package module

import (
	"github.com/go-errors/errors"
	_ "unsafe"
)

type PCDataEntry struct {
	Offset uintptr
	Value  int32
}


func decodePCDataEntries(p []byte) (pcDataEntries []PCDataEntry) {
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
	var nearestMapping AddressMapping
	nearestDistance := -1

	for _, mapping := range addressMappings {
		if originalAddress == mapping.OriginalAddress {
			return mapping.NewAddress, true
		}

		
		if mapping.OriginalAddress == BPMarker || mapping.OriginalAddress == PrologueMarker {
			continue
		}

		distance := int(originalAddress - mapping.OriginalAddress)
		if distance < 0 {
			distance = -distance
		}
		if nearestDistance == -1 || nearestDistance > distance {
			nearestDistance = distance
			nearestMapping = mapping
		}
	}

	newAddress := nearestMapping.NewAddress
	newAddress += originalAddress - nearestMapping.OriginalAddress
	return newAddress, false
}





func updatePCDataEntries(pcDataEntries []PCDataEntry, offsetMappings []AddressMapping, onlyExact bool) error {
	for i := 0; i < len(pcDataEntries); i++ {
		newAddress, exact := findNewAddressByOriginalAddress(pcDataEntries[i].Offset, offsetMappings)
		if !onlyExact || exact {
			pcDataEntries[i].Offset = newAddress
		}
	}

	return nil
}


func getEntryAfterOffset(offset uintptr, pcDataEntries []PCDataEntry) (int, *PCDataEntry) {
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
	p, err = writeUvarintToBytes(p, uint64(offset))
	if err != nil {
		return nil, err
	}

	return p, nil
}


func encodePCDataEntries(pcDataEntries []PCDataEntry) (encoded []byte, err error) {
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


func insertPCDataEntry(index int, newEntry PCDataEntry, pcDataEntries []PCDataEntry) []PCDataEntry {
	var newPCDataEntries []PCDataEntry
	newPCDataEntries = append(newPCDataEntries, pcDataEntries[:index]...)
	newPCDataEntries = append(newPCDataEntries, newEntry)
	newPCDataEntries = append(newPCDataEntries, pcDataEntries[index:]...)
	return newPCDataEntries
}


func getEntryAtOffset(offset uintptr, pcDataEntries []PCDataEntry) (*PCDataEntry, bool) {
	for _, entry := range pcDataEntries {
		if entry.Offset == offset {
			return &entry, true
		}
	}

	return nil, false
}




func addCallbackEntry(pcDataEntries []PCDataEntry, callbackOffset uintptr, callbackPcInfos []CallbackPCDataInfo) ([]PCDataEntry, error) {
	callbackEntryIndex, entryAfterCallback := getEntryAfterOffset(callbackOffset, pcDataEntries)
	if callbackEntryIndex == -1 {
		return nil, errors.New("No PCData entry in table after breakpoint")
	}

	newPCDataEntries := pcDataEntries
	callbackPcDataEntries := generateCallbackPCDataEntries(callbackOffset, entryAfterCallback.Value, callbackPcInfos)

	if _, exists := getEntryAtOffset(callbackOffset, pcDataEntries); !exists {
		newPCDataEntries = insertPCDataEntry(callbackEntryIndex, callbackPcDataEntries[0], newPCDataEntries)
		callbackEntryIndex++
	}

	for _, callbackPcDataEntry := range callbackPcDataEntries[1:] {
		newPCDataEntries = insertPCDataEntry(callbackEntryIndex, callbackPcDataEntry, newPCDataEntries)
		callbackEntryIndex++
	}

	return newPCDataEntries, nil
}

func generateCallbackPCDataEntries(callbackOffset uintptr, entryAfterCallbackValue int32, callbackPcInfos []CallbackPCDataInfo) []PCDataEntry {
	var callbackPcDataEntries []PCDataEntry

	for _, callbackPcInfo := range callbackPcInfos {
		callbackPcDataEntries = append(callbackPcDataEntries, PCDataEntry{callbackOffset + callbackPcInfo.offset, entryAfterCallbackValue + callbackPcInfo.valueDiff})
	}

	return callbackPcDataEntries
}


func addCallbacksEntries(pcDataEntries []PCDataEntry, offsetMappings []AddressMapping, callbackMarkerToCallbackPcInfos map[uintptr][]CallbackPCDataInfo) ([]PCDataEntry, error) {
	for mapIndex, mapping := range offsetMappings {
		for marker, callbackPCSPInfos := range callbackMarkerToCallbackPcInfos {
			if mapping.OriginalAddress == marker {
				callbackOffset := mapping.NewAddress
				lineAfterCallbackMapping := offsetMappings[mapIndex+1]
				lineAfterCallbackOffset := lineAfterCallbackMapping.NewAddress
				callbackAddressMapping := AddressMapping{OriginalAddress: lineAfterCallbackOffset, NewAddress: callbackOffset}

				if err := updatePCDataEntries(pcDataEntries, []AddressMapping{callbackAddressMapping}, true); err != nil {
					return nil, err
				}

				newPCDataEntries, err := addCallbackEntry(pcDataEntries, callbackOffset, callbackPCSPInfos)
				if err != nil {
					return nil, err
				}
				pcDataEntries = newPCDataEntries
				break 
			}
		}
	}

	return pcDataEntries, nil
}




func updatePCDataOffsets(p []byte, offsetMappings []AddressMapping, callbackMarkerToCallbackPCSPInfos map[uintptr][]CallbackPCDataInfo) ([]byte, uintptr, error) {
	pcDataEntries := decodePCDataEntries(p)
	if err := updatePCDataEntries(pcDataEntries, offsetMappings, false); err != nil {
		return nil, 0, err
	}

	if callbackMarkerToCallbackPCSPInfos != nil {
		newPCDataEntries, err := addCallbacksEntries(pcDataEntries, offsetMappings, callbackMarkerToCallbackPCSPInfos)
		if err != nil {
			return nil, 0, err
		}
		pcDataEntries = newPCDataEntries
	}

	encoded, err := encodePCDataEntries(pcDataEntries)
	if err != nil {
		return nil, 0, err
	}
	lastNewPCDataOffset := pcDataEntries[len(pcDataEntries)-1].Offset
	lastNewValidPCOffset := lastNewPCDataOffset - 1 
	return encoded, lastNewValidPCOffset, nil
}
