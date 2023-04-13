package variable

import (
	"debug/dwarf"
	"errors"
	"strings"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/services/collection/memory"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/frame"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)

func GetVariableLocators(pc uint64, line int, function *binary_info.Function, binaryInfo *binary_info.BinaryInfo) ([]*VariableLocator, error) {
	root, err := godwarf.LoadTree(function.Offset, binaryInfo.Dwarf, binaryInfo.Images[0].StaticBase)
	if err != nil {
		return nil, err
	}

	variableLocators := getVariableLocators(root, 0, pc, line, function, binaryInfo)
	return variableLocators, nil
}

func getVariableLocators(root *godwarf.Tree, depth int, pc uint64, line int, function *binary_info.Function, binaryInfo *binary_info.BinaryInfo) []*VariableLocator {
	switch root.Tag {
	case dwarf.TagInlinedSubroutine, dwarf.TagLexDwarfBlock, dwarf.TagSubprogram:
		var variables []*VariableLocator
		if root.ContainsPC(pc) {
			for _, child := range root.Children {
				variables = append(variables, getVariableLocators(child, depth+1, pc, line, function, binaryInfo)...)
			}
		}
		return variables
	case dwarf.TagFormalParameter, dwarf.TagVariable:
		if name := root.Val(dwarf.AttrName).(string); strings.HasPrefix(name, "~") {
			return nil
		}
		visibilityOffset := 0
		if root.Tag != dwarf.TagFormalParameter {
			
			
			
			visibilityOffset = 1
		}
		if declLine, ok := root.Val(dwarf.AttrDeclLine).(int64); !ok || line >= int(declLine)+visibilityOffset {
			newVar, err := NewVariableLocator(root, function, pc, binaryInfo)
			if err != nil {
				return nil
			}
			return []*VariableLocator{newVar}
		}
	}

	return nil
}

type VariableLocator struct {
	VariableName string
	variableType godwarf.Type
	locator      *op.DwarfLocator
	binaryInfo   *binary_info.BinaryInfo
	function     *binary_info.Function
	pc           uint64
}

func NewVariableLocator(entry *godwarf.Tree, function *binary_info.Function, pc uint64, binaryInfo *binary_info.BinaryInfo) (*VariableLocator, error) {
	name, varType, err := binaryInfo.ReadVariableEntry(entry)
	if err != nil {
		return nil, err
	}

	instr, _, err := binaryInfo.LocationExpr(entry.Entry, dwarf.AttrLocation, pc)
	if err != nil {
		return nil, err
	}

	locator, err := op.NewDwarfLocator(instr, binaryInfo.PointerSize)
	if err != nil {
		return nil, err
	}

	v := &VariableLocator{VariableName: name, variableType: varType, locator: locator, binaryInfo: binaryInfo, function: function, pc: pc}
	return v, nil
}

func (v *VariableLocator) Locate(regs registers.Registers, dictAddr uint64, variablesCache map[VariablesCacheKey]VariablesCacheValue, objectDumpConfig config.ObjectDumpConfig) *Variable {
	var mem memory.MemoryReader = &memory.ProcMemory{}
	var addr int64
	var pieces []op.Piece

	dwarfRegs, err := v.advanceRegs(regs)
	if err != nil {
		logger.Logger().WithError(err).Warningf("Failed to advance regs")
	} else {
		addr, pieces, err = v.locator.Locate(dwarfRegs)
		if err != nil {
			logger.Logger().WithError(err).Warningf("Failed to locate")
		}
		if pieces != nil {
			addr = memory.FakeAddress
			var cmem *memory.CompositeMemory
			cmem, err = memory.NewCompositeMemory(mem, dwarfRegs, pieces, v.binaryInfo.PointerSize)
			if cmem != nil {
				mem = cmem
			}
		}
	}

	newVar := NewVariable(v.VariableName, uint64(addr), v.variableType, mem, v.binaryInfo, objectDumpConfig, dictAddr, variablesCache)
	if pieces != nil {
		newVar.Flags |= VariableFakeAddress
	}
	if err != nil {
		newVar.Unreadable = err
	}
	return newVar
}

func (v *VariableLocator) advanceRegs(regs registers.Registers) (op.DwarfRegisters, error) {
	dwarfRegs := binary_info.RegistersToDwarfRegisters(0, regs)
	fde, err := v.binaryInfo.FrameEntries.FDEForPC(v.pc) 
	var framectx *frame.FrameContext
	if _, nofde := err.(*frame.ErrNoFDEForPC); nofde {
		framectx = binary_info.FixFrameUnwindContext(nil, v.pc, v.binaryInfo)
	} else {
		framectx = binary_info.FixFrameUnwindContext(fde.EstablishFrame(v.pc), v.pc, v.binaryInfo)
	}

	cfareg, err := executeFrameRegRule(0, framectx.CFA, 0, dwarfRegs, v.binaryInfo)
	if err != nil {
		return op.DwarfRegisters{}, err
	}
	if cfareg == nil {
		return op.DwarfRegisters{}, errors.New("no cfareg")
	}

	callimage := v.binaryInfo.PCToImage(v.pc)

	dwarfRegs.StaticBase = callimage.StaticBase
	dwarfRegs.CFA = int64(cfareg.Uint64Val)
	dwarfRegs.FrameBase = v.frameBase(dwarfRegs)

	
	
	
	
	
	
	
	dwarfRegs.AddReg(dwarfRegs.SPRegNum, cfareg)

	for i, regRule := range framectx.Regs {
		reg, err := executeFrameRegRule(i, regRule, dwarfRegs.CFA, dwarfRegs, v.binaryInfo)
		dwarfRegs.AddReg(i, reg)
		if i == framectx.RetAddrReg {
			if reg == nil {
				if err == nil {
					return op.DwarfRegisters{}, err
				}
			}
		}
	}
	return dwarfRegs, nil
}

func executeFrameRegRule(regnum uint64, rule frame.DWRule, cfa int64, dwarfRegs op.DwarfRegisters, binaryInfo *binary_info.BinaryInfo) (*op.DwarfRegister, error) {
	switch rule.Rule {
	default:
		fallthrough
	case frame.RuleUndefined:
		return nil, nil
	case frame.RuleSameVal:
		if dwarfRegs.Reg(regnum) == nil {
			return nil, nil
		}
		reg := *dwarfRegs.Reg(regnum)
		return &reg, nil
	case frame.RuleOffset:
		return readRegisterAt(regnum, uint64(cfa+rule.Offset))
	case frame.RuleValOffset:
		return op.DwarfRegisterFromUint64(uint64(cfa + rule.Offset)), nil
	case frame.RuleRegister:
		return dwarfRegs.Reg(rule.Reg), nil
	case frame.RuleExpression:
		v, _, err := op.ExecuteStackProgram(dwarfRegs, rule.Expression, binaryInfo.PointerSize)
		if err != nil {
			return nil, err
		}
		return readRegisterAt(regnum, uint64(v))
	case frame.RuleValExpression:
		v, _, err := op.ExecuteStackProgram(dwarfRegs, rule.Expression, binaryInfo.PointerSize)
		if err != nil {
			return nil, err
		}
		return op.DwarfRegisterFromUint64(uint64(v)), nil
	case frame.RuleArchitectural:
		return nil, errors.New("architectural frame rules are unsupported")
	case frame.RuleCFA:
		if dwarfRegs.Reg(rule.Reg) == nil {
			return nil, nil
		}
		return op.DwarfRegisterFromUint64(uint64(int64(dwarfRegs.Uint64Val(rule.Reg)) + rule.Offset)), nil
	case frame.RuleFramePointer:
		curReg := dwarfRegs.Reg(rule.Reg)
		if curReg == nil {
			return nil, nil
		}
		if curReg.Uint64Val <= uint64(cfa) {
			return readRegisterAt(regnum, curReg.Uint64Val)
		}
		newReg := *curReg
		return &newReg, nil
	}
}

func readRegisterAt(regnum uint64, addr uint64) (*op.DwarfRegister, error) {
	buf := make([]byte, binary_info.RegSize(regnum))
	memReader := memory.ProcMemory{}
	_, err := memReader.ReadMemory(buf, addr)
	if err != nil {
		return nil, err
	}
	return op.DwarfRegisterFromBytes(buf), nil
}

func (v *VariableLocator) frameBase(regs op.DwarfRegisters) int64 {
	dwarfTree, err := v.binaryInfo.Images[0].GetDwarfTree(v.function.Offset)
	if err != nil {
		return 0
	}
	fb, _, _, _ := v.binaryInfo.Location(dwarfTree.Entry, dwarf.AttrFrameBase, v.pc, regs)
	return fb
}
