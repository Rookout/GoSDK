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

package godwarf

import (
	"debug/dwarf"
	"fmt"
	"sort"
	"sync"
)




type Entry interface {
	Val(dwarf.Attr) interface{}
}

type compositeEntry []*dwarf.Entry

func (ce compositeEntry) Val(attr dwarf.Attr) interface{} {
	for _, e := range ce {
		if r := e.Val(attr); r != nil {
			return r
		}
	}
	return nil
}




func LoadAbstractOrigin(entry *dwarf.Entry, aordr *dwarf.Reader) (Entry, dwarf.Offset) {
	ao, ok := entry.Val(dwarf.AttrAbstractOrigin).(dwarf.Offset)
	if !ok {
		return entry, entry.Offset
	}

	r := []*dwarf.Entry{entry}

	for {
		aordr.Seek(ao)
		e, _ := aordr.Next()
		if e == nil {
			break
		}
		r = append(r, e)

		ao, ok = e.Val(dwarf.AttrAbstractOrigin).(dwarf.Offset)
		if !ok {
			break
		}
	}

	return compositeEntry(r), entry.Offset
}


type Tree struct {
	Entry
	typ      Type
	Tag      dwarf.Tag
	Offset   dwarf.Offset
	Ranges   [][2]uint64
	Children []*Tree
}






func LoadTree(off dwarf.Offset, dw *dwarf.Data, staticBase uint64) (*Tree, error) {
	rdr := dw.Reader()
	rdr.Seek(off)

	e, err := rdr.Next()
	if err != nil {
		return nil, err
	}
	r := entryToTreeInternal(e)
	r.Children, err = loadTreeChildren(e, rdr)
	if err != nil {
		return nil, err
	}

	err = r.resolveRanges(dw, staticBase)
	if err != nil {
		return nil, err
	}
	r.resolveAbstractEntries(rdr)

	return r, nil
}


func EntryToTree(entry *dwarf.Entry) *Tree {
	if entry.Children {
		panic(fmt.Sprintf("EntryToTree called on entry with children; "+
			"LoadTree should have been used instead. entry: %+v", entry))
	}
	return entryToTreeInternal(entry)
}

func entryToTreeInternal(entry *dwarf.Entry) *Tree {
	return &Tree{Entry: entry, Offset: entry.Offset, Tag: entry.Tag}
}

func loadTreeChildren(e *dwarf.Entry, rdr *dwarf.Reader) ([]*Tree, error) {
	if !e.Children {
		return nil, nil
	}
	children := []*Tree{}
	for {
		e, err := rdr.Next()
		if err != nil {
			return nil, err
		}
		if e.Tag == 0 {
			break
		}
		child := entryToTreeInternal(e)
		child.Children, err = loadTreeChildren(e, rdr)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

func (n *Tree) resolveRanges(dw *dwarf.Data, staticBase uint64) error {
	var err error
	n.Ranges, err = dw.Ranges(n.Entry.(*dwarf.Entry))
	if err != nil {
		return err
	}
	for i := range n.Ranges {
		n.Ranges[i][0] += staticBase
		n.Ranges[i][1] += staticBase
	}
	n.Ranges = normalizeRanges(n.Ranges)

	for _, child := range n.Children {
		err := child.resolveRanges(dw, staticBase)
		if err != nil {
			return err
		}
		n.Ranges = fuseRanges(n.Ranges, child.Ranges)
	}
	return nil
}


func normalizeRanges(rngs [][2]uint64) [][2]uint64 {
	const (
		start = 0
		end   = 1
	)

	if len(rngs) == 0 {
		return rngs
	}

	sort.Slice(rngs, func(i, j int) bool {
		return rngs[i][start] <= rngs[j][start]
	})

	
	out := rngs[:0]
	for i := range rngs {
		if rngs[i][start] < rngs[i][end] {
			out = append(out, rngs[i])
		}
	}
	rngs = out

	
	out = rngs[:1]
	for i := 1; i < len(rngs); i++ {
		cur := rngs[i]
		if cur[start] <= out[len(out)-1][end] {
			out[len(out)-1][end] = max(cur[end], out[len(out)-1][end])
		} else {
			out = append(out, cur)
		}
	}
	return out
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}






func fuseRanges(rngs1, rngs2 [][2]uint64) [][2]uint64 {
	if rangesContains(rngs1, rngs2) {
		return rngs1
	}

	return normalizeRanges(append(rngs1, rngs2...))
}


func rangesContains(rngs1, rngs2 [][2]uint64) bool {
	i, j := 0, 0
	for {
		if i >= len(rngs1) {
			return false
		}
		if j >= len(rngs2) {
			return true
		}
		if rangeContains(rngs1[i], rngs2[j]) {
			j++
		} else {
			i++
		}
	}
}


func rangeContains(a, b [2]uint64) bool {
	return a[0] <= b[0] && a[1] >= b[1]
}

func (n *Tree) resolveAbstractEntries(rdr *dwarf.Reader) {
	n.Entry, n.Offset = LoadAbstractOrigin(n.Entry.(*dwarf.Entry), rdr)
	for _, child := range n.Children {
		child.resolveAbstractEntries(rdr)
	}
}


func (n *Tree) ContainsPC(pc uint64) bool {
	for _, rng := range n.Ranges {
		if rng[0] > pc {
			return false
		}
		if rng[0] <= pc && pc < rng[1] {
			return true
		}
	}
	return false
}

func (n *Tree) Type(dw *dwarf.Data, index int, typeCache *sync.Map) (Type, error) {
	if n.typ == nil {
		offset, ok := n.Val(dwarf.AttrType).(dwarf.Offset)
		if !ok {
			return nil, fmt.Errorf("malformed variable DIE (offset)")
		}

		var err error
		n.typ, err = ReadType(dw, index, offset, typeCache)
		if err != nil {
			return nil, err
		}
	}
	return n.typ, nil
}
