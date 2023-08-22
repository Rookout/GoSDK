package variable

import (
	"reflect"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/logger"
)

type VariablesPool chan *internalVariable

var variablesPool = make(VariablesPool, 102400)

var internalVariableSize = func() int {
	return int(unsafe.Sizeof(internalVariable{}))
}()

func (v *VariablesPool) get() *internalVariable {
	var i *internalVariable
	select {
	case i = <-variablesPool:
	default:
		i = &internalVariable{}
	}
	i.inPool = false
	return i
}

func clear(i *internalVariable) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger().Errorf("Failed to clear variable, error: %v", r)
		}
	}()

	a := unsafe.Pointer(reflect.ValueOf(i).Pointer())
	for i := 0; i < internalVariableSize; i++ {
		b := (*byte)(a)
		*b = 0
		a = unsafe.Pointer(uintptr(a) + unsafe.Sizeof(*b))
	}
}

func (v *VariablesPool) set(i *internalVariable) {
	clear(i)
	i.inPool = true

	select {
	case variablesPool <- i:
	default:
	}
}
