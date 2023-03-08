//go:build go1.15 && !go1.20
// +build go1.15,!go1.20

package module

import (
	_ "unsafe"

	"github.com/go-errors/errors"
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
	p, err = writeUvarintToBytes(p, uint64(offset/pcQuantum))
	if err != nil {
		return nil, err
	}

	return p, nil
}

func removeDuplicateValues(pcDataEntries []PCDataEntry) (noDups []PCDataEntry) {
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
	encoded = make([]byte, 0, len(pcDataEntries)*20)
	prevOffset := int32(0)
	prevValue := int32(-1)

	pcDataEntries = removeDuplicateValues(pcDataEntries)

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




func addCallbackEntry(pcDataEntries []PCDataEntry, callbackOffset uintptr, newEntries []PCDataEntry) ([]PCDataEntry, error) {
	callbackEntryIndex, entryAfterCallback := getEntryAfterOffset(callbackOffset, pcDataEntries)
	if callbackEntryIndex == -1 {
		return nil, errors.New("No PCData entry in table after breakpoint")
	}

	newPCDataEntries := pcDataEntries
	callbackPCDataEntries := generateCallbackPCDataEntries(callbackOffset, entryAfterCallback.Value, newEntries)

	if _, exists := getEntryAtOffset(callbackOffset, pcDataEntries); !exists {
		newPCDataEntries = insertPCDataEntry(callbackEntryIndex, callbackPCDataEntries[0], newPCDataEntries)
		callbackEntryIndex++
	}

	for _, callbackPCDataEntry := range callbackPCDataEntries[1:] {
		newPCDataEntries = insertPCDataEntry(callbackEntryIndex, callbackPCDataEntry, newPCDataEntries)
		callbackEntryIndex++
	}

	return newPCDataEntries, nil
}

func generateCallbackPCDataEntries(callbackOffset uintptr, entryAfterCallbackValue int32, newEntries []PCDataEntry) []PCDataEntry {
	var callbackPCDataEntries []PCDataEntry

	for _, callbackPCInfo := range newEntries {
		callbackPCDataEntries = append(callbackPCDataEntries, PCDataEntry{callbackOffset + callbackPCInfo.Offset, entryAfterCallbackValue + callbackPCInfo.Value})
	}

	return callbackPCDataEntries
}


func addCallbacksEntries(pcDataEntries []PCDataEntry, offsetMappings []AddressMapping, pcDataGenerator func(uintptr, uintptr) ([]PCDataEntry, error)) ([]PCDataEntry, error) {
	for mapIndex, mapping := range offsetMappings {
		if _, ok := CallbacksMarkers[mapping.OriginalAddress]; ok {
			entries, err := pcDataGenerator(mapping.NewAddress, offsetMappings[mapIndex+1].NewAddress)
			if err != nil {
				return nil, err
			}
			callbackOffset := mapping.NewAddress
			lineAfterCallbackMapping := offsetMappings[mapIndex+1]
			lineAfterCallbackOffset := lineAfterCallbackMapping.NewAddress
			callbackAddressMapping := AddressMapping{OriginalAddress: lineAfterCallbackOffset, NewAddress: callbackOffset}

			if err := updatePCDataEntries(pcDataEntries, []AddressMapping{callbackAddressMapping}, true); err != nil {
				return nil, err
			}

			newPCDataEntries, err := addCallbackEntry(pcDataEntries, callbackOffset, entries)
			if err != nil {
				return nil, err
			}
			pcDataEntries = newPCDataEntries
		}
	}

	return pcDataEntries, nil
}




func updatePCDataOffsets(p []byte, offsetMappings []AddressMapping, pcDataGenerator func(uintptr, uintptr) ([]PCDataEntry, error)) ([]byte, uintptr, error) {
	pcDataEntries := decodePCDataEntries(p)
	if err := updatePCDataEntries(pcDataEntries, offsetMappings, false); err != nil {
		return nil, 0, err
	}

	if pcDataGenerator != nil {
		newPCDataEntries, err := addCallbacksEntries(pcDataEntries, offsetMappings, pcDataGenerator)
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
