// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

// Package arch defines architecture-specific information and support functions.
package arch

import (
	"fmt"
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/arm"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/arm64"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj/x86"
)


const (
	RFP = -(iota + 1)
	RSB
	RSP
	RPC
)


type Arch struct {
	*obj.LinkArch
	
	Instructions map[string]obj.As
	
	Register map[string]int16
	
	RegisterPrefix map[string]bool
	
	RegisterNumber func(string, int16) (int16, bool)
	
	IsJump func(word string) bool
}



func nilRegisterNumber(name string, n int16) (int16, bool) {
	return 0, false
}



func Set(GOARCH string, shared bool) *Arch {
	switch GOARCH {
	case "386":
		return archX86(&x86.Link386)
	case "amd64":
		return archX86(&x86.Linkamd64)
	case "arm":
		return archArm()
	case "arm64":
		return archArm64()
	}
	return nil
}

func jumpX86(word string) bool {
	return word[0] == 'J' || word == "CALL" || strings.HasPrefix(word, "LOOP") || word == "XBEGIN"
}

func archX86(linkArch *obj.LinkArch) *Arch {
	register := make(map[string]int16)
	
	for i, s := range x86.Register {
		register[s] = int16(i + x86.REG_AL)
	}
	
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	if linkArch == &x86.Linkamd64 {
		
		register["g"] = x86.REGG
	}
	

	instructions := make(map[string]obj.As)
	for i, s := range obj.Anames {
		instructions[s] = obj.As(i)
	}
	for i, s := range x86.Anames {
		if obj.As(i) >= obj.A_ARCHSPECIFIC {
			instructions[s] = obj.As(i) + obj.ABaseAMD64
		}
	}
	
	instructions["JA"] = x86.AJHI   
	instructions["JAE"] = x86.AJCC  
	instructions["JB"] = x86.AJCS   
	instructions["JBE"] = x86.AJLS  
	instructions["JC"] = x86.AJCS   
	instructions["JCC"] = x86.AJCC  
	instructions["JCS"] = x86.AJCS  
	instructions["JE"] = x86.AJEQ   
	instructions["JEQ"] = x86.AJEQ  
	instructions["JG"] = x86.AJGT   
	instructions["JGE"] = x86.AJGE  
	instructions["JGT"] = x86.AJGT  
	instructions["JHI"] = x86.AJHI  
	instructions["JHS"] = x86.AJCC  
	instructions["JL"] = x86.AJLT   
	instructions["JLE"] = x86.AJLE  
	instructions["JLO"] = x86.AJCS  
	instructions["JLS"] = x86.AJLS  
	instructions["JLT"] = x86.AJLT  
	instructions["JMI"] = x86.AJMI  
	instructions["JNA"] = x86.AJLS  
	instructions["JNAE"] = x86.AJCS 
	instructions["JNB"] = x86.AJCC  
	instructions["JNBE"] = x86.AJHI 
	instructions["JNC"] = x86.AJCC  
	instructions["JNE"] = x86.AJNE  
	instructions["JNG"] = x86.AJLE  
	instructions["JNGE"] = x86.AJLT 
	instructions["JNL"] = x86.AJGE  
	instructions["JNLE"] = x86.AJGT 
	instructions["JNO"] = x86.AJOC  
	instructions["JNP"] = x86.AJPC  
	instructions["JNS"] = x86.AJPL  
	instructions["JNZ"] = x86.AJNE  
	instructions["JO"] = x86.AJOS   
	instructions["JOC"] = x86.AJOC  
	instructions["JOS"] = x86.AJOS  
	instructions["JP"] = x86.AJPS   
	instructions["JPC"] = x86.AJPC  
	instructions["JPE"] = x86.AJPS  
	instructions["JPL"] = x86.AJPL  
	instructions["JPO"] = x86.AJPC  
	instructions["JPS"] = x86.AJPS  
	instructions["JS"] = x86.AJMI   
	instructions["JZ"] = x86.AJEQ   
	instructions["MASKMOVDQU"] = x86.AMASKMOVOU
	instructions["MOVD"] = x86.AMOVQ
	instructions["MOVDQ2Q"] = x86.AMOVQ
	instructions["MOVNTDQ"] = x86.AMOVNTO
	instructions["MOVOA"] = x86.AMOVO
	instructions["PSLLDQ"] = x86.APSLLO
	instructions["PSRLDQ"] = x86.APSRLO
	instructions["PADDD"] = x86.APADDL
	
	instructions["MOVBELL"] = x86.AMOVBEL
	instructions["MOVBEQQ"] = x86.AMOVBEQ
	instructions["MOVBEWW"] = x86.AMOVBEW

	return &Arch{
		LinkArch:       linkArch,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: nil,
		RegisterNumber: nilRegisterNumber,
		IsJump:         jumpX86,
	}
}

func archArm() *Arch {
	register := make(map[string]int16)
	
	
	for i := arm.REG_R0; i < arm.REG_SPSR; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	
	delete(register, "R10")
	register["g"] = arm.REG_R10
	for i := 0; i < 16; i++ {
		register[fmt.Sprintf("C%d", i)] = int16(i)
	}

	
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	register["SP"] = RSP
	registerPrefix := map[string]bool{
		"F": true,
		"R": true,
	}

	
	register["MB_SY"] = arm.REG_MB_SY
	register["MB_ST"] = arm.REG_MB_ST
	register["MB_ISH"] = arm.REG_MB_ISH
	register["MB_ISHST"] = arm.REG_MB_ISHST
	register["MB_NSH"] = arm.REG_MB_NSH
	register["MB_NSHST"] = arm.REG_MB_NSHST
	register["MB_OSH"] = arm.REG_MB_OSH
	register["MB_OSHST"] = arm.REG_MB_OSHST

	instructions := make(map[string]obj.As)
	for i, s := range obj.Anames {
		instructions[s] = obj.As(i)
	}
	for i, s := range arm.Anames {
		if obj.As(i) >= obj.A_ARCHSPECIFIC {
			instructions[s] = obj.As(i) + obj.ABaseARM
		}
	}
	
	instructions["B"] = obj.AJMP
	instructions["BL"] = obj.ACALL
	
	
	
	instructions["MCR"] = aMCR

	return &Arch{
		LinkArch:       &arm.Linkarm,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: registerPrefix,
		RegisterNumber: armRegisterNumber,
		IsJump:         jumpArm,
	}
}

func archArm64() *Arch {
	register := make(map[string]int16)
	
	
	register[obj.Rconv(arm64.REGSP)] = int16(arm64.REGSP)
	for i := arm64.REG_R0; i <= arm64.REG_R31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	
	register["R18_PLATFORM"] = register["R18"]
	delete(register, "R18")
	for i := arm64.REG_F0; i <= arm64.REG_F31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}
	for i := arm64.REG_V0; i <= arm64.REG_V31; i++ {
		register[obj.Rconv(i)] = int16(i)
	}

	
	for i := 0; i < len(arm64.SystemReg); i++ {
		register[arm64.SystemReg[i].Name] = arm64.SystemReg[i].Reg
	}

	register["LR"] = arm64.REGLINK

	
	register["SB"] = RSB
	register["FP"] = RFP
	register["PC"] = RPC
	register["SP"] = RSP
	
	delete(register, "R28")
	register["g"] = arm64.REG_R28
	registerPrefix := map[string]bool{
		"F": true,
		"R": true,
		"V": true,
	}

	instructions := make(map[string]obj.As)
	for i, s := range obj.Anames {
		instructions[s] = obj.As(i)
	}
	for i, s := range arm64.Anames {
		if obj.As(i) >= obj.A_ARCHSPECIFIC {
			instructions[s] = obj.As(i) + obj.ABaseARM64
		}
	}
	
	instructions["B"] = arm64.AB
	instructions["BL"] = arm64.ABL

	return &Arch{
		LinkArch:       &arm64.Linkarm64,
		Instructions:   instructions,
		Register:       register,
		RegisterPrefix: registerPrefix,
		RegisterNumber: arm64RegisterNumber,
		IsJump:         jumpArm64,
	}

}
