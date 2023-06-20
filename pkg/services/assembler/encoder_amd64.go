package assembler

import (
	"encoding/binary"
	"math"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

const (
	j  = "\xe9"
	jl = "\x0f\x8c"
)

func encodeBranch(src uintptr, dst uintptr, op string) ([]byte, rookoutErrors.RookoutError) {
	relativeAddr := int64(dst - (src + uintptr(len(op)) + 4))
	if relativeAddr > math.MaxInt32 || relativeAddr < math.MinInt32 {
		return nil, rookoutErrors.NewBranchDestTooFar(src, dst)
	}

	offset := make([]byte, 4)
	binary.LittleEndian.PutUint32(offset, uint32(relativeAddr))

	return append([]byte(op), offset...), nil
}

func EncodeJmp(src uintptr, dst uintptr) ([]byte, rookoutErrors.RookoutError) {
	return encodeBranch(src, dst, j)
}

func EncodeJL(src uintptr, dst uintptr) ([]byte, rookoutErrors.RookoutError) {
	return encodeBranch(src, dst, jl)
}
