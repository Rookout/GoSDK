// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the ARM
// instruction set, to minimize its interaction with the core of the
// assembler.

package arch

import (
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/arm"
)

var armLS = map[string]uint8{
	"U":  arm.C_UBIT,
	"S":  arm.C_SBIT,
	"W":  arm.C_WBIT,
	"P":  arm.C_PBIT,
	"PW": arm.C_WBIT | arm.C_PBIT,
	"WP": arm.C_WBIT | arm.C_PBIT,
}

var armSCOND = map[string]uint8{
	"EQ":  arm.C_SCOND_EQ,
	"NE":  arm.C_SCOND_NE,
	"CS":  arm.C_SCOND_HS,
	"HS":  arm.C_SCOND_HS,
	"CC":  arm.C_SCOND_LO,
	"LO":  arm.C_SCOND_LO,
	"MI":  arm.C_SCOND_MI,
	"PL":  arm.C_SCOND_PL,
	"VS":  arm.C_SCOND_VS,
	"VC":  arm.C_SCOND_VC,
	"HI":  arm.C_SCOND_HI,
	"LS":  arm.C_SCOND_LS,
	"GE":  arm.C_SCOND_GE,
	"LT":  arm.C_SCOND_LT,
	"GT":  arm.C_SCOND_GT,
	"LE":  arm.C_SCOND_LE,
	"AL":  arm.C_SCOND_NONE,
	"U":   arm.C_UBIT,
	"S":   arm.C_SBIT,
	"W":   arm.C_WBIT,
	"P":   arm.C_PBIT,
	"PW":  arm.C_WBIT | arm.C_PBIT,
	"WP":  arm.C_WBIT | arm.C_PBIT,
	"F":   arm.C_FBIT,
	"IBW": arm.C_WBIT | arm.C_PBIT | arm.C_UBIT,
	"IAW": arm.C_WBIT | arm.C_UBIT,
	"DBW": arm.C_WBIT | arm.C_PBIT,
	"DAW": arm.C_WBIT,
	"IB":  arm.C_PBIT | arm.C_UBIT,
	"IA":  arm.C_UBIT,
	"DB":  arm.C_PBIT,
	"DA":  0,
}

var armJump = map[string]bool{
	"B":    true,
	"BL":   true,
	"BX":   true,
	"BEQ":  true,
	"BNE":  true,
	"BCS":  true,
	"BHS":  true,
	"BCC":  true,
	"BLO":  true,
	"BMI":  true,
	"BPL":  true,
	"BVS":  true,
	"BVC":  true,
	"BHI":  true,
	"BLS":  true,
	"BGE":  true,
	"BLT":  true,
	"BGT":  true,
	"BLE":  true,
	"CALL": true,
	"JMP":  true,
}

func jumpArm(word string) bool {
	return armJump[word]
}



func IsARMCMP(op obj.As) bool {
	switch op {
	case arm.ACMN, arm.ACMP, arm.ATEQ, arm.ATST:
		return true
	}
	return false
}



func IsARMSTREX(op obj.As) bool {
	switch op {
	case arm.ASTREX, arm.ASTREXD, arm.ASWPW, arm.ASWPBU:
		return true
	}
	return false
}




const aMCR = arm.ALAST + 1



func IsARMMRC(op obj.As) bool {
	switch op {
	case arm.AMRC, aMCR: 
		return true
	}
	return false
}



func IsARMBFX(op obj.As) bool {
	switch op {
	case arm.ABFX, arm.ABFXU, arm.ABFC, arm.ABFI:
		return true
	}
	return false
}


func IsARMFloatCmp(op obj.As) bool {
	switch op {
	case arm.ACMPF, arm.ACMPD:
		return true
	}
	return false
}





func ARMMRCOffset(op obj.As, cond string, x0, x1, x2, x3, x4, x5 int64) (offset int64, op0 obj.As, ok bool) {
	op1 := int64(0)
	if op == arm.AMRC {
		op1 = 1
	}
	bits, ok := ParseARMCondition(cond)
	if !ok {
		return
	}
	offset = (0xe << 24) | 
		(op1 << 20) | 
		((int64(bits) ^ arm.C_SCOND_XOR) << 28) | 
		((x0 & 15) << 8) | 
		((x1 & 7) << 21) | 
		((x2 & 15) << 12) | 
		((x3 & 15) << 16) | 
		((x4 & 15) << 0) | 
		((x5 & 7) << 5) | 
		(1 << 4) 
	return offset, arm.AMRC, true
}



func IsARMMULA(op obj.As) bool {
	switch op {
	case arm.AMULA, arm.AMULS, arm.AMMULA, arm.AMMULS, arm.AMULABB, arm.AMULAWB, arm.AMULAWT:
		return true
	}
	return false
}

var bcode = []obj.As{
	arm.ABEQ,
	arm.ABNE,
	arm.ABCS,
	arm.ABCC,
	arm.ABMI,
	arm.ABPL,
	arm.ABVS,
	arm.ABVC,
	arm.ABHI,
	arm.ABLS,
	arm.ABGE,
	arm.ABLT,
	arm.ABGT,
	arm.ABLE,
	arm.AB,
	obj.ANOP,
}



func ARMConditionCodes(prog *obj.Prog, cond string) bool {
	if cond == "" {
		return true
	}
	bits, ok := ParseARMCondition(cond)
	if !ok {
		return false
	}
	
	if prog.As == arm.AB {
		prog.As = bcode[(bits^arm.C_SCOND_XOR)&0xf]
		bits = (bits &^ 0xf) | arm.C_SCOND_NONE
	}
	prog.Scond = bits
	return true
}




func ParseARMCondition(cond string) (uint8, bool) {
	return parseARMCondition(cond, armLS, armSCOND)
}

func parseARMCondition(cond string, ls, scond map[string]uint8) (uint8, bool) {
	cond = strings.TrimPrefix(cond, ".")
	if cond == "" {
		return arm.C_SCOND_NONE, true
	}
	names := strings.Split(cond, ".")
	bits := uint8(0)
	for _, name := range names {
		if b, present := ls[name]; present {
			bits |= b
			continue
		}
		if b, present := scond[name]; present {
			bits = (bits &^ arm.C_SCOND) | b
			continue
		}
		return 0, false
	}
	return bits, true
}

func armRegisterNumber(name string, n int16) (int16, bool) {
	if n < 0 || 15 < n {
		return 0, false
	}
	switch name {
	case "R":
		return arm.REG_R0 + n, true
	case "F":
		return arm.REG_F0 + n, true
	}
	return 0, false
}
