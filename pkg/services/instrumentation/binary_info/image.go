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

package binary_info

import (
	"debug/dwarf"
	"io"
	"sort"
	"sync"

	godwarf2 "github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/hashicorp/golang-lru/simplelru"
)


type Image struct {
	Path       string
	StaticBase uint64
	addr       uint64

	Index int 

	closer         io.Closer
	sepDebugCloser io.Closer

	Dwarf     *dwarf.Data
	debugAddr *godwarf2.DebugAddrSection

	TypeCache sync.Map

	compileUnits []*compileUnit 

	dwarfTreeCache *simplelru.LRU
	dwarfTreeLock  sync.Mutex

	
	
	
	
	RuntimeTypeToDIE map[uint64]runtimeTypeDIE

	loadErr      error
	debugLineStr []byte
}

type runtimeTypeDIE struct {
	Offset dwarf.Offset
	Kind   int64
}

func (i *Image) registerRuntimeTypeToDIE(entry *dwarf.Entry) {
	if off, ok := entry.Val(godwarf2.AttrGoRuntimeType).(uint64); ok {
		if _, ok := i.RuntimeTypeToDIE[off]; !ok {
			i.RuntimeTypeToDIE[off] = runtimeTypeDIE{entry.Offset, -1}
		}
	}
}

func (i *Image) GetDwarfTree(off dwarf.Offset) (*godwarf2.Tree, error) {
	i.dwarfTreeLock.Lock()
	defer i.dwarfTreeLock.Unlock()

	if r, ok := i.dwarfTreeCache.Get(off); ok {
		return r.(*godwarf2.Tree), nil
	}
	r, err := godwarf2.LoadTree(off, i.Dwarf, i.StaticBase)
	if err != nil {
		return nil, err
	}
	i.dwarfTreeCache.Add(off, r)
	return r, nil
}

func (i *Image) FindCompileUnitForOffset(off dwarf.Offset) *compileUnit {
	index := sort.Search(len(i.compileUnits), func(index int) bool {
		return i.compileUnits[index].offset >= off
	})
	if index > 0 {
		index--
	}
	return i.compileUnits[index]
}
