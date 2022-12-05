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
	"github.com/Rookout/GoSDK/pkg/services/collection"
	"github.com/Rookout/GoSDK/pkg/services/collection/go_id"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/utils"
)

var BinaryInfo *binary_info.BinaryInfo
var locationsSet *locations_set.LocationsSet
var originalGCPercent int
var collectionCounter int
var setGCPercentLock sync.Mutex

func SetBinaryInfo(binaryInfoIn *binary_info.BinaryInfo) {
	BinaryInfo = binaryInfoIn
}

func SetLocationsSet(locationsSetIn *locations_set.LocationsSet) {
	locationsSet = locationsSetIn
}

type BreakpointInfo struct {
	Stacktrace []collection.Stackframe
	regs       registers.Registers
}


func collectStacktrace(bp *augs.Breakpoint, pcs []uintptr) (*BreakpointInfo, error) {
	bpi := &BreakpointInfo{}
	if bp.Stacktrace > 0 {
		frames := runtime.CallersFrames(pcs)
		frameCount := len(pcs)
		if len(pcs) > bp.Stacktrace-1 {
			frameCount = bp.Stacktrace - 1
		}

		
		bpi.Stacktrace = append(bpi.Stacktrace, collection.Stackframe{
			File:     bp.File,
			Line:     bp.Line,
			Function: &collection.Function{Name: bp.FunctionName},
		})

		more := true
		frame := runtime.Frame{}
		for i := 0; i < frameCount; i++ {
			if !more {
				
				logger.Logger().Warningf("Expected more frames but more is false: %d/%d\n", i, frameCount)
				break
			}

			frame, more = frames.Next()
			
			
			
			
			
			function := &collection.Function{
				Name:      frame.Func.Name(),
				Type:      0,
				Value:     uint64(frame.Entry),
				GoType:    0,
				Optimized: false,
			}
			bpi.Stacktrace = append(bpi.Stacktrace,
				collection.Stackframe{
					File:     frame.File,
					Line:     frame.Line,
					Function: function,
				})
		}
	}

	return bpi, nil
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

const MaxStacktrace = 4

func callback(stackRegs registers.OnStackRegisters) {
	setGCPercentLock.Lock()
	if collectionCounter == 0 {
		originalGCPercent = debug.SetGCPercent(-1)
	}
	collectionCounter++
	setGCPercentLock.Unlock()

	defer func() {
		setGCPercentLock.Lock()
		defer setGCPercentLock.Unlock()

		collectionCounter--
		if collectionCounter == 0 {
			debug.SetGCPercent(originalGCPercent)
		}
	}()
	
	
	goid := go_id.CurrentGoID()
	pcs := make([]uintptr, MaxStacktrace-1)
	
	frameCount := runtime.Callers(4, pcs)
	utils.CreateBlockingGoroutine(func() {
		pcs = pcs[:frameCount]
		collectBreakpoint(stackRegs, pcs, goid)
	})
}

func collectBreakpoint(regs registers.OnStackRegisters, pcs []uintptr, goid int) {
	bp, bpInstance, ok := locationsSet.FindBreakpointByAddr(regs.PC())
	if !ok {
		file, line, function := BinaryInfo.PCToLine(regs.PC())
		var functionName string
		if function != nil {
			functionName = function.Name
		}

		logger.Logger().Errorf("Breakpoint triggered on unknown address, 0x%x (%s:%d - %s)", regs.PC(), file, line, functionName)
		return
	}

	bpInfo, err := collectStacktrace(bp, pcs)
	if err != nil {
		logger.Logger().WithError(err).Errorf("failed to collect breakpoint info")
		return
	}

	bpInfo.regs = regs
	reportBreakpoint(bp, bpInstance, bpInfo, goid)
}

func reportBreakpoint(bp *augs.Breakpoint, bpInstance *augs.BreakpointInstance, bpInfo *BreakpointInfo, goid int) {
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
