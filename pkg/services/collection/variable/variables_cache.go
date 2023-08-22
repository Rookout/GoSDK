package variable

import (
	"github.com/Rookout/GoSDK/pkg/services/collection/memory"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)

type variablesCacheKey struct {
	addr  uint64
	typ   string
	memID string
}
type VariablesCache map[variablesCacheKey]*internalVariable

func (v VariablesCache) get(addr uint64, typ godwarf.Type, mem memory.MemoryReader) (*internalVariable, bool) {
	defer recover()

	iv, ok := v[variablesCacheKey{addr, typ.String(), mem.ID()}]
	return iv, ok
}

func (v VariablesCache) set(iv *internalVariable) {
	v[variablesCacheKey{iv.Addr, iv.DwarfType.String(), iv.Mem.ID()}] = iv
}
