package assembler

import (
	"encoding/binary"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

const b = 0x14000000

func abs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

func EncodeJmp(src uintptr, dst uintptr) ([]byte, rookoutErrors.RookoutError) {
	
	relativeAddr := int64(dst-src) / int64(4)

	
	
	
	if relativeAddr%4 != 0 {
		return nil, rookoutErrors.NewInvalidBranchDest(src, dst)
	} else if abs(relativeAddr)&0b1111111111111111111111111 != abs(relativeAddr) {
		return nil, rookoutErrors.NewBranchDestTooFar(src, dst)
	}

	
	encodedOffset := uint32(int32(relativeAddr) & 0b11111111111111111111111111)

	
	encodedInst := uint32(b) | encodedOffset

	
	encodedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(encodedBytes, encodedInst)

	return encodedBytes, nil
}
