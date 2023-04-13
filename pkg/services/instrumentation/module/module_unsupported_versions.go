//go:build !go1.15 || go1.21
// +build !go1.15 go1.21

package module

import (
	_ "unsafe"
)



type functab struct {
}

type _func struct {
}

type pcHeader struct {
}

type moduledata struct {
	text          uintptr
	types, etypes uintptr
	next          *moduledata
}

func getPCTab(m *moduledata) []byte {
	return nil
}

func (f *FuncInfo) getEntry() uintptr {
	return 0
}


func findFuncOffsetInModule(pc uintptr, datap *moduledata) (uintptr, bool) {
	return 0, false
}

func (md *moduledata) GetTypeMap() map[TypeOff]uintptr {
	return nil
}
