package utils

import (
	"context"
	"runtime/debug"
	_ "unsafe"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type OnPanicFuncType func(error)

var OnPanicFunc OnPanicFuncType

type panicInfo struct {
	didPanic bool
}

const allowPanic = false

func SetOnPanicFunc(onPanic OnPanicFuncType) {
	OnPanicFunc = onPanic
}

func createHandlePanicFunc(info *panicInfo) func() {
	return func() {
		if allowPanic {
			return
		}

		if v := recover(); v != nil {
			if OnPanicFunc != nil {
				OnPanicFunc(rookoutErrors.NewRookPanicInGoroutine(v))
			}

			if info != nil {
				info.didPanic = true
			}

			return
		}

		if info != nil {
			info.didPanic = false
		}
	}
}

func CreateGoroutine(f func()) {
	handlePanic := createHandlePanicFunc(nil)

	
	go func() {
		defer handlePanic()
		debug.SetPanicOnFault(true)

		f()
	}()
}



func CreateRetryingGoroutine(ctx context.Context, f func()) {
	CreateGoroutine(func() {
		for {
			info := &panicInfo{}
			handlePanic := createHandlePanicFunc(info)

			func() {
				defer handlePanic()
				f()
			}()

			select {
			case <-ctx.Done():
				return
			default:
				if !info.didPanic {
					return
				}
			}
		}
	})
}


func CreateBlockingGoroutine(f func()) {
	waitChan := make(chan interface{})

	CreateGoroutine(func() {
		defer func() {
			waitChan <- nil
		}()

		f()
	})

	<-waitChan
}
