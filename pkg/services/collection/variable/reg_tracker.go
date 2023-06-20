package variable

import (
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)


const dummyVal = 100

type regTracker struct {
	regsUsed map[uint64]struct{}
}

func newRegTracker() *regTracker {
	return &regTracker{regsUsed: make(map[uint64]struct{})}
}

func (r *regTracker) Reg(i uint64) *op.DwarfRegister {
	return &op.DwarfRegister{
		Uint64Val: dummyVal,
		Bytes:     []byte{dummyVal},
	}
}

func (r *regTracker) Uint64Val(i uint64) uint64 {
	return dummyVal
}

func (r *regTracker) CFA() int64         { return dummyVal }
func (r *regTracker) StaticBase() uint64 { return dummyVal }
func (r *regTracker) FrameBase() int64   { return dummyVal }

func (r *regTracker) ResolveRegsUsed(variableType godwarf.Type, pieces []op.Piece) ([]assembler.Reg, rookoutErrors.RookoutError) {
	if len(pieces) == 0 {
		return nil, nil
	}

	pointerPieces, _ := r.resolveRegsUsed(variableType, pieces)

	var regs []assembler.Reg
	for _, piece := range pointerPieces {
		if piece.Kind != op.RegPiece {
			continue
		}

		reg, ok := assembler.DwarfRegToAsmReg(piece.Val)
		if !ok {
			return nil, rookoutErrors.NewInvalidDwarfRegister(piece.Val)
		}
		regs = append(regs, reg)
	}
	return regs, nil
}


func (r *regTracker) resolveRegsUsed(variableType godwarf.Type, pieces []op.Piece) (pointerPieces []op.Piece, basicPieces []op.Piece) {
	switch t := resolveTypedef(variableType).(type) {
	
	case *godwarf.StringType:
		return pieces[:1], pieces[1:2]

	
	case *godwarf.InterfaceType:
		return pieces[:2], nil

	
	case *godwarf.MapType, *godwarf.FuncType, *godwarf.PtrType, *godwarf.ChanType:
		return pieces[:1], nil

	
	case *godwarf.SliceType, *godwarf.DotDotDotType:
		return pieces[:1], pieces[1:3]

	
	case *godwarf.IntType, *godwarf.CharType, *godwarf.ComplexType, *godwarf.FloatType, *godwarf.BoolType, *godwarf.EnumType, *godwarf.UintType, *godwarf.UcharType:
		return nil, pieces[:1]

	
	case *godwarf.StructType:
		for _, field := range t.Field {
			pointers, basics := r.resolveRegsUsed(field.Type, pieces)
			pointerPieces = append(pointerPieces, pointers...)
			basicPieces = append(basicPieces, basics...)
			pieces = pieces[len(pointers)+len(basics):]
		}
		return pointerPieces, basicPieces

	
	case *godwarf.ParametricType:
		return pieces, nil
	}

	logger.Logger().Warningf("Unknown type for regs used: %T\n", resolveTypedef(variableType))
	return pieces, nil
}
