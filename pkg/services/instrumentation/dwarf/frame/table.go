package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
)


type DWRule struct {
	Rule       Rule
	Offset     int64
	Reg        uint64
	Expression []byte
}


type FrameContext struct {
	loc           uint64
	order         binary.ByteOrder
	address       uint64
	CFA           DWRule
	Regs          map[uint64]DWRule
	initialRegs   map[uint64]DWRule
	prevRegs      map[uint64]DWRule
	buf           *bytes.Buffer
	cie           *CommonInformationEntry
	RetAddrReg    uint64
	codeAlignment uint64
	dataAlignment int64
}


const (
	DW_CFA_nop                = 0x0      
	DW_CFA_set_loc            = 0x01     
	DW_CFA_advance_loc1       = iota     
	DW_CFA_advance_loc2                  
	DW_CFA_advance_loc4                  
	DW_CFA_offset_extended               
	DW_CFA_restore_extended              
	DW_CFA_undefined                     
	DW_CFA_same_value                    
	DW_CFA_register                      
	DW_CFA_remember_state                
	DW_CFA_restore_state                 
	DW_CFA_def_cfa                       
	DW_CFA_def_cfa_register              
	DW_CFA_def_cfa_offset                
	DW_CFA_def_cfa_expression            
	DW_CFA_expression                    
	DW_CFA_offset_extended_sf            
	DW_CFA_def_cfa_sf                    
	DW_CFA_def_cfa_offset_sf             
	DW_CFA_val_offset                    
	DW_CFA_val_offset_sf                 
	DW_CFA_val_expression                
	DW_CFA_lo_user            = 0x1c     
	DW_CFA_hi_user            = 0x3f     
	DW_CFA_advance_loc        = 0x1 << 6 
	DW_CFA_offset             = 0x2 << 6 
	DW_CFA_restore            = 0x3 << 6 
)


type Rule byte

const (
	RuleUndefined Rule = iota
	RuleSameVal
	RuleOffset
	RuleValOffset
	RuleRegister
	RuleExpression
	RuleValExpression
	RuleArchitectural
	RuleCFA          
	RuleFramePointer 
)

const low_6_offset = 0x3f

type instruction func(frame *FrameContext)


var fnlookup = map[byte]instruction{
	DW_CFA_advance_loc:        advanceloc,
	DW_CFA_offset:             offset,
	DW_CFA_restore:            restore,
	DW_CFA_set_loc:            setloc,
	DW_CFA_advance_loc1:       advanceloc1,
	DW_CFA_advance_loc2:       advanceloc2,
	DW_CFA_advance_loc4:       advanceloc4,
	DW_CFA_offset_extended:    offsetextended,
	DW_CFA_restore_extended:   restoreextended,
	DW_CFA_undefined:          undefined,
	DW_CFA_same_value:         samevalue,
	DW_CFA_register:           register,
	DW_CFA_remember_state:     rememberstate,
	DW_CFA_restore_state:      restorestate,
	DW_CFA_def_cfa:            defcfa,
	DW_CFA_def_cfa_register:   defcfaregister,
	DW_CFA_def_cfa_offset:     defcfaoffset,
	DW_CFA_def_cfa_expression: defcfaexpression,
	DW_CFA_expression:         expression,
	DW_CFA_offset_extended_sf: offsetextendedsf,
	DW_CFA_def_cfa_sf:         defcfasf,
	DW_CFA_def_cfa_offset_sf:  defcfaoffsetsf,
	DW_CFA_val_offset:         valoffset,
	DW_CFA_val_offset_sf:      valoffsetsf,
	DW_CFA_val_expression:     valexpression,
	DW_CFA_lo_user:            louser,
	DW_CFA_hi_user:            hiuser,
}

func executeCIEInstructions(cie *CommonInformationEntry) *FrameContext {
	initialInstructions := make([]byte, len(cie.InitialInstructions))
	copy(initialInstructions, cie.InitialInstructions)
	frame := &FrameContext{
		cie:           cie,
		Regs:          make(map[uint64]DWRule),
		RetAddrReg:    cie.ReturnAddressRegister,
		initialRegs:   make(map[uint64]DWRule),
		prevRegs:      make(map[uint64]DWRule),
		codeAlignment: cie.CodeAlignmentFactor,
		dataAlignment: cie.DataAlignmentFactor,
		buf:           bytes.NewBuffer(initialInstructions),
	}

	frame.executeDwarfProgram()
	return frame
}


func executeDwarfProgramUntilPC(fde *FrameDescriptionEntry, pc uint64) *FrameContext {
	frame := executeCIEInstructions(fde.CIE)
	frame.order = fde.order
	frame.loc = fde.Begin()
	frame.address = pc
	frame.ExecuteUntilPC(fde.Instructions)

	return frame
}

func (frame *FrameContext) executeDwarfProgram() {
	for frame.buf.Len() > 0 {
		executeDwarfInstruction(frame)
	}
}


func (frame *FrameContext) ExecuteUntilPC(instructions []byte) {
	frame.buf.Truncate(0)
	frame.buf.Write(instructions)

	
	
	
	for frame.address >= frame.loc && frame.buf.Len() > 0 {
		executeDwarfInstruction(frame)
	}
}

func executeDwarfInstruction(frame *FrameContext) {
	instruction, err := frame.buf.ReadByte()
	if err != nil {
		panic("Could not read from instruction buffer")
	}

	if instruction == DW_CFA_nop {
		return
	}

	fn := lookupFunc(instruction, frame.buf)

	fn(frame)
}

func lookupFunc(instruction byte, buf *bytes.Buffer) instruction {
	const high_2_bits = 0xc0
	var restore bool

	
	switch instruction & high_2_bits {
	case DW_CFA_advance_loc:
		instruction = DW_CFA_advance_loc
		restore = true

	case DW_CFA_offset:
		instruction = DW_CFA_offset
		restore = true

	case DW_CFA_restore:
		instruction = DW_CFA_restore
		restore = true
	}

	if restore {
		
		err := buf.UnreadByte()
		if err != nil {
			panic("Could not unread byte")
		}
	}

	fn, ok := fnlookup[instruction]
	if !ok {
		panic(fmt.Sprintf("Encountered an unexpected DWARF CFA opcode: %#v", instruction))
	}

	return fn
}

func advanceloc(frame *FrameContext) {
	b, err := frame.buf.ReadByte()
	if err != nil {
		panic("Could not read byte")
	}

	delta := b & low_6_offset
	frame.loc += uint64(delta) * frame.codeAlignment
}

func advanceloc1(frame *FrameContext) {
	delta, err := frame.buf.ReadByte()
	if err != nil {
		panic("Could not read byte")
	}

	frame.loc += uint64(delta) * frame.codeAlignment
}

func advanceloc2(frame *FrameContext) {
	var delta uint16
	binary.Read(frame.buf, frame.order, &delta)

	frame.loc += uint64(delta) * frame.codeAlignment
}

func advanceloc4(frame *FrameContext) {
	var delta uint32
	binary.Read(frame.buf, frame.order, &delta)

	frame.loc += uint64(delta) * frame.codeAlignment
}

func offset(frame *FrameContext) {
	b, err := frame.buf.ReadByte()
	if err != nil {
		panic(err)
	}

	var (
		reg       = b & low_6_offset
		offset, _ = util.DecodeULEB128(frame.buf)
	)

	frame.Regs[uint64(reg)] = DWRule{Offset: int64(offset) * frame.dataAlignment, Rule: RuleOffset}
}

func restore(frame *FrameContext) {
	b, err := frame.buf.ReadByte()
	if err != nil {
		panic(err)
	}

	reg := uint64(b & low_6_offset)
	oldrule, ok := frame.initialRegs[reg]
	if ok {
		frame.Regs[reg] = DWRule{Offset: oldrule.Offset, Rule: RuleOffset}
	} else {
		frame.Regs[reg] = DWRule{Rule: RuleUndefined}
	}
}

func setloc(frame *FrameContext) {
	var loc uint64
	binary.Read(frame.buf, frame.order, &loc)

	frame.loc = loc + frame.cie.staticBase
}

func offsetextended(frame *FrameContext) {
	var (
		reg, _    = util.DecodeULEB128(frame.buf)
		offset, _ = util.DecodeULEB128(frame.buf)
	)

	frame.Regs[reg] = DWRule{Offset: int64(offset) * frame.dataAlignment, Rule: RuleOffset}
}

func undefined(frame *FrameContext) {
	reg, _ := util.DecodeULEB128(frame.buf)
	frame.Regs[reg] = DWRule{Rule: RuleUndefined}
}

func samevalue(frame *FrameContext) {
	reg, _ := util.DecodeULEB128(frame.buf)
	frame.Regs[reg] = DWRule{Rule: RuleSameVal}
}

func register(frame *FrameContext) {
	reg1, _ := util.DecodeULEB128(frame.buf)
	reg2, _ := util.DecodeULEB128(frame.buf)
	frame.Regs[reg1] = DWRule{Reg: reg2, Rule: RuleRegister}
}

func rememberstate(frame *FrameContext) {
	frame.prevRegs = frame.Regs
}

func restorestate(frame *FrameContext) {
	frame.Regs = frame.prevRegs
}

func restoreextended(frame *FrameContext) {
	reg, _ := util.DecodeULEB128(frame.buf)

	oldrule, ok := frame.initialRegs[reg]
	if ok {
		frame.Regs[reg] = DWRule{Offset: oldrule.Offset, Rule: RuleOffset}
	} else {
		frame.Regs[reg] = DWRule{Rule: RuleUndefined}
	}
}

func defcfa(frame *FrameContext) {
	reg, _ := util.DecodeULEB128(frame.buf)
	offset, _ := util.DecodeULEB128(frame.buf)

	frame.CFA.Rule = RuleCFA
	frame.CFA.Reg = reg
	frame.CFA.Offset = int64(offset)
}

func defcfaregister(frame *FrameContext) {
	reg, _ := util.DecodeULEB128(frame.buf)
	frame.CFA.Reg = reg
}

func defcfaoffset(frame *FrameContext) {
	offset, _ := util.DecodeULEB128(frame.buf)
	frame.CFA.Offset = int64(offset)
}

func defcfasf(frame *FrameContext) {
	reg, _ := util.DecodeULEB128(frame.buf)
	offset, _ := util.DecodeSLEB128(frame.buf)

	frame.CFA.Rule = RuleCFA
	frame.CFA.Reg = reg
	frame.CFA.Offset = offset * frame.dataAlignment
}

func defcfaoffsetsf(frame *FrameContext) {
	offset, _ := util.DecodeSLEB128(frame.buf)
	offset *= frame.dataAlignment
	frame.CFA.Offset = offset
}

func defcfaexpression(frame *FrameContext) {
	var (
		l, _ = util.DecodeULEB128(frame.buf)
		expr = frame.buf.Next(int(l))
	)

	frame.CFA.Expression = expr
	frame.CFA.Rule = RuleExpression
}

func expression(frame *FrameContext) {
	var (
		reg, _ = util.DecodeULEB128(frame.buf)
		l, _   = util.DecodeULEB128(frame.buf)
		expr   = frame.buf.Next(int(l))
	)

	frame.Regs[reg] = DWRule{Rule: RuleExpression, Expression: expr}
}

func offsetextendedsf(frame *FrameContext) {
	var (
		reg, _    = util.DecodeULEB128(frame.buf)
		offset, _ = util.DecodeSLEB128(frame.buf)
	)

	frame.Regs[reg] = DWRule{Offset: offset * frame.dataAlignment, Rule: RuleOffset}
}

func valoffset(frame *FrameContext) {
	var (
		reg, _    = util.DecodeULEB128(frame.buf)
		offset, _ = util.DecodeULEB128(frame.buf)
	)

	frame.Regs[reg] = DWRule{Offset: int64(offset), Rule: RuleValOffset}
}

func valoffsetsf(frame *FrameContext) {
	var (
		reg, _    = util.DecodeULEB128(frame.buf)
		offset, _ = util.DecodeSLEB128(frame.buf)
	)

	frame.Regs[reg] = DWRule{Offset: offset * frame.dataAlignment, Rule: RuleValOffset}
}

func valexpression(frame *FrameContext) {
	var (
		reg, _ = util.DecodeULEB128(frame.buf)
		l, _   = util.DecodeULEB128(frame.buf)
		expr   = frame.buf.Next(int(l))
	)

	frame.Regs[reg] = DWRule{Rule: RuleValExpression, Expression: expr}
}

func louser(frame *FrameContext) {
	frame.buf.Next(1)
}

func hiuser(frame *FrameContext) {
	frame.buf.Next(1)
}
