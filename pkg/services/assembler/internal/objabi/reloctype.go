// Derived from Inferno utils/6l/l.h and related files.
// https://bitbucket.org/inferno-os/inferno-os/src/master/utils/6l/l.h
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package objabi

type RelocType int16

//go:generate stringer -type=RelocType
const (
	R_ADDR RelocType = 1 + iota
	
	
	
	
	
	R_ADDRPOWER
	
	
	R_ADDRARM64
	
	
	R_ADDRMIPS
	
	
	R_ADDROFF
	R_SIZE
	R_CALL
	R_CALLARM
	R_CALLARM64
	R_CALLIND
	R_CALLPOWER
	
	
	R_CALLMIPS
	R_CONST
	R_PCREL
	
	
	
	
	R_TLS_LE
	
	
	
	
	
	R_TLS_IE
	R_GOTOFF
	R_PLT0
	R_PLT1
	R_PLT2
	R_USEFIELD
	
	
	
	
	R_USETYPE
	
	
	
	
	
	R_USEIFACE
	
	
	
	
	
	R_USEIFACEMETHOD
	
	
	
	
	R_USEGENERICIFACEMETHOD
	
	
	
	
	
	R_METHODOFF
	
	
	R_KEEP
	R_POWER_TOC
	R_GOTPCREL
	
	
	
	R_JMPMIPS

	
	
	R_DWARFSECREF

	
	
	
	R_DWARFFILEREF

	
	
	
	
	
	

	

	
	
	
	R_ARM64_TLS_LE

	
	
	
	R_ARM64_TLS_IE

	
	
	R_ARM64_GOTPCREL

	
	
	R_ARM64_GOT

	
	
	R_ARM64_PCREL

	
	
	R_ARM64_PCREL_LDST8

	
	
	R_ARM64_PCREL_LDST16

	
	
	R_ARM64_PCREL_LDST32

	
	
	R_ARM64_PCREL_LDST64

	
	R_ARM64_LDST8

	
	R_ARM64_LDST16

	
	R_ARM64_LDST32

	
	R_ARM64_LDST64

	
	R_ARM64_LDST128

	

	
	
	
	
	R_POWER_TLS_LE

	
	
	
	
	
	R_POWER_TLS_IE

	
	
	
	
	
	
	R_POWER_TLS

	
	
	
	R_POWER_TLS_IE_PCREL34

	
	
	R_POWER_TLS_LE_TPREL34

	
	
	
	
	
	R_ADDRPOWER_DS

	
	
	R_ADDRPOWER_GOT

	
	
	R_ADDRPOWER_GOT_PCREL34

	
	
	
	R_ADDRPOWER_PCREL

	
	
	
	R_ADDRPOWER_TOCREL

	
	
	
	R_ADDRPOWER_TOCREL_DS

	
	
	
	R_ADDRPOWER_D34

	
	
	
	R_ADDRPOWER_PCREL34

	

	
	
	R_RISCV_CALL

	
	
	
	R_RISCV_CALL_TRAMP

	
	
	R_RISCV_PCREL_ITYPE

	
	
	R_RISCV_PCREL_STYPE

	
	
	R_RISCV_TLS_IE_ITYPE

	
	
	R_RISCV_TLS_IE_STYPE

	
	
	R_PCRELDBL

	

	
	
	R_ADDRLOONG64

	
	
	R_ADDRLOONG64U

	
	
	R_ADDRLOONG64TLS

	
	
	R_ADDRLOONG64TLSU

	
	
	R_CALLLOONG64

	
	
	R_LOONG64_TLS_IE_PCREL_HI
	R_LOONG64_TLS_IE_LO

	
	
	R_JMPLOONG64

	
	
	R_ADDRMIPSU
	
	
	R_ADDRMIPSTLS

	
	
	R_ADDRCUOFF

	
	R_WASMIMPORT

	
	
	
	R_XCOFFREF

	
	
	R_PEIMAGEOFF

	
	
	
	
	R_INITORDER

	
	
	
	
	R_WEAK = -1 << 15

	R_WEAKADDR    = R_WEAK | R_ADDR
	R_WEAKADDROFF = R_WEAK | R_ADDROFF
)






func (r RelocType) IsDirectCall() bool {
	switch r {
	case R_CALL, R_CALLARM, R_CALLARM64, R_CALLLOONG64, R_CALLMIPS, R_CALLPOWER, R_RISCV_CALL, R_RISCV_CALL_TRAMP:
		return true
	}
	return false
}






func (r RelocType) IsDirectJump() bool {
	switch r {
	case R_JMPMIPS:
		return true
	case R_JMPLOONG64:
		return true
	}
	return false
}



func (r RelocType) IsDirectCallOrJump() bool {
	return r.IsDirectCall() || r.IsDirectJump()
}
