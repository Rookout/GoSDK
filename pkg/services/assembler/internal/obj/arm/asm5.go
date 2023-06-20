// Inferno utils/5l/span.c
// https://bitbucket.org/inferno-os/inferno-os/src/master/utils/5l/span.c
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

package arm

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"
	"log"
	"math"
	"sort"
)




type ctxt5 struct {
	ctxt       *obj.Link
	newprog    obj.ProgAlloc
	cursym     *obj.LSym
	printp     *obj.Prog
	blitrl     *obj.Prog
	elitrl     *obj.Prog
	autosize   int64
	instoffset int64
	pc         int64
	pool       struct {
		start uint32
		size  uint32
		extra uint32
	}
}

type Optab struct {
	as       obj.As
	a1       uint8
	a2       int8
	a3       uint8
	type_    uint8
	size     int8
	param    int16
	flag     int8
	pcrelsiz uint8
	scond    uint8 
}

type Opcross [32][2][32]uint8

const (
	LFROM  = 1 << 0
	LTO    = 1 << 1
	LPOOL  = 1 << 2
	LPCREL = 1 << 3
)

var optab = []Optab{
	
	{obj.ATEXT, C_ADDR, C_NONE, C_TEXTSIZE, 0, 0, 0, 0, 0, 0},
	{AADD, C_REG, C_REG, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AADD, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AAND, C_REG, C_REG, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AAND, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AORR, C_REG, C_REG, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AORR, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AMOVW, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{AMVN, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, C_SBIT},
	{ACMP, C_REG, C_REG, C_NONE, 1, 4, 0, 0, 0, 0},
	{AADD, C_RCON, C_REG, C_REG, 2, 4, 0, 0, 0, C_SBIT},
	{AADD, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0, C_SBIT},
	{AAND, C_RCON, C_REG, C_REG, 2, 4, 0, 0, 0, C_SBIT},
	{AAND, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0, C_SBIT},
	{AORR, C_RCON, C_REG, C_REG, 2, 4, 0, 0, 0, C_SBIT},
	{AORR, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0, C_SBIT},
	{AMOVW, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0, 0},
	{AMVN, C_RCON, C_NONE, C_REG, 2, 4, 0, 0, 0, 0},
	{ACMP, C_RCON, C_REG, C_NONE, 2, 4, 0, 0, 0, 0},
	{AADD, C_SHIFT, C_REG, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{AADD, C_SHIFT, C_NONE, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{AAND, C_SHIFT, C_REG, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{AAND, C_SHIFT, C_NONE, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{AORR, C_SHIFT, C_REG, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{AORR, C_SHIFT, C_NONE, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{AMVN, C_SHIFT, C_NONE, C_REG, 3, 4, 0, 0, 0, C_SBIT},
	{ACMP, C_SHIFT, C_REG, C_NONE, 3, 4, 0, 0, 0, 0},
	{AMOVW, C_RACON, C_NONE, C_REG, 4, 4, REGSP, 0, 0, C_SBIT},
	{AB, C_NONE, C_NONE, C_SBRA, 5, 4, 0, LPOOL, 0, 0},
	{ABL, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0, 0},
	{ABX, C_NONE, C_NONE, C_SBRA, 74, 20, 0, 0, 0, 0},
	{ABEQ, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0, 0},
	{ABEQ, C_RCON, C_NONE, C_SBRA, 5, 4, 0, 0, 0, 0}, 
	{AB, C_NONE, C_NONE, C_ROREG, 6, 4, 0, LPOOL, 0, 0},
	{ABL, C_NONE, C_NONE, C_ROREG, 7, 4, 0, 0, 0, 0},
	{ABL, C_REG, C_NONE, C_ROREG, 7, 4, 0, 0, 0, 0},
	{ABX, C_NONE, C_NONE, C_ROREG, 75, 12, 0, 0, 0, 0},
	{ABXRET, C_NONE, C_NONE, C_ROREG, 76, 4, 0, 0, 0, 0},
	{ASLL, C_RCON, C_REG, C_REG, 8, 4, 0, 0, 0, C_SBIT},
	{ASLL, C_RCON, C_NONE, C_REG, 8, 4, 0, 0, 0, C_SBIT},
	{ASLL, C_REG, C_NONE, C_REG, 9, 4, 0, 0, 0, C_SBIT},
	{ASLL, C_REG, C_REG, C_REG, 9, 4, 0, 0, 0, C_SBIT},
	{ASWI, C_NONE, C_NONE, C_NONE, 10, 4, 0, 0, 0, 0},
	{ASWI, C_NONE, C_NONE, C_LCON, 10, 4, 0, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_LCON, 11, 4, 0, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_LCONADDR, 11, 4, 0, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_ADDR, 11, 4, 0, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_TLS_LE, 103, 4, 0, 0, 0, 0},
	{AWORD, C_NONE, C_NONE, C_TLS_IE, 104, 4, 0, 0, 0, 0},
	{AMOVW, C_NCON, C_NONE, C_REG, 12, 4, 0, 0, 0, 0},
	{AMOVW, C_SCON, C_NONE, C_REG, 12, 4, 0, 0, 0, 0},
	{AMOVW, C_LCON, C_NONE, C_REG, 12, 4, 0, LFROM, 0, 0},
	{AMOVW, C_LCONADDR, C_NONE, C_REG, 12, 4, 0, LFROM | LPCREL, 4, 0},
	{AMVN, C_NCON, C_NONE, C_REG, 12, 4, 0, 0, 0, 0},
	{AADD, C_NCON, C_REG, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AADD, C_NCON, C_NONE, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AAND, C_NCON, C_REG, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AAND, C_NCON, C_NONE, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AORR, C_NCON, C_REG, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AORR, C_NCON, C_NONE, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{ACMP, C_NCON, C_REG, C_NONE, 13, 8, 0, 0, 0, 0},
	{AADD, C_SCON, C_REG, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AADD, C_SCON, C_NONE, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AAND, C_SCON, C_REG, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AAND, C_SCON, C_NONE, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AORR, C_SCON, C_REG, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AORR, C_SCON, C_NONE, C_REG, 13, 8, 0, 0, 0, C_SBIT},
	{AMVN, C_SCON, C_NONE, C_REG, 13, 8, 0, 0, 0, 0},
	{ACMP, C_SCON, C_REG, C_NONE, 13, 8, 0, 0, 0, 0},
	{AADD, C_RCON2A, C_REG, C_REG, 106, 8, 0, 0, 0, 0},
	{AADD, C_RCON2A, C_NONE, C_REG, 106, 8, 0, 0, 0, 0},
	{AORR, C_RCON2A, C_REG, C_REG, 106, 8, 0, 0, 0, 0},
	{AORR, C_RCON2A, C_NONE, C_REG, 106, 8, 0, 0, 0, 0},
	{AADD, C_RCON2S, C_REG, C_REG, 107, 8, 0, 0, 0, 0},
	{AADD, C_RCON2S, C_NONE, C_REG, 107, 8, 0, 0, 0, 0},
	{AADD, C_LCON, C_REG, C_REG, 13, 8, 0, LFROM, 0, C_SBIT},
	{AADD, C_LCON, C_NONE, C_REG, 13, 8, 0, LFROM, 0, C_SBIT},
	{AAND, C_LCON, C_REG, C_REG, 13, 8, 0, LFROM, 0, C_SBIT},
	{AAND, C_LCON, C_NONE, C_REG, 13, 8, 0, LFROM, 0, C_SBIT},
	{AORR, C_LCON, C_REG, C_REG, 13, 8, 0, LFROM, 0, C_SBIT},
	{AORR, C_LCON, C_NONE, C_REG, 13, 8, 0, LFROM, 0, C_SBIT},
	{AMVN, C_LCON, C_NONE, C_REG, 13, 8, 0, LFROM, 0, 0},
	{ACMP, C_LCON, C_REG, C_NONE, 13, 8, 0, LFROM, 0, 0},
	{AMOVB, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, 0},
	{AMOVBS, C_REG, C_NONE, C_REG, 14, 8, 0, 0, 0, 0},
	{AMOVBU, C_REG, C_NONE, C_REG, 58, 4, 0, 0, 0, 0},
	{AMOVH, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, 0},
	{AMOVHS, C_REG, C_NONE, C_REG, 14, 8, 0, 0, 0, 0},
	{AMOVHU, C_REG, C_NONE, C_REG, 14, 8, 0, 0, 0, 0},
	{AMUL, C_REG, C_REG, C_REG, 15, 4, 0, 0, 0, C_SBIT},
	{AMUL, C_REG, C_NONE, C_REG, 15, 4, 0, 0, 0, C_SBIT},
	{ADIV, C_REG, C_REG, C_REG, 16, 4, 0, 0, 0, 0},
	{ADIV, C_REG, C_NONE, C_REG, 16, 4, 0, 0, 0, 0},
	{ADIVHW, C_REG, C_REG, C_REG, 105, 4, 0, 0, 0, 0},
	{ADIVHW, C_REG, C_NONE, C_REG, 105, 4, 0, 0, 0, 0},
	{AMULL, C_REG, C_REG, C_REGREG, 17, 4, 0, 0, 0, C_SBIT},
	{ABFX, C_LCON, C_REG, C_REG, 18, 4, 0, 0, 0, 0},  
	{ABFX, C_LCON, C_NONE, C_REG, 18, 4, 0, 0, 0, 0}, 
	{AMOVW, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_REG, C_NONE, C_SAUTO, 20, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_REG, C_NONE, C_SOREG, 20, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_SAUTO, C_NONE, C_REG, 21, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_SOREG, C_NONE, C_REG, 21, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_SAUTO, C_NONE, C_REG, 21, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_SOREG, C_NONE, C_REG, 21, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AXTAB, C_SHIFT, C_REG, C_REG, 22, 4, 0, 0, 0, 0},
	{AXTAB, C_SHIFT, C_NONE, C_REG, 22, 4, 0, 0, 0, 0},
	{AMOVW, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, C_SBIT},
	{AMOVB, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, 0},
	{AMOVBS, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, 0},
	{AMOVBU, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, 0},
	{AMOVH, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, 0},
	{AMOVHS, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, 0},
	{AMOVHU, C_SHIFT, C_NONE, C_REG, 23, 4, 0, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_REG, C_NONE, C_LAUTO, 30, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_REG, C_NONE, C_LOREG, 30, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_REG, C_NONE, C_ADDR, 64, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_TLS_LE, C_NONE, C_REG, 101, 4, 0, LFROM, 0, 0},
	{AMOVW, C_TLS_IE, C_NONE, C_REG, 102, 8, 0, LFROM, 0, 0},
	{AMOVW, C_LAUTO, C_NONE, C_REG, 31, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_LOREG, C_NONE, C_REG, 31, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_ADDR, C_NONE, C_REG, 65, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_LAUTO, C_NONE, C_REG, 31, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_LOREG, C_NONE, C_REG, 31, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_ADDR, C_NONE, C_REG, 65, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_LACON, C_NONE, C_REG, 34, 8, REGSP, LFROM, 0, C_SBIT},
	{AMOVW, C_PSR, C_NONE, C_REG, 35, 4, 0, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_PSR, 36, 4, 0, 0, 0, 0},
	{AMOVW, C_RCON, C_NONE, C_PSR, 37, 4, 0, 0, 0, 0},
	{AMOVM, C_REGLIST, C_NONE, C_SOREG, 38, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVM, C_SOREG, C_NONE, C_REGLIST, 39, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{ASWPW, C_SOREG, C_REG, C_REG, 40, 4, 0, 0, 0, 0},
	{ARFE, C_NONE, C_NONE, C_NONE, 41, 4, 0, 0, 0, 0},
	{AMOVF, C_FREG, C_NONE, C_FAUTO, 50, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_FREG, C_NONE, C_FOREG, 50, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_FAUTO, C_NONE, C_FREG, 51, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_FOREG, C_NONE, C_FREG, 51, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_FREG, C_NONE, C_LAUTO, 52, 12, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_FREG, C_NONE, C_LOREG, 52, 12, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_LAUTO, C_NONE, C_FREG, 53, 12, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_LOREG, C_NONE, C_FREG, 53, 12, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_FREG, C_NONE, C_ADDR, 68, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVF, C_ADDR, C_NONE, C_FREG, 69, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AADDF, C_FREG, C_NONE, C_FREG, 54, 4, 0, 0, 0, 0},
	{AADDF, C_FREG, C_FREG, C_FREG, 54, 4, 0, 0, 0, 0},
	{AMOVF, C_FREG, C_NONE, C_FREG, 55, 4, 0, 0, 0, 0},
	{ANEGF, C_FREG, C_NONE, C_FREG, 55, 4, 0, 0, 0, 0},
	{AMOVW, C_REG, C_NONE, C_FCR, 56, 4, 0, 0, 0, 0},
	{AMOVW, C_FCR, C_NONE, C_REG, 57, 4, 0, 0, 0, 0},
	{AMOVW, C_SHIFTADDR, C_NONE, C_REG, 59, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_SHIFTADDR, C_NONE, C_REG, 59, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_SHIFTADDR, C_NONE, C_REG, 60, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_SHIFTADDR, C_NONE, C_REG, 60, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_SHIFTADDR, C_NONE, C_REG, 60, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_SHIFTADDR, C_NONE, C_REG, 60, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_SHIFTADDR, C_NONE, C_REG, 60, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVW, C_REG, C_NONE, C_SHIFTADDR, 61, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_REG, C_NONE, C_SHIFTADDR, 61, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_REG, C_NONE, C_SHIFTADDR, 61, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBU, C_REG, C_NONE, C_SHIFTADDR, 61, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_REG, C_NONE, C_SHIFTADDR, 62, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_REG, C_NONE, C_SHIFTADDR, 62, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_REG, C_NONE, C_SHIFTADDR, 62, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_REG, C_NONE, C_HAUTO, 70, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_REG, C_NONE, C_HOREG, 70, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_REG, C_NONE, C_HAUTO, 70, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_REG, C_NONE, C_HOREG, 70, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_REG, C_NONE, C_HAUTO, 70, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_REG, C_NONE, C_HOREG, 70, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_HAUTO, C_NONE, C_REG, 71, 4, REGSP, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_HOREG, C_NONE, C_REG, 71, 4, 0, 0, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_REG, C_NONE, C_LAUTO, 72, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_REG, C_NONE, C_LOREG, 72, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_REG, C_NONE, C_ADDR, 94, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_REG, C_NONE, C_LAUTO, 72, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_REG, C_NONE, C_LOREG, 72, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_REG, C_NONE, C_ADDR, 94, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_REG, C_NONE, C_LAUTO, 72, 8, REGSP, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_REG, C_NONE, C_LOREG, 72, 8, 0, LTO, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_REG, C_NONE, C_ADDR, 94, 8, 0, LTO | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVB, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVBS, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVH, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHS, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_LAUTO, C_NONE, C_REG, 73, 8, REGSP, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_LOREG, C_NONE, C_REG, 73, 8, 0, LFROM, 0, C_PBIT | C_WBIT | C_UBIT},
	{AMOVHU, C_ADDR, C_NONE, C_REG, 93, 8, 0, LFROM | LPCREL, 4, C_PBIT | C_WBIT | C_UBIT},
	{ALDREX, C_SOREG, C_NONE, C_REG, 77, 4, 0, 0, 0, 0},
	{ASTREX, C_SOREG, C_REG, C_REG, 78, 4, 0, 0, 0, 0},
	{ADMB, C_NONE, C_NONE, C_NONE, 110, 4, 0, 0, 0, 0},
	{ADMB, C_LCON, C_NONE, C_NONE, 110, 4, 0, 0, 0, 0},
	{ADMB, C_SPR, C_NONE, C_NONE, 110, 4, 0, 0, 0, 0},
	{AMOVF, C_ZFCON, C_NONE, C_FREG, 80, 8, 0, 0, 0, 0},
	{AMOVF, C_SFCON, C_NONE, C_FREG, 81, 4, 0, 0, 0, 0},
	{ACMPF, C_FREG, C_FREG, C_NONE, 82, 8, 0, 0, 0, 0},
	{ACMPF, C_FREG, C_NONE, C_NONE, 83, 8, 0, 0, 0, 0},
	{AMOVFW, C_FREG, C_NONE, C_FREG, 84, 4, 0, 0, 0, C_UBIT},
	{AMOVWF, C_FREG, C_NONE, C_FREG, 85, 4, 0, 0, 0, C_UBIT},
	{AMOVFW, C_FREG, C_NONE, C_REG, 86, 8, 0, 0, 0, C_UBIT},
	{AMOVWF, C_REG, C_NONE, C_FREG, 87, 8, 0, 0, 0, C_UBIT},
	{AMOVW, C_REG, C_NONE, C_FREG, 88, 4, 0, 0, 0, 0},
	{AMOVW, C_FREG, C_NONE, C_REG, 89, 4, 0, 0, 0, 0},
	{ALDREXD, C_SOREG, C_NONE, C_REG, 91, 4, 0, 0, 0, 0},
	{ASTREXD, C_SOREG, C_REG, C_REG, 92, 4, 0, 0, 0, 0},
	{APLD, C_SOREG, C_NONE, C_NONE, 95, 4, 0, 0, 0, 0},
	{obj.AUNDEF, C_NONE, C_NONE, C_NONE, 96, 4, 0, 0, 0, 0},
	{ACLZ, C_REG, C_NONE, C_REG, 97, 4, 0, 0, 0, 0},
	{AMULWT, C_REG, C_REG, C_REG, 98, 4, 0, 0, 0, 0},
	{AMULA, C_REG, C_REG, C_REGREG2, 99, 4, 0, 0, 0, C_SBIT},
	{AMULAWT, C_REG, C_REG, C_REGREG2, 99, 4, 0, 0, 0, 0},
	{obj.APCDATA, C_LCON, C_NONE, C_LCON, 0, 0, 0, 0, 0, 0},
	{obj.AFUNCDATA, C_LCON, C_NONE, C_ADDR, 0, 0, 0, 0, 0, 0},
	{obj.ANOP, C_NONE, C_NONE, C_NONE, 0, 0, 0, 0, 0, 0},
	{obj.ANOP, C_LCON, C_NONE, C_NONE, 0, 0, 0, 0, 0, 0}, 
	{obj.ANOP, C_REG, C_NONE, C_NONE, 0, 0, 0, 0, 0, 0},
	{obj.ANOP, C_FREG, C_NONE, C_NONE, 0, 0, 0, 0, 0, 0},
	{obj.ADUFFZERO, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0, 0}, 
	{obj.ADUFFCOPY, C_NONE, C_NONE, C_SBRA, 5, 4, 0, 0, 0, 0}, 
	{obj.AXXX, C_NONE, C_NONE, C_NONE, 0, 4, 0, 0, 0, 0},
}

var mbOp = []struct {
	reg int16
	enc uint32
}{
	{REG_MB_SY, 15},
	{REG_MB_ST, 14},
	{REG_MB_ISH, 11},
	{REG_MB_ISHST, 10},
	{REG_MB_NSH, 7},
	{REG_MB_NSHST, 6},
	{REG_MB_OSH, 3},
	{REG_MB_OSHST, 2},
}

var oprange [ALAST & obj.AMask][]Optab

var xcmp [C_GOK + 1][C_GOK + 1]bool

var (
	symdiv  *obj.LSym
	symdivu *obj.LSym
	symmod  *obj.LSym
	symmodu *obj.LSym
)






func checkSuffix(c *ctxt5, p *obj.Prog, o *Optab) {
	if p.Scond&C_SBIT != 0 && o.scond&C_SBIT == 0 {
		c.ctxt.Diag("invalid .S suffix: %v", p)
	}
	if p.Scond&C_PBIT != 0 && o.scond&C_PBIT == 0 {
		c.ctxt.Diag("invalid .P suffix: %v", p)
	}
	if p.Scond&C_WBIT != 0 && o.scond&C_WBIT == 0 {
		c.ctxt.Diag("invalid .W suffix: %v", p)
	}
	if p.Scond&C_UBIT != 0 && o.scond&C_UBIT == 0 {
		c.ctxt.Diag("invalid .U suffix: %v", p)
	}
}

func span5(ctxt *obj.Link, cursym *obj.LSym, newprog obj.ProgAlloc) {
	if ctxt.Retpoline {
		ctxt.Diag("-spectre=ret not supported on arm")
		ctxt.Retpoline = false 
	}

	var p *obj.Prog
	var op *obj.Prog

	p = cursym.Func().Text
	if p == nil || p.Link == nil { 
		return
	}

	if oprange[AAND&obj.AMask] == nil {
		ctxt.Diag("arm ops not initialized, call arm.buildop first")
	}

	c := ctxt5{ctxt: ctxt, newprog: newprog, cursym: cursym, autosize: p.To.Offset + 4}
	pc := int32(0)

	op = p
	p = p.Link
	var m int
	var o *Optab
	for ; p != nil || c.blitrl != nil; op, p = p, p.Link {
		if p == nil {
			if c.checkpool(op, pc) {
				p = op
				continue
			}

			
			ctxt.Diag("internal inconsistency")

			break
		}

		p.Pc = int64(pc)
		o = c.oplook(p)
		m = int(o.size)

		if m%4 != 0 || p.Pc%4 != 0 {
			ctxt.Diag("!pc invalid: %v size=%d", p, m)
		}

		
		if c.blitrl != nil {
			
			
			if c.checkpool(op, pc+int32(m)) {
				
				
				
				p = op
				continue
			}
		}

		if m == 0 && (p.As != obj.AFUNCDATA && p.As != obj.APCDATA && p.As != obj.ANOP) {
			ctxt.Diag("zero-width instruction\n%v", p)
			continue
		}

		switch o.flag & (LFROM | LTO | LPOOL) {
		case LFROM:
			c.addpool(p, &p.From)

		case LTO:
			c.addpool(p, &p.To)

		case LPOOL:
			if p.Scond&C_SCOND == C_SCOND_NONE {
				c.flushpool(p, 0, 0)
			}
		}

		if p.As == AMOVW && p.To.Type == obj.TYPE_REG && p.To.Reg == REGPC && p.Scond&C_SCOND == C_SCOND_NONE {
			c.flushpool(p, 0, 0)
		}

		pc += int32(m)
	}

	c.cursym.Size = int64(pc)

	
	times := 0

	var bflag int
	var opc int32
	var out [6 + 3]uint32
	for {
		bflag = 0
		pc = 0
		times++
		c.cursym.Func().Text.Pc = 0 
		for p = c.cursym.Func().Text; p != nil; p = p.Link {
			o = c.oplook(p)
			if int64(pc) > p.Pc {
				p.Pc = int64(pc)
			}

			
			opc = int32(p.Pc)
			m = int(o.size)
			if p.Pc != int64(opc) {
				bflag = 1
			}

			
			pc = int32(p.Pc + int64(m))

			if m%4 != 0 || p.Pc%4 != 0 {
				ctxt.Diag("pc invalid: %v size=%d", p, m)
			}

			if m/4 > len(out) {
				ctxt.Diag("instruction size too large: %d > %d", m/4, len(out))
			}
			if m == 0 && (p.As != obj.AFUNCDATA && p.As != obj.APCDATA && p.As != obj.ANOP) {
				if p.As == obj.ATEXT {
					c.autosize = p.To.Offset + 4
					continue
				}

				ctxt.Diag("zero-width instruction\n%v", p)
				continue
			}
		}

		c.cursym.Size = int64(pc)
		if bflag == 0 {
			break
		}
	}

	if pc%4 != 0 {
		ctxt.Diag("sym->size=%d, invalid", pc)
	}

	

	p = c.cursym.Func().Text
	c.autosize = p.To.Offset + 4
	c.cursym.Grow(c.cursym.Size)

	bp := c.cursym.P
	pc = int32(p.Pc) 
	var v int
	for p = p.Link; p != nil; p = p.Link {
		c.pc = p.Pc
		o = c.oplook(p)
		opc = int32(p.Pc)
		c.asmout(p, o, out[:])
		m = int(o.size)

		if m%4 != 0 || p.Pc%4 != 0 {
			ctxt.Diag("final stage: pc invalid: %v size=%d", p, m)
		}

		if int64(pc) > p.Pc {
			ctxt.Diag("PC padding invalid: want %#d, has %#d: %v", p.Pc, pc, p)
		}
		for int64(pc) != p.Pc {
			
			bp[0] = 0x00
			bp = bp[1:]

			bp[0] = 0x00
			bp = bp[1:]
			bp[0] = 0xa0
			bp = bp[1:]
			bp[0] = 0xe1
			bp = bp[1:]
			pc += 4
		}

		for i := 0; i < m/4; i++ {
			v = int(out[i])
			bp[0] = byte(v)
			bp = bp[1:]
			bp[0] = byte(v >> 8)
			bp = bp[1:]
			bp[0] = byte(v >> 16)
			bp = bp[1:]
			bp[0] = byte(v >> 24)
			bp = bp[1:]
		}

		pc += int32(m)
	}
}










func (c *ctxt5) checkpool(p *obj.Prog, nextpc int32) bool {
	poolLast := nextpc
	poolLast += 4                      
	poolLast += int32(c.pool.size) - 4 

	refPC := int32(c.pool.start) 

	v := poolLast - refPC - 8 

	if c.pool.size >= 0xff0 || immaddr(v) == 0 {
		return c.flushpool(p, 1, 0)
	} else if p.Link == nil {
		return c.flushpool(p, 2, 0)
	}
	return false
}

func (c *ctxt5) flushpool(p *obj.Prog, skip int, force int) bool {
	if c.blitrl != nil {
		if skip != 0 {
			if false && skip == 1 {
				fmt.Printf("note: flush literal pool at %x: len=%d ref=%x\n", uint64(p.Pc+4), c.pool.size, c.pool.start)
			}
			q := c.newprog()
			q.As = AB
			q.To.Type = obj.TYPE_BRANCH
			q.To.SetTarget(p.Link)
			q.Link = c.blitrl
			q.Pos = p.Pos
			c.blitrl = q
		} else if force == 0 && (p.Pc+int64(c.pool.size)-int64(c.pool.start) < 2048) {
			return false
		}

		
		
		
		for q := c.blitrl; q != nil; q = q.Link {
			q.Pos = p.Pos
		}

		c.elitrl.Link = p.Link
		p.Link = c.blitrl

		c.blitrl = nil 
		c.elitrl = nil
		c.pool.size = 0
		c.pool.start = 0
		c.pool.extra = 0
		return true
	}

	return false
}

func (c *ctxt5) addpool(p *obj.Prog, a *obj.Addr) {
	t := c.newprog()
	t.As = AWORD

	switch c.aclass(a) {
	default:
		t.To.Offset = a.Offset
		t.To.Sym = a.Sym
		t.To.Type = a.Type
		t.To.Name = a.Name

		if c.ctxt.Flag_shared && t.To.Sym != nil {
			t.Rel = p
		}

	case C_SROREG,
		C_LOREG,
		C_ROREG,
		C_FOREG,
		C_SOREG,
		C_HOREG,
		C_FAUTO,
		C_SAUTO,
		C_LAUTO,
		C_LACON:
		t.To.Type = obj.TYPE_CONST
		t.To.Offset = c.instoffset
	}

	if t.Rel == nil {
		for q := c.blitrl; q != nil; q = q.Link { 
			if q.Rel == nil && q.To == t.To {
				p.Pool = q
				return
			}
		}
	}

	q := c.newprog()
	*q = *t
	q.Pc = int64(c.pool.size)

	if c.blitrl == nil {
		c.blitrl = q
		c.pool.start = uint32(p.Pc)
	} else {
		c.elitrl.Link = q
	}
	c.elitrl = q
	c.pool.size += 4

	
	p.Pool = q
}

func (c *ctxt5) regoff(a *obj.Addr) int32 {
	c.instoffset = 0
	c.aclass(a)
	return int32(c.instoffset)
}

func immrot(v uint32) int32 {
	for i := 0; i < 16; i++ {
		if v&^0xff == 0 {
			return int32(uint32(int32(i)<<8) | v | 1<<25)
		}
		v = v<<2 | v>>30
	}

	return 0
}




func immrot2a(v uint32) (uint32, uint32) {
	for i := uint(1); i < 32; i++ {
		m := uint32(1<<i - 1)
		if x, y := immrot(v&m), immrot(v&^m); x != 0 && y != 0 {
			return uint32(x), uint32(y)
		}
	}
	
	
	return 0, 0
}




func immrot2s(v uint32) (uint32, uint32) {
	if immrot(v) != 0 {
		return v, 0
	}
	
	
	var i uint32
	for i = 2; i < 32; i += 2 {
		if v&(1<<i-1) != 0 {
			break
		}
	}
	
	i += 6
	
	x := 1<<i - v&(1<<i-1)
	y := v + x
	if y, x = uint32(immrot(y)), uint32(immrot(x)); y != 0 && x != 0 {
		return y, x
	}
	return 0, 0
}

func immaddr(v int32) int32 {
	if v >= 0 && v <= 0xfff {
		return v&0xfff | 1<<24 | 1<<23  
	}
	if v >= -0xfff && v < 0 {
		return -v&0xfff | 1<<24 
	}
	return 0
}

func immfloat(v int32) bool {
	return v&0xC03 == 0 
}

func immhalf(v int32) bool {
	if v >= 0 && v <= 0xff {
		return v|1<<24|1<<23 != 0  
	}
	if v >= -0xff && v < 0 {
		return -v&0xff|1<<24 != 0 
	}
	return false
}

func (c *ctxt5) aclass(a *obj.Addr) int {
	switch a.Type {
	case obj.TYPE_NONE:
		return C_NONE

	case obj.TYPE_REG:
		c.instoffset = 0
		if REG_R0 <= a.Reg && a.Reg <= REG_R15 {
			return C_REG
		}
		if REG_F0 <= a.Reg && a.Reg <= REG_F15 {
			return C_FREG
		}
		if a.Reg == REG_FPSR || a.Reg == REG_FPCR {
			return C_FCR
		}
		if a.Reg == REG_CPSR || a.Reg == REG_SPSR {
			return C_PSR
		}
		if a.Reg >= REG_SPECIAL {
			return C_SPR
		}
		return C_GOK

	case obj.TYPE_REGREG:
		return C_REGREG

	case obj.TYPE_REGREG2:
		return C_REGREG2

	case obj.TYPE_REGLIST:
		return C_REGLIST

	case obj.TYPE_SHIFT:
		if a.Reg == 0 {
			
			return C_SHIFT
		} else {
			
			return C_SHIFTADDR
		}

	case obj.TYPE_MEM:
		switch a.Name {
		case obj.NAME_EXTERN,
			obj.NAME_GOTREF,
			obj.NAME_STATIC:
			if a.Sym == nil || a.Sym.Name == "" {
				fmt.Printf("null sym external\n")
				return C_GOK
			}

			c.instoffset = 0 
			if a.Sym.Type == objabi.STLSBSS {
				if c.ctxt.Flag_shared {
					return C_TLS_IE
				} else {
					return C_TLS_LE
				}
			}

			return C_ADDR

		case obj.NAME_AUTO:
			if a.Reg == REGSP {
				
				
				a.Reg = obj.REG_NONE
			}
			c.instoffset = c.autosize + a.Offset
			if t := immaddr(int32(c.instoffset)); t != 0 {
				if immhalf(int32(c.instoffset)) {
					if immfloat(t) {
						return C_HFAUTO
					}
					return C_HAUTO
				}

				if immfloat(t) {
					return C_FAUTO
				}
				return C_SAUTO
			}

			return C_LAUTO

		case obj.NAME_PARAM:
			if a.Reg == REGSP {
				
				
				a.Reg = obj.REG_NONE
			}
			c.instoffset = c.autosize + a.Offset + 4
			if t := immaddr(int32(c.instoffset)); t != 0 {
				if immhalf(int32(c.instoffset)) {
					if immfloat(t) {
						return C_HFAUTO
					}
					return C_HAUTO
				}

				if immfloat(t) {
					return C_FAUTO
				}
				return C_SAUTO
			}

			return C_LAUTO

		case obj.NAME_NONE:
			c.instoffset = a.Offset
			if t := immaddr(int32(c.instoffset)); t != 0 {
				if immhalf(int32(c.instoffset)) { 
					if immfloat(t) {
						return C_HFOREG
					}
					return C_HOREG
				}

				if immfloat(t) {
					return C_FOREG 
				}
				if immrot(uint32(c.instoffset)) != 0 {
					return C_SROREG
				}
				if immhalf(int32(c.instoffset)) {
					return C_HOREG
				}
				return C_SOREG
			}

			if immrot(uint32(c.instoffset)) != 0 {
				return C_ROREG
			}
			return C_LOREG
		}

		return C_GOK

	case obj.TYPE_FCONST:
		if c.chipzero5(a.Val.(float64)) >= 0 {
			return C_ZFCON
		}
		if c.chipfloat5(a.Val.(float64)) >= 0 {
			return C_SFCON
		}
		return C_LFCON

	case obj.TYPE_TEXTSIZE:
		return C_TEXTSIZE

	case obj.TYPE_CONST,
		obj.TYPE_ADDR:
		switch a.Name {
		case obj.NAME_NONE:
			c.instoffset = a.Offset
			if a.Reg != 0 {
				return c.aconsize()
			}

			if immrot(uint32(c.instoffset)) != 0 {
				return C_RCON
			}
			if immrot(^uint32(c.instoffset)) != 0 {
				return C_NCON
			}
			if uint32(c.instoffset) <= 0xffff && buildcfg.GOARM == 7 {
				return C_SCON
			}
			if x, y := immrot2a(uint32(c.instoffset)); x != 0 && y != 0 {
				return C_RCON2A
			}
			if y, x := immrot2s(uint32(c.instoffset)); x != 0 && y != 0 {
				return C_RCON2S
			}
			return C_LCON

		case obj.NAME_EXTERN,
			obj.NAME_GOTREF,
			obj.NAME_STATIC:
			s := a.Sym
			if s == nil {
				break
			}
			c.instoffset = 0 
			return C_LCONADDR

		case obj.NAME_AUTO:
			if a.Reg == REGSP {
				
				
				a.Reg = obj.REG_NONE
			}
			c.instoffset = c.autosize + a.Offset
			return c.aconsize()

		case obj.NAME_PARAM:
			if a.Reg == REGSP {
				
				
				a.Reg = obj.REG_NONE
			}
			c.instoffset = c.autosize + a.Offset + 4
			return c.aconsize()
		}

		return C_GOK

	case obj.TYPE_BRANCH:
		return C_SBRA
	}

	return C_GOK
}

func (c *ctxt5) aconsize() int {
	if immrot(uint32(c.instoffset)) != 0 {
		return C_RACON
	}
	if immrot(uint32(-c.instoffset)) != 0 {
		return C_RACON
	}
	return C_LACON
}

func (c *ctxt5) oplook(p *obj.Prog) *Optab {
	a1 := int(p.Optab)
	if a1 != 0 {
		return &optab[a1-1]
	}
	a1 = int(p.From.Class)
	if a1 == 0 {
		a1 = c.aclass(&p.From) + 1
		p.From.Class = int8(a1)
	}

	a1--
	a3 := int(p.To.Class)
	if a3 == 0 {
		a3 = c.aclass(&p.To) + 1
		p.To.Class = int8(a3)
	}

	a3--
	a2 := C_NONE
	if p.Reg != 0 {
		switch {
		case REG_F0 <= p.Reg && p.Reg <= REG_F15:
			a2 = C_FREG
		case REG_R0 <= p.Reg && p.Reg <= REG_R15:
			a2 = C_REG
		default:
			c.ctxt.Diag("invalid register in %v", p)
		}
	}

	
	switch a1 {
	case C_SOREG, C_LOREG, C_HOREG, C_FOREG, C_ROREG, C_HFOREG, C_SROREG, C_SHIFTADDR:
		if p.From.Reg < REG_R0 || REG_R15 < p.From.Reg {
			c.ctxt.Diag("illegal base register: %v", p)
		}
	default:
	}
	switch a3 {
	case C_SOREG, C_LOREG, C_HOREG, C_FOREG, C_ROREG, C_HFOREG, C_SROREG, C_SHIFTADDR:
		if p.To.Reg < REG_R0 || REG_R15 < p.To.Reg {
			c.ctxt.Diag("illegal base register: %v", p)
		}
	default:
	}

	
	
	if (a1 == C_RCON2A || a1 == C_RCON2S) && p.Scond&C_SBIT != 0 {
		a1 = C_LCON
	}
	if (a3 == C_RCON2A || a3 == C_RCON2S) && p.Scond&C_SBIT != 0 {
		a3 = C_LCON
	}

	if false { 
		fmt.Printf("oplook %v %v %v %v\n", p.As, DRconv(a1), DRconv(a2), DRconv(a3))
		fmt.Printf("\t\t%d %d\n", p.From.Type, p.To.Type)
	}

	ops := oprange[p.As&obj.AMask]
	c1 := &xcmp[a1]
	c3 := &xcmp[a3]
	for i := range ops {
		op := &ops[i]
		if int(op.a2) == a2 && c1[op.a1] && c3[op.a3] {
			p.Optab = uint16(cap(optab) - cap(ops) + i + 1)
			checkSuffix(c, p, op)
			return op
		}
	}

	c.ctxt.Diag("illegal combination %v; %v %v %v; from %d %d; to %d %d", p, DRconv(a1), DRconv(a2), DRconv(a3), p.From.Type, p.From.Name, p.To.Type, p.To.Name)
	if ops == nil {
		ops = optab
	}
	return &ops[0]
}

func cmp(a int, b int) bool {
	if a == b {
		return true
	}
	switch a {
	case C_LCON:
		if b == C_RCON || b == C_NCON || b == C_SCON || b == C_RCON2A || b == C_RCON2S {
			return true
		}

	case C_LACON:
		if b == C_RACON {
			return true
		}

	case C_LFCON:
		if b == C_ZFCON || b == C_SFCON {
			return true
		}

	case C_HFAUTO:
		return b == C_HAUTO || b == C_FAUTO

	case C_FAUTO, C_HAUTO:
		return b == C_HFAUTO

	case C_SAUTO:
		return cmp(C_HFAUTO, b)

	case C_LAUTO:
		return cmp(C_SAUTO, b)

	case C_HFOREG:
		return b == C_HOREG || b == C_FOREG

	case C_FOREG, C_HOREG:
		return b == C_HFOREG

	case C_SROREG:
		return cmp(C_SOREG, b) || cmp(C_ROREG, b)

	case C_SOREG, C_ROREG:
		return b == C_SROREG || cmp(C_HFOREG, b)

	case C_LOREG:
		return cmp(C_SROREG, b)

	case C_LBRA:
		if b == C_SBRA {
			return true
		}

	case C_HREG:
		return cmp(C_SP, b) || cmp(C_PC, b)
	}

	return false
}

type ocmp []Optab

func (x ocmp) Len() int {
	return len(x)
}

func (x ocmp) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x ocmp) Less(i, j int) bool {
	p1 := &x[i]
	p2 := &x[j]
	n := int(p1.as) - int(p2.as)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a1) - int(p2.a1)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a2) - int(p2.a2)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a3) - int(p2.a3)
	if n != 0 {
		return n < 0
	}
	return false
}

func opset(a, b0 obj.As) {
	oprange[a&obj.AMask] = oprange[b0]
}

func buildop(ctxt *obj.Link) {
	if oprange[AAND&obj.AMask] != nil {
		
		
		
		return
	}

	symdiv = ctxt.Lookup("runtime._div")
	symdivu = ctxt.Lookup("runtime._divu")
	symmod = ctxt.Lookup("runtime._mod")
	symmodu = ctxt.Lookup("runtime._modu")

	var n int

	for i := 0; i < C_GOK; i++ {
		for n = 0; n < C_GOK; n++ {
			if cmp(n, i) {
				xcmp[i][n] = true
			}
		}
	}
	for n = 0; optab[n].as != obj.AXXX; n++ {
		if optab[n].flag&LPCREL != 0 {
			if ctxt.Flag_shared {
				optab[n].size += int8(optab[n].pcrelsiz)
			} else {
				optab[n].flag &^= LPCREL
			}
		}
	}

	sort.Sort(ocmp(optab[:n]))
	for i := 0; i < n; i++ {
		r := optab[i].as
		r0 := r & obj.AMask
		start := i
		for optab[i].as == r {
			i++
		}
		oprange[r0] = optab[start:i]
		i--

		switch r {
		default:
			ctxt.Diag("unknown op in build: %v", r)
			ctxt.DiagFlush()
			log.Fatalf("bad code")

		case AADD:
			opset(ASUB, r0)
			opset(ARSB, r0)
			opset(AADC, r0)
			opset(ASBC, r0)
			opset(ARSC, r0)

		case AORR:
			opset(AEOR, r0)
			opset(ABIC, r0)

		case ACMP:
			opset(ATEQ, r0)
			opset(ACMN, r0)
			opset(ATST, r0)

		case AMVN:
			break

		case ABEQ:
			opset(ABNE, r0)
			opset(ABCS, r0)
			opset(ABHS, r0)
			opset(ABCC, r0)
			opset(ABLO, r0)
			opset(ABMI, r0)
			opset(ABPL, r0)
			opset(ABVS, r0)
			opset(ABVC, r0)
			opset(ABHI, r0)
			opset(ABLS, r0)
			opset(ABGE, r0)
			opset(ABLT, r0)
			opset(ABGT, r0)
			opset(ABLE, r0)

		case ASLL:
			opset(ASRL, r0)
			opset(ASRA, r0)

		case AMUL:
			opset(AMULU, r0)

		case ADIV:
			opset(AMOD, r0)
			opset(AMODU, r0)
			opset(ADIVU, r0)

		case ADIVHW:
			opset(ADIVUHW, r0)

		case AMOVW,
			AMOVB,
			AMOVBS,
			AMOVBU,
			AMOVH,
			AMOVHS,
			AMOVHU:
			break

		case ASWPW:
			opset(ASWPBU, r0)

		case AB,
			ABL,
			ABX,
			ABXRET,
			obj.ADUFFZERO,
			obj.ADUFFCOPY,
			ASWI,
			AWORD,
			AMOVM,
			ARFE,
			obj.ATEXT:
			break

		case AADDF:
			opset(AADDD, r0)
			opset(ASUBF, r0)
			opset(ASUBD, r0)
			opset(AMULF, r0)
			opset(AMULD, r0)
			opset(ANMULF, r0)
			opset(ANMULD, r0)
			opset(AMULAF, r0)
			opset(AMULAD, r0)
			opset(AMULSF, r0)
			opset(AMULSD, r0)
			opset(ANMULAF, r0)
			opset(ANMULAD, r0)
			opset(ANMULSF, r0)
			opset(ANMULSD, r0)
			opset(AFMULAF, r0)
			opset(AFMULAD, r0)
			opset(AFMULSF, r0)
			opset(AFMULSD, r0)
			opset(AFNMULAF, r0)
			opset(AFNMULAD, r0)
			opset(AFNMULSF, r0)
			opset(AFNMULSD, r0)
			opset(ADIVF, r0)
			opset(ADIVD, r0)

		case ANEGF:
			opset(ANEGD, r0)
			opset(ASQRTF, r0)
			opset(ASQRTD, r0)
			opset(AMOVFD, r0)
			opset(AMOVDF, r0)
			opset(AABSF, r0)
			opset(AABSD, r0)

		case ACMPF:
			opset(ACMPD, r0)

		case AMOVF:
			opset(AMOVD, r0)

		case AMOVFW:
			opset(AMOVDW, r0)

		case AMOVWF:
			opset(AMOVWD, r0)

		case AMULL:
			opset(AMULAL, r0)
			opset(AMULLU, r0)
			opset(AMULALU, r0)

		case AMULWT:
			opset(AMULWB, r0)
			opset(AMULBB, r0)
			opset(AMMUL, r0)

		case AMULAWT:
			opset(AMULAWB, r0)
			opset(AMULABB, r0)
			opset(AMULS, r0)
			opset(AMMULA, r0)
			opset(AMMULS, r0)

		case ABFX:
			opset(ABFXU, r0)
			opset(ABFC, r0)
			opset(ABFI, r0)

		case ACLZ:
			opset(AREV, r0)
			opset(AREV16, r0)
			opset(AREVSH, r0)
			opset(ARBIT, r0)

		case AXTAB:
			opset(AXTAH, r0)
			opset(AXTABU, r0)
			opset(AXTAHU, r0)

		case ALDREX,
			ASTREX,
			ALDREXD,
			ASTREXD,
			ADMB,
			APLD,
			AAND,
			AMULA,
			obj.AUNDEF,
			obj.AFUNCDATA,
			obj.APCDATA,
			obj.ANOP:
			break
		}
	}
}

func (c *ctxt5) asmout(p *obj.Prog, o *Optab, out []uint32) {
	c.printp = p
	o1 := uint32(0)
	o2 := uint32(0)
	o3 := uint32(0)
	o4 := uint32(0)
	o5 := uint32(0)
	o6 := uint32(0)
	if false { 
		fmt.Printf("%x: %v\ttype %d\n", uint32(p.Pc), p, o.type_)
	}
	switch o.type_ {
	default:
		c.ctxt.Diag("%v: unknown asm %d", p, o.type_)

	case 0: 
		if false { 
			fmt.Printf("%x: %s: arm\n", uint32(p.Pc), p.From.Sym.Name)
		}

	case 1: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		rf := int(p.From.Reg)
		rt := int(p.To.Reg)
		r := int(p.Reg)
		if p.To.Type == obj.TYPE_NONE {
			rt = 0
		}
		if p.As == AMOVB || p.As == AMOVH || p.As == AMOVW || p.As == AMVN {
			r = 0
		} else if r == 0 {
			r = rt
		}
		o1 |= (uint32(rf)&15)<<0 | (uint32(r)&15)<<16 | (uint32(rt)&15)<<12

	case 2: 
		c.aclass(&p.From)

		o1 = c.oprrr(p, p.As, int(p.Scond))
		o1 |= uint32(immrot(uint32(c.instoffset)))
		rt := int(p.To.Reg)
		r := int(p.Reg)
		if p.To.Type == obj.TYPE_NONE {
			rt = 0
		}
		if p.As == AMOVW || p.As == AMVN {
			r = 0
		} else if r == 0 {
			r = rt
		}
		o1 |= (uint32(r)&15)<<16 | (uint32(rt)&15)<<12

	case 106: 
		c.aclass(&p.From)
		r := int(p.Reg)
		rt := int(p.To.Reg)
		if r == 0 {
			r = rt
		}
		x, y := immrot2a(uint32(c.instoffset))
		var as2 obj.As
		switch p.As {
		case AADD, ASUB, AORR, AEOR, ABIC:
			as2 = p.As 
		case ARSB:
			as2 = AADD 
		case AADC:
			as2 = AADD 
		case ASBC:
			as2 = ASUB 
		case ARSC:
			as2 = AADD 
		default:
			c.ctxt.Diag("unknown second op for %v", p)
		}
		o1 = c.oprrr(p, p.As, int(p.Scond))
		o2 = c.oprrr(p, as2, int(p.Scond))
		o1 |= (uint32(r)&15)<<16 | (uint32(rt)&15)<<12
		o2 |= (uint32(rt)&15)<<16 | (uint32(rt)&15)<<12
		o1 |= x
		o2 |= y

	case 107: 
		c.aclass(&p.From)
		r := int(p.Reg)
		rt := int(p.To.Reg)
		if r == 0 {
			r = rt
		}
		y, x := immrot2s(uint32(c.instoffset))
		var as2 obj.As
		switch p.As {
		case AADD:
			as2 = ASUB 
		case ASUB:
			as2 = AADD 
		case ARSB:
			as2 = ASUB 
		case AADC:
			as2 = ASUB 
		case ASBC:
			as2 = AADD 
		case ARSC:
			as2 = ASUB 
		default:
			c.ctxt.Diag("unknown second op for %v", p)
		}
		o1 = c.oprrr(p, p.As, int(p.Scond))
		o2 = c.oprrr(p, as2, int(p.Scond))
		o1 |= (uint32(r)&15)<<16 | (uint32(rt)&15)<<12
		o2 |= (uint32(rt)&15)<<16 | (uint32(rt)&15)<<12
		o1 |= y
		o2 |= x

	case 3: 
		o1 = c.mov(p)

	case 4: 
		c.aclass(&p.From)
		if c.instoffset < 0 {
			o1 = c.oprrr(p, ASUB, int(p.Scond))
			o1 |= uint32(immrot(uint32(-c.instoffset)))
		} else {
			o1 = c.oprrr(p, AADD, int(p.Scond))
			o1 |= uint32(immrot(uint32(c.instoffset)))
		}
		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 |= (uint32(r) & 15) << 16
		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 5: 
		o1 = c.opbra(p, p.As, int(p.Scond))

		v := int32(-8)
		if p.To.Sym != nil {
			rel := obj.Addrel(c.cursym)
			rel.Off = int32(c.pc)
			rel.Siz = 4
			rel.Sym = p.To.Sym
			v += int32(p.To.Offset)
			rel.Add = int64(o1) | (int64(v)>>2)&0xffffff
			rel.Type = objabi.R_CALLARM
			break
		}

		if p.To.Target() != nil {
			v = int32((p.To.Target().Pc - c.pc) - 8)
		}
		o1 |= (uint32(v) >> 2) & 0xffffff

	case 6: 
		c.aclass(&p.To)

		o1 = c.oprrr(p, AADD, int(p.Scond))
		o1 |= uint32(immrot(uint32(c.instoffset)))
		o1 |= (uint32(p.To.Reg) & 15) << 16
		o1 |= (REGPC & 15) << 12

	case 7: 
		c.aclass(&p.To)

		if c.instoffset != 0 {
			c.ctxt.Diag("%v: doesn't support BL offset(REG) with non-zero offset %d", p, c.instoffset)
		}
		o1 = c.oprrr(p, ABL, int(p.Scond))
		o1 |= (uint32(p.To.Reg) & 15) << 0
		rel := obj.Addrel(c.cursym)
		rel.Off = int32(c.pc)
		rel.Siz = 0
		rel.Type = objabi.R_CALLIND

	case 8: 
		c.aclass(&p.From)

		o1 = c.oprrr(p, p.As, int(p.Scond))
		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}
		o1 |= (uint32(r) & 15) << 0
		o1 |= uint32((c.instoffset & 31) << 7)
		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 9: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}
		o1 |= (uint32(r) & 15) << 0
		o1 |= (uint32(p.From.Reg)&15)<<8 | 1<<4
		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 10: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		if p.To.Type != obj.TYPE_NONE {
			c.aclass(&p.To)
			o1 |= uint32(c.instoffset & 0xffffff)
		}

	case 11: 
		c.aclass(&p.To)

		o1 = uint32(c.instoffset)
		if p.To.Sym != nil {
			
			
			rel := obj.Addrel(c.cursym)

			rel.Off = int32(c.pc)
			rel.Siz = 4
			rel.Sym = p.To.Sym
			rel.Add = p.To.Offset

			if c.ctxt.Flag_shared {
				if p.To.Name == obj.NAME_GOTREF {
					rel.Type = objabi.R_GOTPCREL
				} else {
					rel.Type = objabi.R_PCREL
				}
				rel.Add += c.pc - p.Rel.Pc - 8
			} else {
				rel.Type = objabi.R_ADDR
			}
			o1 = 0
		}

	case 12: 
		if o.a1 == C_SCON {
			o1 = c.omvs(p, &p.From, int(p.To.Reg))
		} else if p.As == AMVN {
			o1 = c.omvr(p, &p.From, int(p.To.Reg))
		} else {
			o1 = c.omvl(p, &p.From, int(p.To.Reg))
		}

		if o.flag&LPCREL != 0 {
			o2 = c.oprrr(p, AADD, int(p.Scond)) | (uint32(p.To.Reg)&15)<<0 | (REGPC&15)<<16 | (uint32(p.To.Reg)&15)<<12
		}

	case 13: 
		if o.a1 == C_SCON {
			o1 = c.omvs(p, &p.From, REGTMP)
		} else {
			o1 = c.omvl(p, &p.From, REGTMP)
		}

		if o1 == 0 {
			break
		}
		o2 = c.oprrr(p, p.As, int(p.Scond))
		o2 |= REGTMP & 15
		r := int(p.Reg)
		if p.As == AMVN {
			r = 0
		} else if r == 0 {
			r = int(p.To.Reg)
		}
		o2 |= (uint32(r) & 15) << 16
		if p.To.Type != obj.TYPE_NONE {
			o2 |= (uint32(p.To.Reg) & 15) << 12
		}

	case 14: 
		o1 = c.oprrr(p, ASLL, int(p.Scond))

		if p.As == AMOVBU || p.As == AMOVHU {
			o2 = c.oprrr(p, ASRL, int(p.Scond))
		} else {
			o2 = c.oprrr(p, ASRA, int(p.Scond))
		}

		r := int(p.To.Reg)
		o1 |= (uint32(p.From.Reg)&15)<<0 | (uint32(r)&15)<<12
		o2 |= uint32(r)&15 | (uint32(r)&15)<<12
		if p.As == AMOVB || p.As == AMOVBS || p.As == AMOVBU {
			o1 |= 24 << 7
			o2 |= 24 << 7
		} else {
			o1 |= 16 << 7
			o2 |= 16 << 7
		}

	case 15: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		rf := int(p.From.Reg)
		rt := int(p.To.Reg)
		r := int(p.Reg)
		if r == 0 {
			r = rt
		}

		o1 |= (uint32(rf)&15)<<8 | (uint32(r)&15)<<0 | (uint32(rt)&15)<<16

	case 16: 
		o1 = 0xf << 28

		o2 = 0

	case 17:
		o1 = c.oprrr(p, p.As, int(p.Scond))
		rf := int(p.From.Reg)
		rt := int(p.To.Reg)
		rt2 := int(p.To.Offset)
		r := int(p.Reg)
		o1 |= (uint32(rf)&15)<<8 | (uint32(r)&15)<<0 | (uint32(rt)&15)<<16 | (uint32(rt2)&15)<<12

	case 18: 
		o1 = c.oprrr(p, p.As, int(p.Scond))
		rt := int(p.To.Reg)
		r := int(p.Reg)
		if r == 0 {
			r = rt
		} else if p.As == ABFC { 
			c.ctxt.Diag("illegal combination: %v", p)
		}
		if p.GetFrom3() == nil || p.GetFrom3().Type != obj.TYPE_CONST {
			c.ctxt.Diag("%v: missing or wrong LSB", p)
			break
		}
		lsb := p.GetFrom3().Offset
		width := p.From.Offset
		if lsb < 0 || lsb > 31 || width <= 0 || (lsb+width) > 32 {
			c.ctxt.Diag("%v: wrong width or LSB", p)
		}
		switch p.As {
		case ABFX, ABFXU: 
			o1 |= (uint32(r)&15)<<0 | (uint32(rt)&15)<<12 | uint32(lsb)<<7 | uint32(width-1)<<16
		case ABFC, ABFI: 
			o1 |= (uint32(r)&15)<<0 | (uint32(rt)&15)<<12 | uint32(lsb)<<7 | uint32(lsb+width-1)<<16
		default:
			c.ctxt.Diag("illegal combination: %v", p)
		}

	case 20: 
		c.aclass(&p.To)

		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 = c.osr(p.As, int(p.From.Reg), int32(c.instoffset), r, int(p.Scond))

	case 21: 
		c.aclass(&p.From)

		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 = c.olr(int32(c.instoffset), r, int(p.To.Reg), int(p.Scond))
		if p.As != AMOVW {
			o1 |= 1 << 22
		}

	case 22: 
		o1 = c.oprrr(p, p.As, int(p.Scond))
		switch p.From.Offset &^ 0xf {
		
		case SHIFT_RR, SHIFT_RR | 8<<7, SHIFT_RR | 16<<7, SHIFT_RR | 24<<7:
			o1 |= uint32(p.From.Offset) & 0xc0f
		default:
			c.ctxt.Diag("illegal shift: %v", p)
		}
		rt := p.To.Reg
		r := p.Reg
		if r == 0 {
			r = rt
		}
		o1 |= (uint32(rt)&15)<<12 | (uint32(r)&15)<<16

	case 23: 
		switch p.As {
		case AMOVW:
			o1 = c.mov(p)
		case AMOVBU, AMOVBS, AMOVB, AMOVHU, AMOVHS, AMOVH:
			o1 = c.movxt(p)
		default:
			c.ctxt.Diag("illegal combination: %v", p)
		}

	case 30: 
		o1 = c.omvl(p, &p.To, REGTMP)

		if o1 == 0 {
			break
		}
		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 = c.osrr(int(p.From.Reg), REGTMP&15, r, int(p.Scond))
		if p.As != AMOVW {
			o2 |= 1 << 22
		}

	case 31: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}
		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 = c.olrr(REGTMP&15, r, int(p.To.Reg), int(p.Scond))
		if p.As == AMOVBU || p.As == AMOVBS || p.As == AMOVB {
			o2 |= 1 << 22
		}

	case 34: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}

		o2 = c.oprrr(p, AADD, int(p.Scond))
		o2 |= REGTMP & 15
		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 |= (uint32(r) & 15) << 16
		if p.To.Type != obj.TYPE_NONE {
			o2 |= (uint32(p.To.Reg) & 15) << 12
		}

	case 35: 
		o1 = 2<<23 | 0xf<<16 | 0<<0

		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
		o1 |= (uint32(p.From.Reg) & 1) << 22
		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 36: 
		o1 = 2<<23 | 0x2cf<<12 | 0<<4

		if p.Scond&C_FBIT != 0 {
			o1 ^= 0x010 << 12
		}
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
		o1 |= (uint32(p.To.Reg) & 1) << 22
		o1 |= (uint32(p.From.Reg) & 15) << 0

	case 37: 
		c.aclass(&p.From)

		o1 = 2<<23 | 0x2cf<<12 | 0<<4
		if p.Scond&C_FBIT != 0 {
			o1 ^= 0x010 << 12
		}
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
		o1 |= uint32(immrot(uint32(c.instoffset)))
		o1 |= (uint32(p.To.Reg) & 1) << 22
		o1 |= (uint32(p.From.Reg) & 15) << 0

	case 38, 39:
		switch o.type_ {
		case 38: 
			o1 = 0x4 << 25

			o1 |= uint32(p.From.Offset & 0xffff)
			o1 |= (uint32(p.To.Reg) & 15) << 16
			c.aclass(&p.To)

		case 39: 
			o1 = 0x4<<25 | 1<<20

			o1 |= uint32(p.To.Offset & 0xffff)
			o1 |= (uint32(p.From.Reg) & 15) << 16
			c.aclass(&p.From)
		}

		if c.instoffset != 0 {
			c.ctxt.Diag("offset must be zero in MOVM; %v", p)
		}
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
		if p.Scond&C_PBIT != 0 {
			o1 |= 1 << 24
		}
		if p.Scond&C_UBIT != 0 {
			o1 |= 1 << 23
		}
		if p.Scond&C_WBIT != 0 {
			o1 |= 1 << 21
		}

	case 40: 
		c.aclass(&p.From)

		if c.instoffset != 0 {
			c.ctxt.Diag("offset must be zero in SWP")
		}
		o1 = 0x2<<23 | 0x9<<4
		if p.As != ASWPW {
			o1 |= 1 << 22
		}
		o1 |= (uint32(p.From.Reg) & 15) << 16
		o1 |= (uint32(p.Reg) & 15) << 0
		o1 |= (uint32(p.To.Reg) & 15) << 12
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 41: 
		o1 = 0xe8fd8000

	case 50: 
		v := c.regoff(&p.To)

		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 = c.ofsr(p.As, int(p.From.Reg), v, r, int(p.Scond), p)

	case 51: 
		v := c.regoff(&p.From)

		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 = c.ofsr(p.As, int(p.To.Reg), v, r, int(p.Scond), p) | 1<<20

	case 52: 
		o1 = c.omvl(p, &p.To, REGTMP)

		if o1 == 0 {
			break
		}
		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 = c.oprrr(p, AADD, int(p.Scond)) | (REGTMP&15)<<12 | (REGTMP&15)<<16 | (uint32(r)&15)<<0
		o3 = c.ofsr(p.As, int(p.From.Reg), 0, REGTMP, int(p.Scond), p)

	case 53: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}
		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 = c.oprrr(p, AADD, int(p.Scond)) | (REGTMP&15)<<12 | (REGTMP&15)<<16 | (uint32(r)&15)<<0
		o3 = c.ofsr(p.As, int(p.To.Reg), 0, (REGTMP&15), int(p.Scond), p) | 1<<20

	case 54: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		rf := int(p.From.Reg)
		rt := int(p.To.Reg)
		r := int(p.Reg)
		if r == 0 {
			switch p.As {
			case AMULAD, AMULAF, AMULSF, AMULSD, ANMULAF, ANMULAD, ANMULSF, ANMULSD,
				AFMULAD, AFMULAF, AFMULSF, AFMULSD, AFNMULAF, AFNMULAD, AFNMULSF, AFNMULSD:
				c.ctxt.Diag("illegal combination: %v", p)
			default:
				r = rt
			}
		}

		o1 |= (uint32(rf)&15)<<0 | (uint32(r)&15)<<16 | (uint32(rt)&15)<<12

	case 55: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		rf := int(p.From.Reg)
		rt := int(p.To.Reg)

		o1 |= (uint32(rf)&15)<<0 | (uint32(rt)&15)<<12

	case 56: 
		o1 = ((uint32(p.Scond)&C_SCOND)^C_SCOND_XOR)<<28 | 0xee1<<16 | 0xa1<<4

		o1 |= (uint32(p.From.Reg) & 15) << 12

	case 57: 
		o1 = ((uint32(p.Scond)&C_SCOND)^C_SCOND_XOR)<<28 | 0xef1<<16 | 0xa1<<4

		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 58: 
		o1 = c.oprrr(p, AAND, int(p.Scond))

		o1 |= uint32(immrot(0xff))
		rt := int(p.To.Reg)
		r := int(p.From.Reg)
		if p.To.Type == obj.TYPE_NONE {
			rt = 0
		}
		if r == 0 {
			r = rt
		}
		o1 |= (uint32(r)&15)<<16 | (uint32(rt)&15)<<12

	case 59: 
		if p.From.Reg == 0 {
			c.ctxt.Diag("source operand is not a memory address: %v", p)
			break
		}
		if p.From.Offset&(1<<4) != 0 {
			c.ctxt.Diag("bad shift in LDR")
			break
		}
		o1 = c.olrr(int(p.From.Offset), int(p.From.Reg), int(p.To.Reg), int(p.Scond))
		if p.As == AMOVBU {
			o1 |= 1 << 22
		}

	case 60: 
		if p.From.Reg == 0 {
			c.ctxt.Diag("source operand is not a memory address: %v", p)
			break
		}
		if p.From.Offset&(^0xf) != 0 {
			c.ctxt.Diag("bad shift: %v", p)
			break
		}
		o1 = c.olhrr(int(p.From.Offset), int(p.From.Reg), int(p.To.Reg), int(p.Scond))
		switch p.As {
		case AMOVB, AMOVBS:
			o1 ^= 1<<5 | 1<<6
		case AMOVH, AMOVHS:
			o1 ^= 1 << 6
		default:
		}
		if p.Scond&C_UBIT != 0 {
			o1 &^= 1 << 23
		}

	case 61: 
		if p.To.Reg == 0 {
			c.ctxt.Diag("MOV to shifter operand")
		}
		o1 = c.osrr(int(p.From.Reg), int(p.To.Offset), int(p.To.Reg), int(p.Scond))
		if p.As == AMOVB || p.As == AMOVBS || p.As == AMOVBU {
			o1 |= 1 << 22
		}

	case 62: 
		if p.To.Reg == 0 {
			c.ctxt.Diag("MOV to shifter operand")
		}
		if p.To.Offset&(^0xf) != 0 {
			c.ctxt.Diag("bad shift: %v", p)
		}
		o1 = c.olhrr(int(p.To.Offset), int(p.To.Reg), int(p.From.Reg), int(p.Scond))
		o1 ^= 1 << 20
		if p.Scond&C_UBIT != 0 {
			o1 &^= 1 << 23
		}

		
	case 64: 
		o1 = c.omvl(p, &p.To, REGTMP)

		if o1 == 0 {
			break
		}
		o2 = c.osr(p.As, int(p.From.Reg), 0, REGTMP, int(p.Scond))
		if o.flag&LPCREL != 0 {
			o3 = o2
			o2 = c.oprrr(p, AADD, int(p.Scond)) | REGTMP&15 | (REGPC&15)<<16 | (REGTMP&15)<<12
		}

	case 65: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}
		o2 = c.olr(0, REGTMP, int(p.To.Reg), int(p.Scond))
		if p.As == AMOVBU || p.As == AMOVBS || p.As == AMOVB {
			o2 |= 1 << 22
		}
		if o.flag&LPCREL != 0 {
			o3 = o2
			o2 = c.oprrr(p, AADD, int(p.Scond)) | REGTMP&15 | (REGPC&15)<<16 | (REGTMP&15)<<12
		}

	case 101: 
		o1 = c.omvl(p, &p.From, int(p.To.Reg))

	case 102: 
		o1 = c.omvl(p, &p.From, int(p.To.Reg))
		o2 = c.olrr(int(p.To.Reg)&15, (REGPC & 15), int(p.To.Reg), int(p.Scond))

	case 103: 
		if p.To.Sym == nil {
			c.ctxt.Diag("nil sym in tls %v", p)
		}
		if p.To.Offset != 0 {
			c.ctxt.Diag("offset against tls var in %v", p)
		}
		
		
		rel := obj.Addrel(c.cursym)

		rel.Off = int32(c.pc)
		rel.Siz = 4
		rel.Sym = p.To.Sym
		rel.Type = objabi.R_TLS_LE
		o1 = 0

	case 104: 
		if p.To.Sym == nil {
			c.ctxt.Diag("nil sym in tls %v", p)
		}
		if p.To.Offset != 0 {
			c.ctxt.Diag("offset against tls var in %v", p)
		}
		rel := obj.Addrel(c.cursym)
		rel.Off = int32(c.pc)
		rel.Siz = 4
		rel.Sym = p.To.Sym
		rel.Type = objabi.R_TLS_IE
		rel.Add = c.pc - p.Rel.Pc - 8 - int64(rel.Siz)

	case 68: 
		o1 = c.omvl(p, &p.To, REGTMP)

		if o1 == 0 {
			break
		}
		o2 = c.ofsr(p.As, int(p.From.Reg), 0, REGTMP, int(p.Scond), p)
		if o.flag&LPCREL != 0 {
			o3 = o2
			o2 = c.oprrr(p, AADD, int(p.Scond)) | REGTMP&15 | (REGPC&15)<<16 | (REGTMP&15)<<12
		}

	case 69: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}
		o2 = c.ofsr(p.As, int(p.To.Reg), 0, (REGTMP&15), int(p.Scond), p) | 1<<20
		if o.flag&LPCREL != 0 {
			o3 = o2
			o2 = c.oprrr(p, AADD, int(p.Scond)) | REGTMP&15 | (REGPC&15)<<16 | (REGTMP&15)<<12
		}

		
	case 70: 
		c.aclass(&p.To)

		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 = c.oshr(int(p.From.Reg), int32(c.instoffset), r, int(p.Scond))

	case 71: 
		c.aclass(&p.From)

		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o1 = c.olhr(int32(c.instoffset), r, int(p.To.Reg), int(p.Scond))
		if p.As == AMOVB || p.As == AMOVBS {
			o1 ^= 1<<5 | 1<<6
		} else if p.As == AMOVH || p.As == AMOVHS {
			o1 ^= (1 << 6)
		}

	case 72: 
		o1 = c.omvl(p, &p.To, REGTMP)

		if o1 == 0 {
			break
		}
		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 = c.oshrr(int(p.From.Reg), REGTMP&15, r, int(p.Scond))

	case 73: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}
		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}
		o2 = c.olhrr(REGTMP&15, r, int(p.To.Reg), int(p.Scond))
		if p.As == AMOVB || p.As == AMOVBS {
			o2 ^= 1<<5 | 1<<6
		} else if p.As == AMOVH || p.As == AMOVHS {
			o2 ^= (1 << 6)
		}

	case 74: 
		c.ctxt.Diag("ABX $I")

	case 75: 
		c.aclass(&p.To)

		if c.instoffset != 0 {
			c.ctxt.Diag("non-zero offset in ABX")
		}

		
		
		o1 = c.oprrr(p, AADD, int(p.Scond))

		o1 |= uint32(immrot(uint32(c.instoffset)))
		o1 |= (uint32(p.To.Reg) & 15) << 16
		o1 |= (REGTMP & 15) << 12
		o2 = c.oprrr(p, AADD, int(p.Scond)) | uint32(immrot(0)) | (REGPC&15)<<16 | (REGLINK&15)<<12 
		o3 = ((uint32(p.Scond)&C_SCOND)^C_SCOND_XOR)<<28 | 0x12fff<<8 | 1<<4 | REGTMP&15            

	case 76: 
		c.ctxt.Diag("ABXRET")

	case 77: 
		c.aclass(&p.From)

		if c.instoffset != 0 {
			c.ctxt.Diag("offset must be zero in LDREX")
		}
		o1 = 0x19<<20 | 0xf9f
		o1 |= (uint32(p.From.Reg) & 15) << 16
		o1 |= (uint32(p.To.Reg) & 15) << 12
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 78: 
		c.aclass(&p.From)

		if c.instoffset != 0 {
			c.ctxt.Diag("offset must be zero in STREX")
		}
		if p.To.Reg == p.From.Reg || p.To.Reg == p.Reg {
			c.ctxt.Diag("cannot use same register as both source and destination: %v", p)
		}
		o1 = 0x18<<20 | 0xf90
		o1 |= (uint32(p.From.Reg) & 15) << 16
		o1 |= (uint32(p.Reg) & 15) << 0
		o1 |= (uint32(p.To.Reg) & 15) << 12
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 80: 
		if p.As == AMOVD {
			o1 = 0xeeb00b00 
			o2 = c.oprrr(p, ASUBD, int(p.Scond))
		} else {
			o1 = 0x0eb00a00 
			o2 = c.oprrr(p, ASUBF, int(p.Scond))
		}

		v := int32(0x70) 
		r := (int(p.To.Reg) & 15) << 0

		
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

		o1 |= (uint32(r) & 15) << 12
		o1 |= (uint32(v) & 0xf) << 0
		o1 |= (uint32(v) & 0xf0) << 12

		
		o2 |= (uint32(r)&15)<<0 | (uint32(r)&15)<<16 | (uint32(r)&15)<<12

	case 81: 
		o1 = 0x0eb00a00 
		if p.As == AMOVD {
			o1 = 0xeeb00b00 
		}
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
		o1 |= (uint32(p.To.Reg) & 15) << 12
		v := int32(c.chipfloat5(p.From.Val.(float64)))
		o1 |= (uint32(v) & 0xf) << 0
		o1 |= (uint32(v) & 0xf0) << 12

	case 82: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.Reg)&15)<<12 | (uint32(p.From.Reg)&15)<<0
		o2 = 0x0ef1fa10 
		o2 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 83: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.From.Reg)&15)<<12 | 1<<16
		o2 = 0x0ef1fa10 
		o2 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 84: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.From.Reg) & 15) << 0
		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 85: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.From.Reg) & 15) << 0
		o1 |= (uint32(p.To.Reg) & 15) << 12

		
	case 86: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.From.Reg) & 15) << 0
		o1 |= (FREGTMP & 15) << 12
		o2 = c.oprrr(p, -AMOVFW, int(p.Scond))
		o2 |= (FREGTMP & 15) << 16
		o2 |= (uint32(p.To.Reg) & 15) << 12

		
	case 87: 
		o1 = c.oprrr(p, -AMOVWF, int(p.Scond))

		o1 |= (uint32(p.From.Reg) & 15) << 12
		o1 |= (FREGTMP & 15) << 16
		o2 = c.oprrr(p, p.As, int(p.Scond))
		o2 |= (FREGTMP & 15) << 0
		o2 |= (uint32(p.To.Reg) & 15) << 12

	case 88: 
		o1 = c.oprrr(p, -AMOVWF, int(p.Scond))

		o1 |= (uint32(p.From.Reg) & 15) << 12
		o1 |= (uint32(p.To.Reg) & 15) << 16

	case 89: 
		o1 = c.oprrr(p, -AMOVFW, int(p.Scond))

		o1 |= (uint32(p.From.Reg) & 15) << 16
		o1 |= (uint32(p.To.Reg) & 15) << 12

	case 91: 
		c.aclass(&p.From)

		if c.instoffset != 0 {
			c.ctxt.Diag("offset must be zero in LDREX")
		}
		o1 = 0x1b<<20 | 0xf9f
		o1 |= (uint32(p.From.Reg) & 15) << 16
		o1 |= (uint32(p.To.Reg) & 15) << 12
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 92: 
		c.aclass(&p.From)

		if c.instoffset != 0 {
			c.ctxt.Diag("offset must be zero in STREX")
		}
		if p.Reg&1 != 0 {
			c.ctxt.Diag("source register must be even in STREXD: %v", p)
		}
		if p.To.Reg == p.From.Reg || p.To.Reg == p.Reg || p.To.Reg == p.Reg+1 {
			c.ctxt.Diag("cannot use same register as both source and destination: %v", p)
		}
		o1 = 0x1a<<20 | 0xf90
		o1 |= (uint32(p.From.Reg) & 15) << 16
		o1 |= (uint32(p.Reg) & 15) << 0
		o1 |= (uint32(p.To.Reg) & 15) << 12
		o1 |= ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28

	case 93: 
		o1 = c.omvl(p, &p.From, REGTMP)

		if o1 == 0 {
			break
		}
		o2 = c.olhr(0, REGTMP, int(p.To.Reg), int(p.Scond))
		if p.As == AMOVB || p.As == AMOVBS {
			o2 ^= 1<<5 | 1<<6
		} else if p.As == AMOVH || p.As == AMOVHS {
			o2 ^= (1 << 6)
		}
		if o.flag&LPCREL != 0 {
			o3 = o2
			o2 = c.oprrr(p, AADD, int(p.Scond)) | REGTMP&15 | (REGPC&15)<<16 | (REGTMP&15)<<12
		}

	case 94: 
		o1 = c.omvl(p, &p.To, REGTMP)

		if o1 == 0 {
			break
		}
		o2 = c.oshr(int(p.From.Reg), 0, REGTMP, int(p.Scond))
		if o.flag&LPCREL != 0 {
			o3 = o2
			o2 = c.oprrr(p, AADD, int(p.Scond)) | REGTMP&15 | (REGPC&15)<<16 | (REGTMP&15)<<12
		}

	case 95: 
		o1 = 0xf5d0f000

		o1 |= (uint32(p.From.Reg) & 15) << 16
		if p.From.Offset < 0 {
			o1 &^= (1 << 23)
			o1 |= uint32((-p.From.Offset) & 0xfff)
		} else {
			o1 |= uint32(p.From.Offset & 0xfff)
		}

	
	
	
	
	
	case 96: 
		o1 = 0xf7fabcfd

	case 97: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.To.Reg) & 15) << 12
		o1 |= (uint32(p.From.Reg) & 15) << 0

	case 98: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.To.Reg) & 15) << 16
		o1 |= (uint32(p.From.Reg) & 15) << 8
		o1 |= (uint32(p.Reg) & 15) << 0

	case 99: 
		o1 = c.oprrr(p, p.As, int(p.Scond))

		o1 |= (uint32(p.To.Reg) & 15) << 16
		o1 |= (uint32(p.From.Reg) & 15) << 8
		o1 |= (uint32(p.Reg) & 15) << 0
		o1 |= uint32((p.To.Offset & 15) << 12)

	case 105: 
		o1 = c.oprrr(p, p.As, int(p.Scond))
		rf := int(p.From.Reg)
		rt := int(p.To.Reg)
		r := int(p.Reg)
		if r == 0 {
			r = rt
		}
		o1 |= (uint32(rf)&15)<<8 | (uint32(r)&15)<<0 | (uint32(rt)&15)<<16

	case 110: 
		o1 = 0xf57ff050
		mbop := uint32(0)

		switch c.aclass(&p.From) {
		case C_SPR:
			for _, f := range mbOp {
				if f.reg == p.From.Reg {
					mbop = f.enc
					break
				}
			}
		case C_RCON:
			for _, f := range mbOp {
				enc := uint32(c.instoffset)
				if f.enc == enc {
					mbop = enc
					break
				}
			}
		case C_NONE:
			mbop = 0xf
		}

		if mbop == 0 {
			c.ctxt.Diag("illegal mb option:\n%v", p)
		}
		o1 |= mbop
	}

	out[0] = o1
	out[1] = o2
	out[2] = o3
	out[3] = o4
	out[4] = o5
	out[5] = o6
}

func (c *ctxt5) movxt(p *obj.Prog) uint32 {
	o1 := ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
	switch p.As {
	case AMOVB, AMOVBS:
		o1 |= 0x6af<<16 | 0x7<<4
	case AMOVH, AMOVHS:
		o1 |= 0x6bf<<16 | 0x7<<4
	case AMOVBU:
		o1 |= 0x6ef<<16 | 0x7<<4
	case AMOVHU:
		o1 |= 0x6ff<<16 | 0x7<<4
	default:
		c.ctxt.Diag("illegal combination: %v", p)
	}
	switch p.From.Offset &^ 0xf {
	
	case SHIFT_RR, SHIFT_RR | 8<<7, SHIFT_RR | 16<<7, SHIFT_RR | 24<<7:
		o1 |= uint32(p.From.Offset) & 0xc0f
	default:
		c.ctxt.Diag("illegal shift: %v", p)
	}
	o1 |= (uint32(p.To.Reg) & 15) << 12
	return o1
}

func (c *ctxt5) mov(p *obj.Prog) uint32 {
	c.aclass(&p.From)
	o1 := c.oprrr(p, p.As, int(p.Scond))
	o1 |= uint32(p.From.Offset)
	rt := int(p.To.Reg)
	if p.To.Type == obj.TYPE_NONE {
		rt = 0
	}
	r := int(p.Reg)
	if p.As == AMOVW || p.As == AMVN {
		r = 0
	} else if r == 0 {
		r = rt
	}
	o1 |= (uint32(r)&15)<<16 | (uint32(rt)&15)<<12
	return o1
}

func (c *ctxt5) oprrr(p *obj.Prog, a obj.As, sc int) uint32 {
	o := ((uint32(sc) & C_SCOND) ^ C_SCOND_XOR) << 28
	if sc&C_SBIT != 0 {
		o |= 1 << 20
	}
	switch a {
	case ADIVHW:
		return o | 0x71<<20 | 0xf<<12 | 0x1<<4
	case ADIVUHW:
		return o | 0x73<<20 | 0xf<<12 | 0x1<<4
	case AMMUL:
		return o | 0x75<<20 | 0xf<<12 | 0x1<<4
	case AMULS:
		return o | 0x6<<20 | 0x9<<4
	case AMMULA:
		return o | 0x75<<20 | 0x1<<4
	case AMMULS:
		return o | 0x75<<20 | 0xd<<4
	case AMULU, AMUL:
		return o | 0x0<<21 | 0x9<<4
	case AMULA:
		return o | 0x1<<21 | 0x9<<4
	case AMULLU:
		return o | 0x4<<21 | 0x9<<4
	case AMULL:
		return o | 0x6<<21 | 0x9<<4
	case AMULALU:
		return o | 0x5<<21 | 0x9<<4
	case AMULAL:
		return o | 0x7<<21 | 0x9<<4
	case AAND:
		return o | 0x0<<21
	case AEOR:
		return o | 0x1<<21
	case ASUB:
		return o | 0x2<<21
	case ARSB:
		return o | 0x3<<21
	case AADD:
		return o | 0x4<<21
	case AADC:
		return o | 0x5<<21
	case ASBC:
		return o | 0x6<<21
	case ARSC:
		return o | 0x7<<21
	case ATST:
		return o | 0x8<<21 | 1<<20
	case ATEQ:
		return o | 0x9<<21 | 1<<20
	case ACMP:
		return o | 0xa<<21 | 1<<20
	case ACMN:
		return o | 0xb<<21 | 1<<20
	case AORR:
		return o | 0xc<<21

	case AMOVB, AMOVH, AMOVW:
		if sc&(C_PBIT|C_WBIT) != 0 {
			c.ctxt.Diag("invalid .P/.W suffix: %v", p)
		}
		return o | 0xd<<21
	case ABIC:
		return o | 0xe<<21
	case AMVN:
		return o | 0xf<<21
	case ASLL:
		return o | 0xd<<21 | 0<<5
	case ASRL:
		return o | 0xd<<21 | 1<<5
	case ASRA:
		return o | 0xd<<21 | 2<<5
	case ASWI:
		return o | 0xf<<24

	case AADDD:
		return o | 0xe<<24 | 0x3<<20 | 0xb<<8 | 0<<4
	case AADDF:
		return o | 0xe<<24 | 0x3<<20 | 0xa<<8 | 0<<4
	case ASUBD:
		return o | 0xe<<24 | 0x3<<20 | 0xb<<8 | 4<<4
	case ASUBF:
		return o | 0xe<<24 | 0x3<<20 | 0xa<<8 | 4<<4
	case AMULD:
		return o | 0xe<<24 | 0x2<<20 | 0xb<<8 | 0<<4
	case AMULF:
		return o | 0xe<<24 | 0x2<<20 | 0xa<<8 | 0<<4
	case ANMULD:
		return o | 0xe<<24 | 0x2<<20 | 0xb<<8 | 0x4<<4
	case ANMULF:
		return o | 0xe<<24 | 0x2<<20 | 0xa<<8 | 0x4<<4
	case AMULAD:
		return o | 0xe<<24 | 0xb<<8
	case AMULAF:
		return o | 0xe<<24 | 0xa<<8
	case AMULSD:
		return o | 0xe<<24 | 0xb<<8 | 0x4<<4
	case AMULSF:
		return o | 0xe<<24 | 0xa<<8 | 0x4<<4
	case ANMULAD:
		return o | 0xe<<24 | 0x1<<20 | 0xb<<8 | 0x4<<4
	case ANMULAF:
		return o | 0xe<<24 | 0x1<<20 | 0xa<<8 | 0x4<<4
	case ANMULSD:
		return o | 0xe<<24 | 0x1<<20 | 0xb<<8
	case ANMULSF:
		return o | 0xe<<24 | 0x1<<20 | 0xa<<8
	case AFMULAD:
		return o | 0xe<<24 | 0xa<<20 | 0xb<<8
	case AFMULAF:
		return o | 0xe<<24 | 0xa<<20 | 0xa<<8
	case AFMULSD:
		return o | 0xe<<24 | 0xa<<20 | 0xb<<8 | 0x4<<4
	case AFMULSF:
		return o | 0xe<<24 | 0xa<<20 | 0xa<<8 | 0x4<<4
	case AFNMULAD:
		return o | 0xe<<24 | 0x9<<20 | 0xb<<8 | 0x4<<4
	case AFNMULAF:
		return o | 0xe<<24 | 0x9<<20 | 0xa<<8 | 0x4<<4
	case AFNMULSD:
		return o | 0xe<<24 | 0x9<<20 | 0xb<<8
	case AFNMULSF:
		return o | 0xe<<24 | 0x9<<20 | 0xa<<8
	case ADIVD:
		return o | 0xe<<24 | 0x8<<20 | 0xb<<8 | 0<<4
	case ADIVF:
		return o | 0xe<<24 | 0x8<<20 | 0xa<<8 | 0<<4
	case ASQRTD:
		return o | 0xe<<24 | 0xb<<20 | 1<<16 | 0xb<<8 | 0xc<<4
	case ASQRTF:
		return o | 0xe<<24 | 0xb<<20 | 1<<16 | 0xa<<8 | 0xc<<4
	case AABSD:
		return o | 0xe<<24 | 0xb<<20 | 0<<16 | 0xb<<8 | 0xc<<4
	case AABSF:
		return o | 0xe<<24 | 0xb<<20 | 0<<16 | 0xa<<8 | 0xc<<4
	case ANEGD:
		return o | 0xe<<24 | 0xb<<20 | 1<<16 | 0xb<<8 | 0x4<<4
	case ANEGF:
		return o | 0xe<<24 | 0xb<<20 | 1<<16 | 0xa<<8 | 0x4<<4
	case ACMPD:
		return o | 0xe<<24 | 0xb<<20 | 4<<16 | 0xb<<8 | 0xc<<4
	case ACMPF:
		return o | 0xe<<24 | 0xb<<20 | 4<<16 | 0xa<<8 | 0xc<<4

	case AMOVF:
		return o | 0xe<<24 | 0xb<<20 | 0<<16 | 0xa<<8 | 4<<4
	case AMOVD:
		return o | 0xe<<24 | 0xb<<20 | 0<<16 | 0xb<<8 | 4<<4

	case AMOVDF:
		return o | 0xe<<24 | 0xb<<20 | 7<<16 | 0xa<<8 | 0xc<<4 | 1<<8 
	case AMOVFD:
		return o | 0xe<<24 | 0xb<<20 | 7<<16 | 0xa<<8 | 0xc<<4 | 0<<8 

	case AMOVWF:
		if sc&C_UBIT == 0 {
			o |= 1 << 7 
		}
		return o | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 0<<18 | 0<<8 

	case AMOVWD:
		if sc&C_UBIT == 0 {
			o |= 1 << 7 
		}
		return o | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 0<<18 | 1<<8 

	case AMOVFW:
		if sc&C_UBIT == 0 {
			o |= 1 << 16 
		}
		return o | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 1<<18 | 0<<8 | 1<<7 

	case AMOVDW:
		if sc&C_UBIT == 0 {
			o |= 1 << 16 
		}
		return o | 0xe<<24 | 0xb<<20 | 8<<16 | 0xa<<8 | 4<<4 | 1<<18 | 1<<8 | 1<<7 

	case -AMOVWF: 
		return o | 0xe<<24 | 0x0<<20 | 0xb<<8 | 1<<4

	case -AMOVFW: 
		return o | 0xe<<24 | 0x1<<20 | 0xb<<8 | 1<<4

	case -ACMP: 
		return o | 0x3<<24 | 0x5<<20

	case ABFX:
		return o | 0x3d<<21 | 0x5<<4

	case ABFXU:
		return o | 0x3f<<21 | 0x5<<4

	case ABFC:
		return o | 0x3e<<21 | 0x1f

	case ABFI:
		return o | 0x3e<<21 | 0x1<<4

	case AXTAB:
		return o | 0x6a<<20 | 0x7<<4

	case AXTAH:
		return o | 0x6b<<20 | 0x7<<4

	case AXTABU:
		return o | 0x6e<<20 | 0x7<<4

	case AXTAHU:
		return o | 0x6f<<20 | 0x7<<4

		
	case ACLZ:
		return o&(0xf<<28) | 0x16f<<16 | 0xf1<<4

	case AREV:
		return o&(0xf<<28) | 0x6bf<<16 | 0xf3<<4

	case AREV16:
		return o&(0xf<<28) | 0x6bf<<16 | 0xfb<<4

	case AREVSH:
		return o&(0xf<<28) | 0x6ff<<16 | 0xfb<<4

	case ARBIT:
		return o&(0xf<<28) | 0x6ff<<16 | 0xf3<<4

	case AMULWT:
		return o&(0xf<<28) | 0x12<<20 | 0xe<<4

	case AMULWB:
		return o&(0xf<<28) | 0x12<<20 | 0xa<<4

	case AMULBB:
		return o&(0xf<<28) | 0x16<<20 | 0x8<<4

	case AMULAWT:
		return o&(0xf<<28) | 0x12<<20 | 0xc<<4

	case AMULAWB:
		return o&(0xf<<28) | 0x12<<20 | 0x8<<4

	case AMULABB:
		return o&(0xf<<28) | 0x10<<20 | 0x8<<4

	case ABL: 
		return o&(0xf<<28) | 0x12fff3<<4
	}

	c.ctxt.Diag("%v: bad rrr %d", p, a)
	return 0
}

func (c *ctxt5) opbra(p *obj.Prog, a obj.As, sc int) uint32 {
	sc &= C_SCOND
	sc ^= C_SCOND_XOR
	if a == ABL || a == obj.ADUFFZERO || a == obj.ADUFFCOPY {
		return uint32(sc)<<28 | 0x5<<25 | 0x1<<24
	}
	if sc != 0xe {
		c.ctxt.Diag("%v: .COND on bcond instruction", p)
	}
	switch a {
	case ABEQ:
		return 0x0<<28 | 0x5<<25
	case ABNE:
		return 0x1<<28 | 0x5<<25
	case ABCS:
		return 0x2<<28 | 0x5<<25
	case ABHS:
		return 0x2<<28 | 0x5<<25
	case ABCC:
		return 0x3<<28 | 0x5<<25
	case ABLO:
		return 0x3<<28 | 0x5<<25
	case ABMI:
		return 0x4<<28 | 0x5<<25
	case ABPL:
		return 0x5<<28 | 0x5<<25
	case ABVS:
		return 0x6<<28 | 0x5<<25
	case ABVC:
		return 0x7<<28 | 0x5<<25
	case ABHI:
		return 0x8<<28 | 0x5<<25
	case ABLS:
		return 0x9<<28 | 0x5<<25
	case ABGE:
		return 0xa<<28 | 0x5<<25
	case ABLT:
		return 0xb<<28 | 0x5<<25
	case ABGT:
		return 0xc<<28 | 0x5<<25
	case ABLE:
		return 0xd<<28 | 0x5<<25
	case AB:
		return 0xe<<28 | 0x5<<25
	}

	c.ctxt.Diag("%v: bad bra %v", p, a)
	return 0
}

func (c *ctxt5) olr(v int32, b int, r int, sc int) uint32 {
	o := ((uint32(sc) & C_SCOND) ^ C_SCOND_XOR) << 28
	if sc&C_PBIT == 0 {
		o |= 1 << 24
	}
	if sc&C_UBIT == 0 {
		o |= 1 << 23
	}
	if sc&C_WBIT != 0 {
		o |= 1 << 21
	}
	o |= 1<<26 | 1<<20
	if v < 0 {
		if sc&C_UBIT != 0 {
			c.ctxt.Diag(".U on neg offset")
		}
		v = -v
		o ^= 1 << 23
	}

	if v >= 1<<12 || v < 0 {
		c.ctxt.Diag("literal span too large: %d (R%d)\n%v", v, b, c.printp)
	}
	o |= uint32(v)
	o |= (uint32(b) & 15) << 16
	o |= (uint32(r) & 15) << 12
	return o
}

func (c *ctxt5) olhr(v int32, b int, r int, sc int) uint32 {
	o := ((uint32(sc) & C_SCOND) ^ C_SCOND_XOR) << 28
	if sc&C_PBIT == 0 {
		o |= 1 << 24
	}
	if sc&C_WBIT != 0 {
		o |= 1 << 21
	}
	o |= 1<<23 | 1<<20 | 0xb<<4
	if v < 0 {
		v = -v
		o ^= 1 << 23
	}

	if v >= 1<<8 || v < 0 {
		c.ctxt.Diag("literal span too large: %d (R%d)\n%v", v, b, c.printp)
	}
	o |= uint32(v)&0xf | (uint32(v)>>4)<<8 | 1<<22
	o |= (uint32(b) & 15) << 16
	o |= (uint32(r) & 15) << 12
	return o
}

func (c *ctxt5) osr(a obj.As, r int, v int32, b int, sc int) uint32 {
	o := c.olr(v, b, r, sc) ^ (1 << 20)
	if a != AMOVW {
		o |= 1 << 22
	}
	return o
}

func (c *ctxt5) oshr(r int, v int32, b int, sc int) uint32 {
	o := c.olhr(v, b, r, sc) ^ (1 << 20)
	return o
}

func (c *ctxt5) osrr(r int, i int, b int, sc int) uint32 {
	return c.olr(int32(i), b, r, sc) ^ (1<<25 | 1<<20)
}

func (c *ctxt5) oshrr(r int, i int, b int, sc int) uint32 {
	return c.olhr(int32(i), b, r, sc) ^ (1<<22 | 1<<20)
}

func (c *ctxt5) olrr(i int, b int, r int, sc int) uint32 {
	return c.olr(int32(i), b, r, sc) ^ (1 << 25)
}

func (c *ctxt5) olhrr(i int, b int, r int, sc int) uint32 {
	return c.olhr(int32(i), b, r, sc) ^ (1 << 22)
}

func (c *ctxt5) ofsr(a obj.As, r int, v int32, b int, sc int, p *obj.Prog) uint32 {
	o := ((uint32(sc) & C_SCOND) ^ C_SCOND_XOR) << 28
	if sc&C_PBIT == 0 {
		o |= 1 << 24
	}
	if sc&C_WBIT != 0 {
		o |= 1 << 21
	}
	o |= 6<<25 | 1<<24 | 1<<23 | 10<<8
	if v < 0 {
		v = -v
		o ^= 1 << 23
	}

	if v&3 != 0 {
		c.ctxt.Diag("odd offset for floating point op: %d\n%v", v, p)
	} else if v >= 1<<10 || v < 0 {
		c.ctxt.Diag("literal span too large: %d\n%v", v, p)
	}
	o |= (uint32(v) >> 2) & 0xFF
	o |= (uint32(b) & 15) << 16
	o |= (uint32(r) & 15) << 12

	switch a {
	default:
		c.ctxt.Diag("bad fst %v", a)
		fallthrough

	case AMOVD:
		o |= 1 << 8
		fallthrough

	case AMOVF:
		break
	}

	return o
}


func (c *ctxt5) omvs(p *obj.Prog, a *obj.Addr, dr int) uint32 {
	o1 := ((uint32(p.Scond) & C_SCOND) ^ C_SCOND_XOR) << 28
	o1 |= 0x30 << 20
	o1 |= (uint32(dr) & 15) << 12
	o1 |= uint32(a.Offset) & 0x0fff
	o1 |= (uint32(a.Offset) & 0xf000) << 4
	return o1
}


func (c *ctxt5) omvr(p *obj.Prog, a *obj.Addr, dr int) uint32 {
	o1 := c.oprrr(p, AMOVW, int(p.Scond))
	o1 |= (uint32(dr) & 15) << 12
	v := immrot(^uint32(a.Offset))
	if v == 0 {
		c.ctxt.Diag("%v: missing literal", p)
		return 0
	}
	o1 |= uint32(v)
	return o1
}

func (c *ctxt5) omvl(p *obj.Prog, a *obj.Addr, dr int) uint32 {
	var o1 uint32
	if p.Pool == nil {
		c.aclass(a)
		v := immrot(^uint32(c.instoffset))
		if v == 0 {
			c.ctxt.Diag("%v: missing literal", p)
			return 0
		}

		o1 = c.oprrr(p, AMVN, int(p.Scond)&C_SCOND)
		o1 |= uint32(v)
		o1 |= (uint32(dr) & 15) << 12
	} else {
		v := int32(p.Pool.Pc - p.Pc - 8)
		o1 = c.olr(v, REGPC, dr, int(p.Scond)&C_SCOND)
	}

	return o1
}

func (c *ctxt5) chipzero5(e float64) int {
	
	if buildcfg.GOARM < 7 || math.Float64bits(e) != 0 {
		return -1
	}
	return 0
}

func (c *ctxt5) chipfloat5(e float64) int {
	
	if buildcfg.GOARM < 7 {
		return -1
	}

	ei := math.Float64bits(e)
	l := uint32(ei)
	h := uint32(ei >> 32)

	if l != 0 || h&0xffff != 0 {
		return -1
	}
	h1 := h & 0x7fc00000
	if h1 != 0x40000000 && h1 != 0x3fc00000 {
		return -1
	}
	n := 0

	
	if h&0x80000000 != 0 {
		n |= 1 << 7
	}

	
	if h1 == 0x3fc00000 {
		n |= 1 << 6
	}

	
	n |= int((h >> 16) & 0x3f)

	
	return n
}

func nocache(p *obj.Prog) {
	p.Optab = 0
	p.From.Class = 0
	if p.GetFrom3() != nil {
		p.GetFrom3().Class = 0
	}
	p.To.Class = 0
}
