package callback

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"unsafe"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/locations_set"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/services/collection/go_id"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/go_runtime"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/utils"
)

var BinaryInfo *binary_info.BinaryInfo
var locationsSet *locations_set.LocationsSet
var triggerChan chan bool

func SetBinaryInfo(binaryInfoIn *binary_info.BinaryInfo) {
	BinaryInfo = binaryInfoIn
}

func SetLocationsSet(locationsSetIn *locations_set.LocationsSet) {
	locationsSet = locationsSetIn
}

func SetTriggerChan(triggerChanIn chan bool) {
	triggerChan = triggerChanIn
}

type BreakpointInfo struct {
	Stacktrace []collection.Stackframe
	regs       registers.Registers
}

func collectStacktrace(regs *registers.OnStackRegisters, g go_runtime.GPtr, maxStacktrace int) []collection.Stackframe {
	if maxStacktrace == 0 {
		maxStacktrace = 1 
	}

	pcs := make([]uintptr, maxStacktrace)
	
	frameCount := go_runtime.Callers(uintptr(regs.PC()), uintptr(regs.SP()), g, pcs)
	pcs = pcs[:frameCount]
	stacktrace := make([]collection.Stackframe, 0, frameCount)
	frames := runtime.CallersFrames(pcs)

	more := true
	frame := runtime.Frame{}
	for i := 0; i < frameCount; i++ {
		if !more {
			
			logger.Logger().Warningf("Expected more frames but more is false: %d/%d\n", i, frameCount)
			break
		}

		frame, more = frames.Next()
		
		
		
		
		
		stacktrace = append(stacktrace,
			collection.Stackframe{
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Func.Name(),
			})
	}

	return stacktrace
}

func printBytesAt(sp, count uint64, prefixes []string) uint64 {
	for i := uint64(0); i < count; i++ {
		//goland:noinspection GoVetUnsafePointer
		stackValue := *(*uint64)(unsafe.Pointer(uintptr(sp)))
		fmt.Printf("0x%016x:\t0x%016x (%s) \n", sp, stackValue, prefixes[i])
		sp = sp - 8
	}
	return sp
}

func printStack(stackRegs registers.OnStackRegisters) {
	stackPtr := stackRegs.SP()

	fmt.Printf("BP: 0x%016x\n", stackRegs.BP())
	fmt.Printf("SP: 0x%016x\n", stackRegs.SP())

	fmt.Println("Native:")
	stackPtr = printBytesAt(stackPtr, 4, []string{"idk", "flags", "rdi", "rdx"})
	stackPtr -= 240
	stackPtr = printBytesAt(stackPtr, 11, []string{"rbx", "rax", "rcx", "rsi", "r8", "r9", "r10", "r11", "rbp", "rdi", "retval"})
	fmt.Println("Golang:")
	if runtime.Version() == "go1.16" {
		stackPtr = printBytesAt(stackPtr, 1, []string{"idk", "idk"})
	} else {
		stackPtr = printBytesAt(stackPtr, 2, []string{"idk", "idk"})
	}
	stackPtr = printBytesAt(stackPtr, 12, []string{"rsp", "rbp", "tls", "rip", "r11", "r10", "r9", "r8", "rsi", "rcx", "rax", "rbx"})
	stackPtr = printBytesAt(stackPtr, 2, []string{"rdx", "rdi"})
}

//go:linkname systemstack runtime.systemstack
func systemstack(func())

//go:nosplit
func Callback() {
	context := getContext()
	g := go_runtime.Getg()

	goCollect(context, g)
}

//go:nosplit
func goCollect(context uintptr, g go_runtime.GPtr) {
	var waitChan chan struct{}

	systemstack(func() {
		triggerChan <- true
		waitChan = make(chan struct{})

		
		go func() {
			defer func() {
				waitChan <- struct{}{}
				triggerChan <- false
				if v := recover(); v != nil {
					if utils.OnPanicFunc != nil {
						utils.OnPanicFunc(rookoutErrors.NewRookPanicInGoroutine(v))
					}

					return
				}
			}()
			debug.SetPanicOnFault(true)

			collectBreakpoint(context, g)
		}()
	})

	<-waitChan
}

func collectBreakpoint(context uintptr, g go_runtime.GPtr) {
	regs := registers.NewOnStackRegisters(context)
	bpInstance, ok := locationsSet.FindBreakpointByAddr(regs.PC())
	if !ok {
		file, line, function := BinaryInfo.PCToLine(regs.PC())
		var functionName string
		if function != nil {
			functionName = function.Name
		}

		logger.Logger().Errorf("Breakpoint triggered on unknown address, 0x%x (%s:%d - %s)", regs.PC(), file, line, functionName)
		return
	}

	goid := go_id.GetGoID(g)
	stacktrace := collectStacktrace(regs, g, bpInstance.Breakpoint.Stacktrace)
	
	stacktrace[0].Line = bpInstance.Breakpoint.Line

	bpInfo := &BreakpointInfo{
		Stacktrace: stacktrace,
		regs:       regs,
	}
	reportBreakpoint(bpInstance, bpInfo, goid)
}

func reportBreakpoint(bpInstance *augs.BreakpointInstance, bpInfo *BreakpointInfo, goid int) {
	bp := bpInstance.Breakpoint
	locations, exists := locationsSet.FindLocationsByBreakpointName(bp.Name)
	if !exists {
		logger.Logger().Errorf("Breakpoint %s (on %s:%d - 0x%x) triggered but the breakpoint doesn't exist.", bp.Name, bp.File, bp.Line, bpInfo.regs.PC())
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(locations))
	for i := range locations {
		utils.CreateGoroutine(func(i int) func() {
			return func() { 
				defer wg.Done()

				collectionService, err := collection.NewCollectionService(bpInfo.regs, BinaryInfo.PointerSize, bpInfo.Stacktrace, bpInstance.VariableLocators, goid)
				if err != nil {
					logger.Logger().WithError(err).Errorf("failed to report breakpoint info")
				}

				locations[i].GetAug().Execute(collectionService)
			}
		}(i))
	}
	wg.Wait()
}
