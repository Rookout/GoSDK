package binary_info

import (
	"debug/dwarf"
	godwarf2 "github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/hashicorp/golang-lru/simplelru"
	"io"
	"sort"
	"sync"
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

	loadErrMu    sync.Mutex
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
			i.RuntimeTypeToDIE[off+i.StaticBase] = runtimeTypeDIE{entry.Offset, -1}
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
