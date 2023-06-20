// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements the compressed encoding of source
// positions using a lookup table.

package src


type XPos struct {
	index int32
	lico
}


var NoXPos XPos




func (p XPos) IsKnown() bool {
	return p.index != 0 || p.Line() != 0
}



func (p XPos) Before(q XPos) bool {
	n, m := p.index, q.index
	return n < m || n == m && p.lico < q.lico
}


func (p XPos) SameFile(q XPos) bool {
	return p.index == q.index
}


func (p XPos) SameFileAndLine(q XPos) bool {
	return p.index == q.index && p.lico.SameLine(q.lico)
}



func (p XPos) After(q XPos) bool {
	n, m := p.index, q.index
	return n > m || n == m && p.lico > q.lico
}


func (p XPos) WithNotStmt() XPos {
	p.lico = p.lico.withNotStmt()
	return p
}


func (p XPos) WithDefaultStmt() XPos {
	p.lico = p.lico.withDefaultStmt()
	return p
}


func (p XPos) WithIsStmt() XPos {
	p.lico = p.lico.withIsStmt()
	return p
}






func (p XPos) WithBogusLine() XPos {
	if p.index == 0 {
		
		panic("Assigning a bogus line to XPos with no file will cause mysterious downstream failures.")
	}
	p.lico = makeBogusLico()
	return p
}


func (p XPos) WithXlogue(x PosXlogue) XPos {
	p.lico = p.lico.withXlogue(x)
	return p
}


func (p XPos) LineNumber() string {
	if !p.IsKnown() {
		return "?"
	}
	return p.lico.lineNumber()
}




func (p XPos) FileIndex() int32 {
	return p.index
}

func (p XPos) LineNumberHTML() string {
	if !p.IsKnown() {
		return "?"
	}
	return p.lico.lineNumberHTML()
}


func (p XPos) AtColumn1() XPos {
	p.lico = p.lico.atColumn1()
	return p
}



type PosTable struct {
	baseList []*PosBase
	indexMap map[*PosBase]int
	nameMap  map[string]int 
}



func (t *PosTable) XPos(pos Pos) XPos {
	m := t.indexMap
	if m == nil {
		
		
		t.baseList = append(t.baseList, nil)
		m = map[*PosBase]int{nil: 0}
		t.indexMap = m
		t.nameMap = make(map[string]int)
	}
	i, ok := m[pos.base]
	if !ok {
		i = len(t.baseList)
		t.baseList = append(t.baseList, pos.base)
		t.indexMap[pos.base] = i
		if _, ok := t.nameMap[pos.base.symFilename]; !ok {
			t.nameMap[pos.base.symFilename] = len(t.nameMap)
		}
	}
	return XPos{int32(i), pos.lico}
}



func (t *PosTable) Pos(p XPos) Pos {
	var base *PosBase
	if p.index != 0 {
		base = t.baseList[p.index]
	}
	return Pos{base, p.lico}
}


func (t *PosTable) FileIndex(filename string) int {
	if v, ok := t.nameMap[filename]; ok {
		return v
	}
	return -1
}


func (t *PosTable) FileTable() []string {
	
	
	
	fileLUT := make([]string, len(t.nameMap))
	for str, i := range t.nameMap {
		fileLUT[i] = str
	}
	return fileLUT
}
