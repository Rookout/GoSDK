package variable

import (
	"sync"

	"github.com/Rookout/GoSDK/pkg/services/collection/memory"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)

type variablesCacheKey struct {
	addr  uint64
	typ   string
	memID string
}
type VariablesCache struct {
	m    map[variablesCacheKey]*internalVariable
	lock sync.RWMutex
}

func NewVariablesCache() *VariablesCache {
	return &VariablesCache{m: make(map[variablesCacheKey]*internalVariable)}
}

func (v *VariablesCache) get(addr uint64, typ godwarf.Type, mem memory.MemoryReader) (*internalVariable, bool) {
	v.lock.RLock()
	defer v.lock.RUnlock()
	defer recover()

	iv, ok := v.m[variablesCacheKey{addr, typ.String(), mem.ID()}]
	return iv, ok
}

func (v *VariablesCache) set(iv *internalVariable) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.m[variablesCacheKey{iv.Addr, iv.DwarfType.String(), iv.Mem.ID()}] = iv
}


func (v *VariablesCache) Len() int {
	return len(v.m)
}
