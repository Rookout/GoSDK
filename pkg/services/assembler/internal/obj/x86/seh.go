// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/obj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/objabi"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/src"
	"encoding/base64"
	"fmt"
	"math"
)

type sehbuf struct {
	ctxt *obj.Link
	data []byte
	off  int
}

func newsehbuf(ctxt *obj.Link, nodes uint8) sehbuf {
	
	
	
	size := 8 + nodes*2
	if nodes%2 != 0 {
		size += 2
	}
	return sehbuf{ctxt, make([]byte, size), 0}
}

func (b *sehbuf) write8(v uint8) {
	b.data[b.off] = v
	b.off++
}

func (b *sehbuf) write32(v uint32) {
	b.ctxt.Arch.ByteOrder.PutUint32(b.data[b.off:], v)
	b.off += 4
}

func (b *sehbuf) writecode(op, value uint8) {
	b.write8(value<<4 | op)
}


func populateSeh(ctxt *obj.Link, s *obj.LSym) (sehsym *obj.LSym) {
	if s.NoFrame() {
		return
	}

	
	
	
	
	
	
	

	
	var pushbp *obj.Prog
	for p := s.Func().Text; p != nil; p = p.Link {
		if p.As == APUSHQ && p.From.Type == obj.TYPE_REG && p.From.Reg == REG_BP {
			pushbp = p
			break
		}
		if p.Pos.Xlogue() == src.PosPrologueEnd {
			break
		}
	}
	if pushbp == nil {
		ctxt.Diag("missing frame pointer instruction: PUSHQ BP")
		return
	}

	
	movbp := pushbp.Link
	if movbp == nil {
		ctxt.Diag("missing frame pointer instruction: MOVQ SP, BP")
		return
	}
	if !(movbp.As == AMOVQ && movbp.From.Type == obj.TYPE_REG && movbp.From.Reg == REG_SP &&
		movbp.To.Type == obj.TYPE_REG && movbp.To.Reg == REG_BP && movbp.From.Offset == 0) {
		ctxt.Diag("unexpected frame pointer instruction\n%v", movbp)
		return
	}
	if movbp.Link.Pc > math.MaxUint8 {
		
		
		
		
		return
	}

	
	

	const (
		UWOP_PUSH_NONVOL = 0
		UWOP_SET_FPREG   = 3
		SEH_REG_BP       = 5
	)

	
	
	
	nodes := uint8(2)
	buf := newsehbuf(ctxt, nodes)
	buf.write8(1)                    
	buf.write8(uint8(movbp.Link.Pc)) 
	buf.write8(nodes)                
	buf.write8(SEH_REG_BP)           

	
	buf.write8(uint8(movbp.Link.Pc))
	buf.writecode(UWOP_SET_FPREG, 0)

	buf.write8(uint8(pushbp.Link.Pc))
	buf.writecode(UWOP_PUSH_NONVOL, SEH_REG_BP)

	
	
	buf.write32(0)

	
	
	
	
	hash := base64.StdEncoding.EncodeToString(buf.data)
	symname := fmt.Sprintf("%d.%s", len(buf.data), hash)
	return ctxt.LookupInit("go:sehuw."+symname, func(s *obj.LSym) {
		s.WriteBytes(ctxt, 0, buf.data)
		s.Type = objabi.SSEHUNWINDINFO
		s.Set(obj.AttrDuplicateOK, true)
		s.Set(obj.AttrLocal, true)
		
		
		
	})
}
