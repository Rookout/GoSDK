// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

// This file implements the encoding of source positions.

package src

import (
	"bytes"
	"fmt"
	"io"
)














type Pos struct {
	base *PosBase
	lico
}


var NoPos Pos



func MakePos(base *PosBase, line, col uint) Pos {
	return Pos{base, makeLico(line, col)}
}




func (p Pos) IsKnown() bool {
	return p.base != nil || p.Line() != 0
}



func (p Pos) Before(q Pos) bool {
	n, m := p.Filename(), q.Filename()
	return n < m || n == m && p.lico < q.lico
}



func (p Pos) After(q Pos) bool {
	n, m := p.Filename(), q.Filename()
	return n > m || n == m && p.lico > q.lico
}

func (p Pos) LineNumber() string {
	if !p.IsKnown() {
		return "?"
	}
	return p.lico.lineNumber()
}

func (p Pos) LineNumberHTML() string {
	if !p.IsKnown() {
		return "?"
	}
	return p.lico.lineNumberHTML()
}


func (p Pos) Filename() string { return p.base.Pos().RelFilename() }


func (p Pos) Base() *PosBase { return p.base }


func (p *Pos) SetBase(base *PosBase) { p.base = base }


func (p Pos) RelFilename() string { return p.base.Filename() }


func (p Pos) RelLine() uint {
	b := p.base
	if b.Line() == 0 {
		
		return 0
	}
	return b.Line() + (p.Line() - b.Pos().Line())
}


func (p Pos) RelCol() uint {
	b := p.base
	if b.Col() == 0 {
		
		
		
		
		return 0
	}
	if p.Line() == b.Pos().Line() {
		
		return b.Col() + (p.Col() - b.Pos().Col())
	}
	return p.Col()
}


func (p Pos) AbsFilename() string { return p.base.AbsFilename() }



func (p Pos) SymFilename() string { return p.base.SymFilename() }

func (p Pos) String() string {
	return p.Format(true, true)
}






func (p Pos) Format(showCol, showOrig bool) string {
	buf := new(bytes.Buffer)
	p.WriteTo(buf, showCol, showOrig)
	return buf.String()
}


func (p Pos) WriteTo(w io.Writer, showCol, showOrig bool) {
	if !p.IsKnown() {
		io.WriteString(w, "<unknown line number>")
		return
	}

	if b := p.base; b == b.Pos().base {
		
		format(w, p.Filename(), p.Line(), p.Col(), showCol)
		return
	}

	
	
	
	
	
	
	
	
	format(w, p.RelFilename(), p.RelLine(), p.RelCol(), showCol)
	if showOrig {
		io.WriteString(w, "[")
		format(w, p.Filename(), p.Line(), p.Col(), showCol)
		io.WriteString(w, "]")
	}
}



func format(w io.Writer, filename string, line, col uint, showCol bool) {
	io.WriteString(w, filename)
	io.WriteString(w, ":")
	fmt.Fprint(w, line)
	
	if showCol && 0 < col && col < colMax {
		io.WriteString(w, ":")
		fmt.Fprint(w, col)
	}
}


func formatstr(filename string, line, col uint, showCol bool) string {
	buf := new(bytes.Buffer)
	format(buf, filename, line, col, showCol)
	return buf.String()
}






type PosBase struct {
	pos         Pos    
	filename    string 
	absFilename string 
	symFilename string 
	line, col   uint   
	inl         int    
}



func NewFileBase(filename, absFilename string) *PosBase {
	base := &PosBase{
		filename:    filename,
		absFilename: absFilename,
		symFilename: FileSymPrefix + absFilename,
		line:        1,
		col:         1,
		inl:         -1,
	}
	base.pos = MakePos(base, 1, 1)
	return base
}







func NewLinePragmaBase(pos Pos, filename, absFilename string, line, col uint) *PosBase {
	return &PosBase{pos, filename, absFilename, FileSymPrefix + absFilename, line, col, -1}
}



func NewInliningBase(old *PosBase, inlTreeIndex int) *PosBase {
	if old == nil {
		base := &PosBase{line: 1, col: 1, inl: inlTreeIndex}
		base.pos = MakePos(base, 1, 1)
		return base
	}
	copy := *old
	base := &copy
	base.inl = inlTreeIndex
	if old == old.pos.base {
		base.pos.base = base
	}
	return base
}

var noPos Pos



func (b *PosBase) Pos() *Pos {
	if b != nil {
		return &b.pos
	}
	return &noPos
}



func (b *PosBase) Filename() string {
	if b != nil {
		return b.filename
	}
	return ""
}



func (b *PosBase) AbsFilename() string {
	if b != nil {
		return b.absFilename
	}
	return ""
}

const FileSymPrefix = "gofile.."




func (b *PosBase) SymFilename() string {
	if b != nil {
		return b.symFilename
	}
	return FileSymPrefix + "??"
}



func (b *PosBase) Line() uint {
	if b != nil {
		return b.line
	}
	return 0
}



func (b *PosBase) Col() uint {
	if b != nil {
		return b.col
	}
	return 0
}




func (b *PosBase) InliningIndex() int {
	if b != nil {
		return b.inl
	}
	return -1
}





type lico uint32











const (
	lineBits, lineMax     = 20, 1<<lineBits - 2
	bogusLine             = 1 
	isStmtBits, isStmtMax = 2, 1<<isStmtBits - 1
	xlogueBits, xlogueMax = 2, 1<<xlogueBits - 1
	colBits, colMax       = 32 - lineBits - xlogueBits - isStmtBits, 1<<colBits - 1

	isStmtShift = 0
	isStmtMask  = isStmtMax << isStmtShift
	xlogueShift = isStmtBits + isStmtShift
	xlogueMask  = xlogueMax << xlogueShift
	colShift    = xlogueBits + xlogueShift
	lineShift   = colBits + colShift
)
const (
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	PosDefaultStmt uint = iota 
	PosIsStmt                  
	PosNotStmt                 
)

type PosXlogue uint

const (
	PosDefaultLogue PosXlogue = iota
	PosPrologueEnd
	PosEpilogueBegin
)

func makeLicoRaw(line, col uint) lico {
	return lico(line<<lineShift | col<<colShift)
}



func makeBogusLico() lico {
	return makeLicoRaw(bogusLine, 0).withIsStmt()
}

func makeLico(line, col uint) lico {
	if line >= lineMax {
		
		line = lineMax
		
		
		col = 0
	}
	if col > colMax {
		
		col = colMax
	}
	
	return makeLicoRaw(line, col)
}

func (x lico) Line() uint           { return uint(x) >> lineShift }
func (x lico) SameLine(y lico) bool { return 0 == (x^y)&^lico(1<<lineShift-1) }
func (x lico) Col() uint            { return uint(x) >> colShift & colMax }
func (x lico) IsStmt() uint {
	if x == 0 {
		return PosNotStmt
	}
	return uint(x) >> isStmtShift & isStmtMax
}
func (x lico) Xlogue() PosXlogue {
	return PosXlogue(uint(x) >> xlogueShift & xlogueMax)
}


func (x lico) withNotStmt() lico {
	return x.withStmt(PosNotStmt)
}


func (x lico) withDefaultStmt() lico {
	return x.withStmt(PosDefaultStmt)
}


func (x lico) withIsStmt() lico {
	return x.withStmt(PosIsStmt)
}


func (x lico) withXlogue(xlogue PosXlogue) lico {
	if x == 0 {
		if xlogue == 0 {
			return x
		}
		
		x = lico(PosNotStmt << isStmtShift)
	}
	return lico(uint(x) & ^uint(xlogueMax<<xlogueShift) | (uint(xlogue) << xlogueShift))
}


func (x lico) withStmt(stmt uint) lico {
	if x == 0 {
		return lico(0)
	}
	return lico(uint(x) & ^uint(isStmtMax<<isStmtShift) | (stmt << isStmtShift))
}

func (x lico) lineNumber() string {
	return fmt.Sprintf("%d", x.Line())
}

func (x lico) lineNumberHTML() string {
	if x.IsStmt() == PosDefaultStmt {
		return fmt.Sprintf("%d", x.Line())
	}
	style, pfx := "b", "+"
	if x.IsStmt() == PosNotStmt {
		style = "s" 
		pfx = ""
	}
	return fmt.Sprintf("<%s>%s%d</%s>", style, pfx, x.Line(), style)
}

func (x lico) atColumn1() lico {
	return makeLico(x.Line(), 1).withIsStmt()
}
