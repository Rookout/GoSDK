package utils

import (
	"context"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"runtime/debug"
)

type OnPanicFuncType func(error)

var onPanicFunc OnPanicFuncType

type panicInfo struct {
	didPanic bool
}

const allowPanic = false

func SetOnPanicFunc(onPanic OnPanicFuncType) {
	onPanicFunc = onPanic
}

func createHandlePanicFunc(info *panicInfo) func() {
	return func() {
		if allowPanic {
			return
		}

		if v := recover(); v != nil {
			if onPanicFunc != nil {
				onPanicFunc(rookoutErrors.NewRookPanicInGoroutine(v))
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
