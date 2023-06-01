// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

//go:build amd64
// +build amd64

package binary_info

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/frame"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
)

const (
	DwarfIPRegNum uint64 = 16
	DwarfSPRegNum uint64 = 7
	DwarfBPRegNum uint64 = 6
)

func FixFrameUnwindContext(fctxt *frame.FrameContext, pc uint64, bi *BinaryInfo) *frame.FrameContext {
	if fctxt == nil || (bi.sigreturnfn != nil && pc >= bi.sigreturnfn.Entry && pc < bi.sigreturnfn.End) {
		
		
		
		
		

		
		
		
		
		
		
		
		
		
		
		

		return &frame.FrameContext{
			RetAddrReg: DwarfIPRegNum,
			Regs: map[uint64]frame.DWRule{
				DwarfIPRegNum: {
					Rule:   frame.RuleOffset,
					Offset: int64(-bi.PointerSize),
				},
				DwarfBPRegNum: {
					Rule:   frame.RuleOffset,
					Offset: int64(-2 * bi.PointerSize),
				},
				DwarfSPRegNum: {
					Rule:   frame.RuleValOffset,
					Offset: 0,
				},
			},
			CFA: frame.DWRule{
				Rule:   frame.RuleCFA,
				Reg:    DwarfBPRegNum,
				Offset: int64(2 * bi.PointerSize),
			},
		}
	}

	if bi.crosscall2fn != nil && pc >= bi.crosscall2fn.Entry && pc < bi.crosscall2fn.End {
		rule := fctxt.CFA
		if rule.Offset == crosscall2SPOffsetBad {
			rule.Offset += crosscall2SPOffset
		}
		fctxt.CFA = rule
	}

	
	
	
	
	if fctxt.Regs[DwarfBPRegNum].Rule == frame.RuleUndefined {
		fctxt.Regs[DwarfBPRegNum] = frame.DWRule{
			Rule:   frame.RuleFramePointer,
			Reg:    DwarfBPRegNum,
			Offset: 0,
		}
	}

	return fctxt
}






func RegSize(regnum uint64) int {
	
	if regnum > DwarfIPRegNum && regnum <= 32 {
		return 16
	}
	
	if regnum >= 33 && regnum <= 40 {
		return 10
	}
	return 8
}






var amd64DwarfToName = map[int]string{
	0:  "Rax",
	1:  "Rdx",
	2:  "Rcx",
	3:  "Rbx",
	4:  "Rsi",
	5:  "Rdi",
	6:  "Rbp",
	7:  "Rsp",
	8:  "R8",
	9:  "R9",
	10: "R10",
	11: "R11",
	12: "R12",
	13: "R13",
	14: "R14",
	15: "R15",
	16: "Rip",
	17: "XMM0",
	18: "XMM1",
	19: "XMM2",
	20: "XMM3",
	21: "XMM4",
	22: "XMM5",
	23: "XMM6",
	24: "XMM7",
	25: "XMM8",
	26: "XMM9",
	27: "XMM10",
	28: "XMM11",
	29: "XMM12",
	30: "XMM13",
	31: "XMM14",
	32: "XMM15",
	33: "ST(0)",
	34: "ST(1)",
	35: "ST(2)",
	36: "ST(3)",
	37: "ST(4)",
	38: "ST(5)",
	39: "ST(6)",
	40: "ST(7)",
	49: "Rflags",
	50: "Es",
	51: "Cs",
	52: "Ss",
	53: "Ds",
	54: "Fs",
	55: "Gs",
	58: "Fs_base",
	59: "Gs_base",
	64: "MXCSR",
	65: "CW",
	66: "SW",
}

var amd64NameToDwarf = func() map[string]int {
	r := make(map[string]int)
	for regNum, regName := range amd64DwarfToName {
		r[strings.ToLower(regName)] = regNum
	}
	r["eflags"] = 49
	r["st0"] = 33
	r["st1"] = 34
	r["st2"] = 35
	r["st3"] = 36
	r["st4"] = 37
	r["st5"] = 38
	r["st6"] = 39
	r["st7"] = 40
	return r
}()

func maxAmd64DwarfRegister() int {
	max := int(DwarfIPRegNum)
	for i := range amd64DwarfToName {
		if i > max {
			max = i
		}
	}
	return max
}

func RegistersToDwarfRegisters(staticBase uint64, regs registers.Registers) op.DwarfRegisters {
	dregs := initDwarfRegistersFromSlice(maxAmd64DwarfRegister(), regs, amd64NameToDwarf)
	dr := op.NewDwarfRegisters(staticBase, dregs, binary.LittleEndian, DwarfIPRegNum, DwarfSPRegNum, DwarfBPRegNum, 0) 
	dr.SetLoadMoreCallback(loadMoreDwarfRegistersFromSliceFunc(dr, regs, amd64NameToDwarf))
	return *dr
}

func initDwarfRegistersFromSlice(maxRegs int, regs registers.Registers, nameToDwarf map[string]int) []*op.DwarfRegister {
	dregs := make([]*op.DwarfRegister, maxRegs+1)
	regslice, _ := regs.Slice(false)
	for _, reg := range regslice {
		if dwarfReg, ok := nameToDwarf[strings.ToLower(reg.Name)]; ok {
			dregs[dwarfReg] = reg.Reg
		}
	}
	return dregs
}

func loadMoreDwarfRegistersFromSliceFunc(dr *op.DwarfRegisters, regs registers.Registers, nameToDwarf map[string]int) func() {
	return func() {
		regslice, err := regs.Slice(true)
		dr.FloatLoadError = err
		for _, reg := range regslice {
			name := strings.ToLower(reg.Name)
			if dwarfReg, ok := nameToDwarf[name]; ok {
				dr.AddReg(uint64(dwarfReg), reg.Reg)
			} else if reg.Reg.Bytes != nil && (strings.HasPrefix(name, "ymm") || strings.HasPrefix(name, "zmm")) {
				xmmIdx, ok := nameToDwarf["x"+name[1:]]
				if !ok {
					continue
				}
				xmmReg := dr.Reg(uint64(xmmIdx))
				if xmmReg == nil || xmmReg.Bytes == nil {
					continue
				}
				nb := make([]byte, 0, len(xmmReg.Bytes)+len(reg.Reg.Bytes))
				nb = append(nb, xmmReg.Bytes...)
				nb = append(nb, reg.Reg.Bytes...)
				xmmReg.Bytes = nb
			}
		}
	}
}

func AddrAndStackRegsToDwarfRegisters(staticBase, pc, sp, bp, lr uint64) op.DwarfRegisters {
	dregs := make([]*op.DwarfRegister, DwarfIPRegNum+1)
	dregs[DwarfIPRegNum] = op.DwarfRegisterFromUint64(pc)
	dregs[DwarfSPRegNum] = op.DwarfRegisterFromUint64(sp)
	dregs[DwarfBPRegNum] = op.DwarfRegisterFromUint64(bp)

	return *op.NewDwarfRegisters(staticBase, dregs, binary.LittleEndian, DwarfIPRegNum, DwarfSPRegNum, DwarfBPRegNum, 0)
}

func formatSSEReg(name string, reg []byte) string {
	out := new(bytes.Buffer)
	formatSSERegInternal(reg, out)
	if len(reg) < 32 {
		return out.String()
	}

	fmt.Fprintf(out, "\n\t[%sh] ", "Y"+name[1:])
	formatSSERegInternal(reg[16:], out)

	if len(reg) < 64 {
		return out.String()
	}

	fmt.Fprintf(out, "\n\t[%shl] ", "Z"+name[1:])
	formatSSERegInternal(reg[32:], out)
	fmt.Fprintf(out, "\n\t[%shh] ", "Z"+name[1:])
	formatSSERegInternal(reg[48:], out)

	return out.String()
}

func formatSSERegInternal(xmm []byte, out *bytes.Buffer) {
	buf := bytes.NewReader(xmm)

	var vi [16]uint8
	for i := range vi {
		binary.Read(buf, binary.LittleEndian, &vi[i])
	}

	fmt.Fprintf(out, "0x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x", vi[15], vi[14], vi[13], vi[12], vi[11], vi[10], vi[9], vi[8], vi[7], vi[6], vi[5], vi[4], vi[3], vi[2], vi[1], vi[0])

	fmt.Fprintf(out, "\tv2_int={ %02x%02x%02x%02x%02x%02x%02x%02x %02x%02x%02x%02x%02x%02x%02x%02x }", vi[7], vi[6], vi[5], vi[4], vi[3], vi[2], vi[1], vi[0], vi[15], vi[14], vi[13], vi[12], vi[11], vi[10], vi[9], vi[8])

	fmt.Fprintf(out, "\tv4_int={ %02x%02x%02x%02x %02x%02x%02x%02x %02x%02x%02x%02x %02x%02x%02x%02x }", vi[3], vi[2], vi[1], vi[0], vi[7], vi[6], vi[5], vi[4], vi[11], vi[10], vi[9], vi[8], vi[15], vi[14], vi[13], vi[12])

	fmt.Fprintf(out, "\tv8_int={ %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x }", vi[1], vi[0], vi[3], vi[2], vi[5], vi[4], vi[7], vi[6], vi[9], vi[8], vi[11], vi[10], vi[13], vi[12], vi[15], vi[14])

	fmt.Fprintf(out, "\tv16_int={ %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x }", vi[0], vi[1], vi[2], vi[3], vi[4], vi[5], vi[6], vi[7], vi[8], vi[9], vi[10], vi[11], vi[12], vi[13], vi[14], vi[15])

	buf.Seek(0, io.SeekStart)
	var v2 [2]float64
	for i := range v2 {
		binary.Read(buf, binary.LittleEndian, &v2[i])
	}
	fmt.Fprintf(out, "\tv2_float={ %g %g }", v2[0], v2[1])

	buf.Seek(0, io.SeekStart)
	var v4 [4]float32
	for i := range v4 {
		binary.Read(buf, binary.LittleEndian, &v4[i])
	}
	fmt.Fprintf(out, "\tv4_float={ %g %g %g %g }", v4[0], v4[1], v4[2], v4[3])
}

func formatX87Reg(b []byte) string {
	if len(b) < 10 {
		return fmt.Sprintf("%#x", b)
	}
	mantissa := binary.LittleEndian.Uint64(b[:8])
	exponent := uint16(binary.LittleEndian.Uint16(b[8:]))

	var f float64
	fset := false

	const (
		_SIGNBIT    = 1 << 15
		_EXP_BIAS   = (1 << 14) - 1 
		_SPECIALEXP = (1 << 15) - 1 
		_HIGHBIT    = 1 << 63
		_QUIETBIT   = 1 << 62
	)

	sign := 1.0
	if exponent&_SIGNBIT != 0 {
		sign = -1.0
	}
	exponent &= ^uint16(_SIGNBIT)

	NaN := math.NaN()
	Inf := math.Inf(+1)

	switch exponent {
	case 0:
		switch {
		case mantissa == 0:
			f = sign * 0.0
			fset = true
		case mantissa&_HIGHBIT != 0:
			f = NaN
			fset = true
		}
	case _SPECIALEXP:
		switch {
		case mantissa&_HIGHBIT == 0:
			f = sign * Inf
			fset = true
		default:
			f = NaN 
			fset = true
		}
	default:
		if mantissa&_HIGHBIT == 0 {
			f = NaN
			fset = true
		}
	}

	if !fset {
		significand := float64(mantissa) / (1 << 63)
		f = sign * math.Ldexp(significand, int(exponent-_EXP_BIAS))
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, exponent)
	binary.Write(&buf, binary.LittleEndian, mantissa)

	return fmt.Sprintf("%#04x%016x\t%g", exponent, mantissa, f)
}
