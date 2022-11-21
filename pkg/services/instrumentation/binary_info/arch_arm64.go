//go:build arm64
// +build arm64

package binary_info

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/frame"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/op"
	"strings"
)

var arm64BreakInstruction = []byte{0x0, 0x0, 0x20, 0xd4}

const (
	DwarfPCRegNum                uint64 = 32
	DwarfSPRegNum                uint64 = 31
	DwarfLRRegNum                uint64 = 30
	DwarfBPRegNum                uint64 = 29
	DwarfV0RegNum                uint64 = 64
	_ARM64_MaxRegNum                    = DwarfV0RegNum + 31
	crosscall2SPOffsetNonWindows        = 0x58
)

func FixFrameUnwindContext(fctxt *frame.FrameContext, pc uint64, bi *BinaryInfo) *frame.FrameContext {
	if fctxt == nil || (bi.sigreturnfn != nil && pc >= bi.sigreturnfn.Entry && pc < bi.sigreturnfn.End) {
		
		
		
		
		

		
		
		
		
		
		
		
		
		
		
		

		return &frame.FrameContext{
			RetAddrReg: DwarfPCRegNum,
			Regs: map[uint64]frame.DWRule{
				DwarfPCRegNum: frame.DWRule{
					Rule:   frame.RuleOffset,
					Offset: int64(-bi.PointerSize),
				},
				DwarfBPRegNum: frame.DWRule{
					Rule:   frame.RuleOffset,
					Offset: int64(-2 * bi.PointerSize),
				},
				DwarfSPRegNum: frame.DWRule{
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
			rule.Offset += crosscall2SPOffsetNonWindows
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
	if fctxt.Regs[DwarfLRRegNum].Rule == frame.RuleUndefined {
		fctxt.Regs[DwarfLRRegNum] = frame.DWRule{
			Rule:   frame.RuleFramePointer,
			Reg:    DwarfLRRegNum,
			Offset: 0,
		}
	}

	return fctxt
}

const arm64cgocallSPOffsetSaveSlot = 0x8
const prevG0schedSPOffsetSaveSlot = 0x10

func RegSize(regnum uint64) int {
	
	if regnum >= 64 && regnum <= 95 {
		return 16
	}

	return 8 
}

func ARM64ToName(num uint64) string {
	switch {
	case num <= 30:
		return fmt.Sprintf("X%d", num)
	case num == DwarfSPRegNum:
		return "SP"
	case num == DwarfPCRegNum:
		return "PC"
	case num >= DwarfV0RegNum && num <= 95:
		return fmt.Sprintf("V%d", num-64)
	default:
		return fmt.Sprintf("unknown%d", num)
	}
}

var NameToDwarf = func() map[string]int {
	r := make(map[string]int)
	for i := 0; i <= 32; i++ {
		r[fmt.Sprintf("x%d", i)] = i
	}
	r["fp"] = 29
	r["lr"] = int(DwarfLRRegNum)
	r["sp"] = 31
	r["pc"] = int(DwarfPCRegNum)
	for i := 0; i <= 31; i++ {
		r[fmt.Sprintf("v%d", i)] = i + 64
	}
	return r
}()

func RegistersToDwarfRegisters(staticBase uint64, regs registers.Registers) op.DwarfRegisters {
	dregs := initDwarfRegistersFromSlice(int(_ARM64_MaxRegNum), regs, NameToDwarf)
	dr := op.NewDwarfRegisters(staticBase, dregs, binary.LittleEndian, DwarfPCRegNum, DwarfSPRegNum, DwarfBPRegNum, DwarfLRRegNum)
	dr.SetLoadMoreCallback(loadMoreDwarfRegistersFromSliceFunc(dr, regs, NameToDwarf))
	return *dr
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

func arm64AddrAndStackRegsToDwarfRegisters(staticBase, pc, sp, bp, lr uint64) op.DwarfRegisters {
	dregs := make([]*op.DwarfRegister, DwarfPCRegNum+1)
	dregs[DwarfPCRegNum] = op.DwarfRegisterFromUint64(pc)
	dregs[DwarfSPRegNum] = op.DwarfRegisterFromUint64(sp)
	dregs[DwarfBPRegNum] = op.DwarfRegisterFromUint64(bp)
	dregs[DwarfLRRegNum] = op.DwarfRegisterFromUint64(lr)

	return *op.NewDwarfRegisters(staticBase, dregs, binary.LittleEndian, DwarfPCRegNum, DwarfSPRegNum, DwarfBPRegNum, DwarfLRRegNum)
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

func arm64DwarfRegisterToString(i int, reg *op.DwarfRegister) (name string, floatingPoint bool, repr string) {
	name = ARM64ToName(uint64(i))

	if reg == nil {
		return name, false, ""
	}

	if reg.Bytes != nil && name[0] == 'V' {
		buf := bytes.NewReader(reg.Bytes)

		var out bytes.Buffer
		var vi [16]uint8
		for i := range vi {
			_ = binary.Read(buf, binary.LittleEndian, &vi[i])
		}
		
		fmt.Fprintf(&out, " {\n\tD = {u = {0x%02x%02x%02x%02x%02x%02x%02x%02x,", vi[7], vi[6], vi[5], vi[4], vi[3], vi[2], vi[1], vi[0])
		fmt.Fprintf(&out, " 0x%02x%02x%02x%02x%02x%02x%02x%02x},", vi[15], vi[14], vi[13], vi[12], vi[11], vi[10], vi[9], vi[8])
		fmt.Fprintf(&out, " s = {0x%02x%02x%02x%02x%02x%02x%02x%02x,", vi[7], vi[6], vi[5], vi[4], vi[3], vi[2], vi[1], vi[0])
		fmt.Fprintf(&out, " 0x%02x%02x%02x%02x%02x%02x%02x%02x}},", vi[15], vi[14], vi[13], vi[12], vi[11], vi[10], vi[9], vi[8])

		
		fmt.Fprintf(&out, " \n\tS = {u = {0x%02x%02x%02x%02x,0x%02x%02x%02x%02x,", vi[3], vi[2], vi[1], vi[0], vi[7], vi[6], vi[5], vi[4])
		fmt.Fprintf(&out, " 0x%02x%02x%02x%02x,0x%02x%02x%02x%02x},", vi[11], vi[10], vi[9], vi[8], vi[15], vi[14], vi[13], vi[12])
		fmt.Fprintf(&out, " s = {0x%02x%02x%02x%02x,0x%02x%02x%02x%02x,", vi[3], vi[2], vi[1], vi[0], vi[7], vi[6], vi[5], vi[4])
		fmt.Fprintf(&out, " 0x%02x%02x%02x%02x,0x%02x%02x%02x%02x}},", vi[11], vi[10], vi[9], vi[8], vi[15], vi[14], vi[13], vi[12])

		
		fmt.Fprintf(&out, " \n\tH = {u = {0x%02x%02x,0x%02x%02x,0x%02x%02x,0x%02x%02x,", vi[1], vi[0], vi[3], vi[2], vi[5], vi[4], vi[7], vi[6])
		fmt.Fprintf(&out, " 0x%02x%02x,0x%02x%02x,0x%02x%02x,0x%02x%02x},", vi[9], vi[8], vi[11], vi[10], vi[13], vi[12], vi[15], vi[14])
		fmt.Fprintf(&out, " s = {0x%02x%02x,0x%02x%02x,0x%02x%02x,0x%02x%02x,", vi[1], vi[0], vi[3], vi[2], vi[5], vi[4], vi[7], vi[6])
		fmt.Fprintf(&out, " 0x%02x%02x,0x%02x%02x,0x%02x%02x,0x%02x%02x}},", vi[9], vi[8], vi[11], vi[10], vi[13], vi[12], vi[15], vi[14])

		
		fmt.Fprintf(&out, " \n\tB = {u = {0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,", vi[0], vi[1], vi[2], vi[3], vi[4], vi[5], vi[6], vi[7])
		fmt.Fprintf(&out, " 0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x},", vi[8], vi[9], vi[10], vi[11], vi[12], vi[13], vi[14], vi[15])
		fmt.Fprintf(&out, " s = {0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,", vi[0], vi[1], vi[2], vi[3], vi[4], vi[5], vi[6], vi[7])
		fmt.Fprintf(&out, " 0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x}}", vi[8], vi[9], vi[10], vi[11], vi[12], vi[13], vi[14], vi[15])

		
		fmt.Fprintf(&out, " \n\tQ = {u = {0x%02x%02x%02x%02x%02x%02x%02x%02x", vi[15], vi[14], vi[13], vi[12], vi[11], vi[10], vi[9], vi[8])
		fmt.Fprintf(&out, "%02x%02x%02x%02x%02x%02x%02x%02x},", vi[7], vi[6], vi[5], vi[4], vi[3], vi[2], vi[1], vi[0])
		fmt.Fprintf(&out, " s = {0x%02x%02x%02x%02x%02x%02x%02x%02x", vi[15], vi[14], vi[13], vi[12], vi[11], vi[10], vi[9], vi[8])
		fmt.Fprintf(&out, "%02x%02x%02x%02x%02x%02x%02x%02x}}\n\t}", vi[7], vi[6], vi[5], vi[4], vi[3], vi[2], vi[1], vi[0])
		return name, true, out.String()
	} else if reg.Bytes == nil || (reg.Bytes != nil && len(reg.Bytes) < 16) {
		return name, false, fmt.Sprintf("%#016x", reg.Uint64Val)
	}
	return name, false, fmt.Sprintf("%#x", reg.Bytes)
}
