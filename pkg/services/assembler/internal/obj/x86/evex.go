// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package x86

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"errors"
	"fmt"
	"strings"
)


type evexBits struct {
	b1 byte 
	b2 byte 

	
	opcode byte
}


func newEVEXBits(z int, enc *opBytes) evexBits {
	return evexBits{
		b1:     enc[z+0],
		b2:     enc[z+1],
		opcode: enc[z+2],
	}
}


func (evex evexBits) P() byte { return (evex.b1 & evexP) >> 0 }


func (evex evexBits) L() byte { return (evex.b1 & evexL) >> 2 }


func (evex evexBits) M() byte { return (evex.b1 & evexM) >> 4 }


func (evex evexBits) W() byte { return (evex.b1 & evexW) >> 7 }


func (evex evexBits) BroadcastEnabled() bool {
	return evex.b2&evexBcst != 0
}


func (evex evexBits) ZeroingEnabled() bool {
	return (evex.b2&evexZeroing)>>2 != 0
}



func (evex evexBits) RoundingEnabled() bool {
	return (evex.b2&evexRounding)>>1 != 0
}


func (evex evexBits) SaeEnabled() bool {
	return (evex.b2&evexSae)>>0 != 0
}




func (evex evexBits) DispMultiplier(bcst bool) int32 {
	if bcst {
		switch evex.b2 & evexBcst {
		case evexBcstN4:
			return 4
		case evexBcstN8:
			return 8
		}
		return 1
	}

	switch evex.b2 & evexN {
	case evexN1:
		return 1
	case evexN2:
		return 2
	case evexN4:
		return 4
	case evexN8:
		return 8
	case evexN16:
		return 16
	case evexN32:
		return 32
	case evexN64:
		return 64
	case evexN128:
		return 128
	}
	return 1
}



const (
	evexW   = 0x80 
	evexWIG = 0 << 7
	evexW0  = 0 << 7
	evexW1  = 1 << 7

	evexM    = 0x30 
	evex0F   = 1 << 4
	evex0F38 = 2 << 4
	evex0F3A = 3 << 4

	evexL   = 0x0C 
	evexLIG = 0 << 2
	evex128 = 0 << 2
	evex256 = 1 << 2
	evex512 = 2 << 2

	evexP  = 0x03 
	evex66 = 1 << 0
	evexF3 = 2 << 0
	evexF2 = 3 << 0

	
	
	
	evexN    = 0xE0 
	evexN1   = 0 << 5
	evexN2   = 1 << 5
	evexN4   = 2 << 5
	evexN8   = 3 << 5
	evexN16  = 4 << 5
	evexN32  = 5 << 5
	evexN64  = 6 << 5
	evexN128 = 7 << 5

	
	evexBcst   = 0x18 
	evexBcstN4 = 1 << 3
	evexBcstN8 = 2 << 3

	
	
	evexZeroing         = 0x4 
	evexZeroingEnabled  = 1 << 2
	evexRounding        = 0x2 
	evexRoundingEnabled = 1 << 1
	evexSae             = 0x1 
	evexSaeEnabled      = 1 << 0
)


func compressedDisp8(disp, elemSize int32) (disp8 byte, ok bool) {
	if disp%elemSize == 0 {
		v := disp / elemSize
		if v >= -128 && v <= 127 {
			return byte(v), true
		}
	}
	return 0, false
}


func evexZcase(zcase uint8) bool {
	return zcase > Zevex_first && zcase < Zevex_last
}









type evexSuffix struct {
	rounding  byte
	sae       bool
	zeroing   bool
	broadcast bool
}



const (
	rcRNSAE = 0 
	rcRDSAE = 1 
	rcRUSAE = 2 
	rcRZSAE = 3 
	rcUnset = 4
)


func newEVEXSuffix() evexSuffix {
	return evexSuffix{rounding: rcUnset}
}



var evexSuffixMap [255]evexSuffix

func init() {
	
	for i := range opSuffixTable {
		suffix := newEVEXSuffix()
		parts := strings.Split(opSuffixTable[i], ".")
		for j := range parts {
			switch parts[j] {
			case "Z":
				suffix.zeroing = true
			case "BCST":
				suffix.broadcast = true
			case "SAE":
				suffix.sae = true

			case "RN_SAE":
				suffix.rounding = rcRNSAE
			case "RD_SAE":
				suffix.rounding = rcRDSAE
			case "RU_SAE":
				suffix.rounding = rcRUSAE
			case "RZ_SAE":
				suffix.rounding = rcRZSAE
			}
		}
		evexSuffixMap[i] = suffix
	}
}


func toDisp8(disp int32, p *obj.Prog, asmbuf *AsmBuf) (disp8 byte, ok bool) {
	if asmbuf.evexflag {
		bcst := evexSuffixMap[p.Scond].broadcast
		elemSize := asmbuf.evex.DispMultiplier(bcst)
		return compressedDisp8(disp, elemSize)
	}
	return byte(disp), disp >= -128 && disp < 128
}



func EncodeRegisterRange(reg0, reg1 int16) int64 {
	return (int64(reg0) << 0) |
		(int64(reg1) << 16) |
		obj.RegListX86Lo
}


func decodeRegisterRange(list int64) (reg0, reg1 int) {
	return int((list >> 0) & 0xFFFF),
		int((list >> 16) & 0xFFFF)
}





func ParseSuffix(p *obj.Prog, cond string) error {
	cond = strings.TrimPrefix(cond, ".")

	suffix := newOpSuffix(cond)
	if !suffix.IsValid() {
		return inferSuffixError(cond)
	}

	p.Scond = uint8(suffix)
	return nil
}












func inferSuffixError(cond string) error {
	suffixSet := make(map[string]bool)  
	unknownSet := make(map[string]bool) 
	hasBcst := false
	hasRoundSae := false
	var msg []string 

	suffixes := strings.Split(cond, ".")
	for i, suffix := range suffixes {
		switch suffix {
		case "Z":
			if i != len(suffixes)-1 {
				msg = append(msg, "Z suffix should be the last")
			}
		case "BCST":
			hasBcst = true
		case "SAE", "RN_SAE", "RZ_SAE", "RD_SAE", "RU_SAE":
			hasRoundSae = true
		default:
			if !unknownSet[suffix] {
				msg = append(msg, fmt.Sprintf("unknown suffix %q", suffix))
			}
			unknownSet[suffix] = true
		}

		if suffixSet[suffix] {
			msg = append(msg, fmt.Sprintf("duplicate suffix %q", suffix))
		}
		suffixSet[suffix] = true
	}

	if hasBcst && hasRoundSae {
		msg = append(msg, "can't combine rounding/SAE and broadcast")
	}

	if len(msg) == 0 {
		return errors.New("bad suffix combination")
	}
	return errors.New(strings.Join(msg, "; "))
}




var opSuffixTable = [...]string{
	"", 

	"Z",

	"SAE",
	"SAE.Z",

	"RN_SAE",
	"RZ_SAE",
	"RD_SAE",
	"RU_SAE",
	"RN_SAE.Z",
	"RZ_SAE.Z",
	"RD_SAE.Z",
	"RU_SAE.Z",

	"BCST",
	"BCST.Z",

	"<bad suffix>",
}





type opSuffix uint8


const badOpSuffix = opSuffix(len(opSuffixTable) - 1)





func newOpSuffix(suffixes string) opSuffix {
	for i := range opSuffixTable {
		if opSuffixTable[i] == suffixes {
			return opSuffix(i)
		}
	}
	return badOpSuffix
}



func (suffix opSuffix) IsValid() bool {
	return suffix != badOpSuffix
}






func (suffix opSuffix) String() string {
	return opSuffixTable[suffix]
}
