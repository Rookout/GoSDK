package rookoutErrors

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-errors/errors"
)

type RookoutError interface {
	error

	StackFrames() []errors.StackFrame
	Stack() []byte

	GetType() string

	GetArguments() map[interface{}]interface{}
	AddArgument(key string, value interface{})
}

type RookoutErrorImpl struct {
	ExternalError error
	Type          string `json:"Type"`
	Arguments     map[string]interface{}
}

func (r *RookoutErrorImpl) Error() string {
	errorString := r.Type

	if nil != r.ExternalError {
		errorString += ": " + r.ExternalError.Error()
	}

	if len(r.Arguments) != 0 {
		errorString = fmt.Sprintf(errorString+" | %v", r.Arguments)
	}

	if nil != r.ExternalError {
		errorString += ": " + r.ExternalError.Error()
	}

	return errorString
}

func (r *RookoutErrorImpl) GetType() string {
	return r.Type
}

func (r *RookoutErrorImpl) GetArguments() map[interface{}]interface{} {
	outputMap := make(map[interface{}]interface{})
	for key, value := range r.Arguments {
		outputMap[key] = value
	}
	return outputMap
}

func (r *RookoutErrorImpl) AddArgument(key string, value interface{}) {
	r.Arguments[key] = value
}

func (r *RookoutErrorImpl) StackFrames() []errors.StackFrame {
	switch e := r.ExternalError.(type) {
	case *errors.Error:
		return e.StackFrames()
	case *RookoutErrorImpl:
		return e.StackFrames()
	default:
		return errors.New(e).StackFrames()
	}
}

func (r *RookoutErrorImpl) Stack() []byte {
	switch e := r.ExternalError.(type) {
	case *errors.Error:
		return e.Stack()
	case *RookoutErrorImpl:
		return e.Stack()
	default:
		return errors.New(e).Stack()
	}
}

func newRookoutError(errorType string, description string, externalError error, arguments map[string]interface{}) *RookoutErrorImpl {
	if _, ok := externalError.(*errors.Error); !ok {
		if externalError != nil {
			externalError = errors.Wrap(externalError.Error(), 2)
		} else {
			externalError = errors.Wrap(description, 2)
		}
	}

	return &RookoutErrorImpl{
		Type:          errorType,
		ExternalError: externalError,
		Arguments:     arguments,
	}
}

func NewRookoutError(errorType string, description string, externalError error, arguments map[string]interface{}) RookoutError {
	return newRookoutError(errorType, description, externalError, arguments)
}

func NewContextEnded(externalErr error) RookoutError {
	return newRookoutError(
		"ContextEnded",
		"",
		externalErr,
		map[string]interface{}{})
}

func NewRookMissingToken() RookoutError {
	return newRookoutError("RookMissingToken", "No Rookout token was supplied. Make sure to pass the Rookout Token when starting the rook", nil, map[string]interface{}{})
}

func NewRookInvalidOptions(desc string) RookoutError {
	return newRookoutError("RookInvalidOptions", desc, nil, map[string]interface{}{})
}

func NewRuntimeError(description string) RookoutError {
	return newRookoutError("RuntimeError", description, nil, map[string]interface{}{})
}

func NewObjectHasNoSizeException(obj interface{}) RookoutError {
	return newRookoutError("ObjectHasNoSize", "Cannot get the size of object (probably isn't an array/list)", nil, map[string]interface{}{
		"obj": obj,
	})
}

func NewRookMethodNotFound(name string) RookoutError {
	return newRookoutError("RookMethodNotFound", name, nil, map[string]interface{}{})
}

func NewNotImplemented() RookoutError {
	return newRookoutError("NotImplementedException", "", nil, map[string]interface{}{})
}

func NewAgentKeyNotFoundException(name string, key interface{}, externalErr error) RookoutError {
	return newRookoutError(
		"RookKeyNotFoundException",
		"Failed to get key",
		externalErr,
		map[string]interface{}{
			"name": name,
			"key":  key,
		})
}

func NewInvalidInterfaceVariable(key interface{}) RookoutError {
	return newRookoutError(
		"InvalidInterfaceVariable",
		"Tried extracting inner variable from interface but got another interface",
		nil,
		map[string]interface{}{
			"key": key,
		})
}

func NewRookAttributeNotFoundException(name string) RookoutError {
	return newRookoutError(
		"RookAttributeNotFound",
		"Failed to get attribute",
		nil,
		map[string]interface{}{
			"attribute": name,
		})
}

func NewRookInvalidArithmeticPathException(configuration interface{},
	externalError error) RookoutError {
	return newRookoutError(
		"RookInvalidArithmeticPath",
		"Invalid arithmetic path configuration",
		externalError,
		map[string]interface{}{
			"configuration": configuration,
		})
}

func NewArithmeticPathException(externalError error) RookoutError {
	return newRookoutError(
		"ArithmeticPathException",
		"Invalid arithmetic path procedure",
		externalError,
		map[string]interface{}{})
}

func NewRookOperationReadOnlyException(operationType string) RookoutError {
	return newRookoutError(
		"RookOperationReadOnly",
		"Operation does not support write",
		nil,
		map[string]interface{}{
			"operation": operationType,
		})
}

func NewJsonMarshallingException(jsonData interface{}, externalError error) RookoutError {
	return newRookoutError(
		"JsonMarshallingException",
		"",
		externalError,
		map[string]interface{}{
			"jsonData": jsonData,
		})
}

func NewNilProtobufNamespaceException() RookoutError {
	return newRookoutError(
		"NilProtobufNamespaceException",
		"",
		nil,
		map[string]interface{}{})
}

func NewRookAugInvalidKey(key string, aug interface{}) RookoutError {
	return newRookoutError(
		"RookAugInvalidKey",
		"Failed to get key from configuration",
		nil,
		map[string]interface{}{
			"key":           key,
			"configuration": aug,
		})
}

func NewBadTypeException(description string, obj interface{}) RookoutError {
	return newRookoutError(
		"BadTypeException",
		description,
		nil,
		map[string]interface{}{
			"obj": obj,
		})
}

func NewBadFunctionNameException(functionName string) RookoutError {
	return newRookoutError(
		"BadFunctionNameException",
		"",
		nil,
		map[string]interface{}{
			"functionName": functionName,
		})
}

func NewInvalidProcMapsStartAddress(line string, startAddress string, externalErr error) RookoutError {
	return newRookoutError(
		"InvalidProcMapsStartEndAddress",
		"Expected start address in proc maps line to be uint",
		externalErr,
		map[string]interface{}{
			"line":         line,
			"startAddress": startAddress,
		},
	)
}

func NewInvalidProcMapsEndAddress(line string, endAddress string, externalErr error) RookoutError {
	return newRookoutError(
		"InvalidProcMapsEndAddress",
		"Expected end address in proc maps line to be uint",
		externalErr,
		map[string]interface{}{
			"line":       line,
			"endAddress": endAddress,
		},
	)
}

func NewInvalidProcMapsAddresses(line string, addresses string) RookoutError {
	return newRookoutError(
		"InvalidProcMapsAddresses",
		"Expected startAddress-endAddress in proc maps line",
		nil,
		map[string]interface{}{
			"line":      line,
			"addresses": addresses,
		},
	)
}

func NewInvalidProcMapsLine(line string) RookoutError {
	return newRookoutError(
		"InvalidProcMapsLine",
		"Expected at least 5 fields in proc maps line",
		nil,
		map[string]interface{}{
			"line": line,
		},
	)
}

func NewFailedToOpenProcMapsFile(externalErr error) RookoutError {
	return newRookoutError(
		"FailedToOpenProcMapsFile",
		"Unable to open /proc/self/maps",
		externalErr,
		map[string]interface{}{},
	)
}

func NewFailedToWriteBytes(errno int) *RookoutErrorImpl {
	return newRookoutError(
		"FailedToWriteBytes",
		"Failed to write hook bytes",
		nil,
		map[string]interface{}{
			"errno": errno,
		},
	)
}

func NewFailedToCollectGoroutinesInfo(numGoroutines int) RookoutError {
	return newRookoutError(
		"FailedToCollectGoroutinesInfo",
		"Failed to collect all goroutines info",
		nil,
		map[string]interface{}{
			"numGoroutines": numGoroutines,
		},
	)
}

func NewUnsafeToInstallHook(reason string) RookoutError {
	return newRookoutError(
		"UnsafeToInstallHook",
		"Detected it's unsafe to install hook at this time",
		nil,
		map[string]interface{}{
			"reason": reason,
		},
	)
}

func NewFailedToGetStateEntryAddr(functionEntry uint64, functionEnd uint64, stateID int, externalErr error) RookoutError {
	return newRookoutError(
		"FailedToGetStateEntryAddr",
		"Unable to get state entry addr",
		externalErr,
		map[string]interface{}{
			"functionEntry": functionEntry,
			"functionEnd":   functionEnd,
			"stateID":       stateID,
		})
}

func NewInvalidBranchDest(src uintptr, dst uintptr) RookoutError {
	return newRookoutError(
		"InvalidBranchDest",
		"Tried to encode an invalid branch instruction - relative distance isn't dividable by 4",
		nil,
		map[string]interface{}{
			"src": src,
			"dst": dst,
		})
}

func NewBranchDestTooFar(src uintptr, dst uintptr) RookoutError {
	return newRookoutError(
		"BranchDestTooFar",
		"Tried to encode an invalid branch instruction - relative distance is too long to be encoded into 26 bit immediate",
		nil,
		map[string]interface{}{
			"src": src,
			"dst": dst,
		})
}

func NewFailedToGetHookAddress(errorMsg string) RookoutError {
	return newRookoutError(
		"FailedToGetHookAddress",
		"Failed to get hook address from native",
		nil,
		map[string]interface{}{
			"errorMsg": errorMsg,
		})
}

func NewFailedToRetrieveStackTrace() RookoutError {
	return newRookoutError(
		"FailedToRetrieveStackTrace",
		"",
		nil,
		map[string]interface{}{})
}

func NewFailedToRetrieveFrameLocals(externalErr error) RookoutError {
	return newRookoutError(
		"FailedToRetrieveFrameLocals",
		"",
		externalErr,
		map[string]interface{}{})
}

func NewFailedToInitNative(nativeError string) RookoutError {
	return newRookoutError(
		"FailedToInitNative",
		"",
		nil,
		map[string]interface{}{
			"nativeError": nativeError,
		})
}

func NewFailedToDestroyNative(nativeError error) RookoutError {
	return newRookoutError(
		"FailedToDestroyNative",
		"",
		nil,
		map[string]interface{}{
			"nativeError": nativeError,
		})
}

func NewBadVariantType(description string, variant interface{}) RookoutError {
	return newRookoutError(
		"BadVariantType",
		description,
		nil,
		map[string]interface{}{"variant": variant})
}

func NewFailedToCreateID(err error) RookoutError {
	return newRookoutError(
		"RookFailedToCreateID",
		"",
		err,
		map[string]interface{}{})
}

func NewInvalidTokenError() RookoutError {
	return newRookoutError(
		"InvalidTokenError",
		"The Rookout token supplied is not valid; please check the token and try again",
		nil,
		map[string]interface{}{})
}

func NewWebSocketError() RookoutError {
	return newRookoutError(
		"WebSocketError",
		"Received HTTP status 400 from the controller, please make sure WebSocket is enabled on the load balancer.",
		nil,
		map[string]interface{}{})
}

func NewInvalidLabelError(label string) RookoutError {
	return newRookoutError(
		"InvalidLabelError",
		"Invalid label: must not start with the '$' character",
		nil,
		map[string]interface{}{"label": label})
}

func NewRookRuntimeVersionNotSupported(currentVersion string) RookoutError {
	return newRookoutError(
		"RookRuntimeVersionNotSupported",
		"This runtime version is not supported by Rookout.",
		nil,
		map[string]interface{}{"currentVersion": currentVersion})
}

func NewRookObjectNameMissing(configuration interface{}) RookoutError {
	return newRookoutError(
		"RookObjectNameMissing",
		"Failed to find object name",
		nil,
		map[string]interface{}{
			"configuration": configuration,
		})
}

func NewRookUnsupportedLocation(name string) RookoutError {
	return newRookoutError(
		"RookUnsupportedLocation",
		"Unsupported aug location was specified",
		nil,
		map[string]interface{}{
			"location": name,
		},
	)
}

func NewRookInvalidActionConfiguration(configuration interface{}) RookoutError {
	return newRookoutError(
		"RookInvalidActionConfiguration",
		"Failed to parse action configuration",
		nil,
		map[string]interface{}{
			"configuration": configuration,
		})
}

func NewRookInvalidOperationConfiguration(configuration interface{}) RookoutError {
	return newRookoutError(
		"RookInvalidOperationConfiguration",
		"Failed to parse operation configuration",
		nil,
		map[string]interface{}{
			"configuration": configuration,
		})
}

func NewRookConnectToControllerTimeout() RookoutError {
	return newRookoutError(
		"RookConnectToControllerTimeout",
		"Failed to connect to the controller - will continue attempting in the background",
		nil,
		map[string]interface{}{})
}

func NewUnknownError(recovered interface{}) RookoutError {
	err, _ := recovered.(error)

	return newRookoutError(
		"Unknown",
		"Unexpected error",
		err,
		map[string]interface{}{
			"recovered": recovered,
		})
}

func NewRookRuleAugRateLimited() RookoutError {
	return newRookoutError(
		"RookRuleAugRateLimited",
		"Breakpoint was disabled due to rate-limiting. \nFor more information: https://docs.rookout.com/docs/breakpoints-tasks.html#rate-limiting",
		nil,
		map[string]interface{}{})
}

func NewRookRuleGlobalRateLimited() RookoutError {
	return newRookoutError(
		"RookRuleGlobalRateLimited",
		"Breakpoint was disabled due to global rate-limiting. \nFor more information: https://docs.rookout.com/docs/breakpoints-tasks.html#rate-limiting",
		nil,
		map[string]interface{}{})
}

var UsingGlobalRateLimiter = false

func NewRookRuleRateLimited() RookoutError {
	if UsingGlobalRateLimiter {
		return NewRookRuleGlobalRateLimited()
	}
	return NewRookRuleAugRateLimited()
}

func NewRookInvalidRateLimitConfiguration(config string) RookoutError {
	return newRookoutError(
		"RookInvalidRateLimitConfiguration",
		fmt.Sprintf("Got an invalid value for the rate limit. (got %s) expected XX/YY or XX\\YY, where XX < YY", config),
		nil,
		map[string]interface{}{
			"config": config,
		})
}

func NewRookRuleMaxExecutionTimeReached() RookoutError {
	return newRookoutError(
		"RookRuleMaxExecutionTimeReached",
		"Breakpoint was disabled because it has reached its maximum execution time",
		nil,
		map[string]interface{}{})
}

func NewRookInvalidMethodArguments(method, arguments string) RookoutError {
	return newRookoutError(
		"RookInvalidMethodArguments",
		"Bad method arguments",
		nil,
		map[string]interface{}{
			"method":    method,
			"arguments": arguments,
		})
}

func NewFileNotFound(filename string) RookoutError {
	return newRookoutError(
		"FileNotFound",
		fmt.Sprintf("No such file found %s", filename),
		nil,
		map[string]interface{}{
			"filename": filename,
		})
}

func NewLineNotFound(filename string, lineno int) RookoutError {
	return newRookoutError(
		"RookLineNotFound",
		fmt.Sprintf("Can't break on line %d in file %s", lineno, filename),
		nil,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewMultipleFilesFound(filename string) RookoutError {
	return newRookoutError(
		"MultipleFilesFound",
		fmt.Sprintf("Found multiple files matching %s, use more specific file path", filename),
		nil,
		map[string]interface{}{
			"filename": filename,
		})
}

func NewFailedToAddBreakpoint(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"FailedToAddBreakpoint",
		"Unable to add a breakpoint at the given address",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewAllTrampolineAddressesInUse() RookoutError {
	return newRookoutError(
		"AllTrampolineAddressesInUse",
		"Can't add another breakpoint since all trampolines are in use",
		nil,
		map[string]interface{}{})
}

func NewFailedToRemoveBreakpoint(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"FailedToRemoveBreakpoint",
		"Unable to remove the breakpoint at the given address",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewFailedToEraseAllBreakpointInstances() RookoutError {
	return newRookoutError(
		"FailedToEraseAllBreakpointInstances",
		"Could not remove at least one breakpoint instance",
		nil,
		map[string]interface{}{})
}

func NewFailedToGetExecutable(err error) RookoutError {
	return newRookoutError(
		"FailedToGetExecutable",
		"",
		err,
		map[string]interface{}{})
}

func NewFailedToLoadBinaryInfo(err error) RookoutError {
	return newRookoutError(
		"FailedToLoadBinaryInfo",
		"",
		err,
		map[string]interface{}{})
}

func NewFailedToPatchModule(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"NewFailedToPatchModule",
		"",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewFailedToApplyBreakpointState(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"NewFailedToApplyBreakpointState",
		"",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}
func NewFailedToGetVariableLocators(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"NewFailedToGetVariableLocators",
		"",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewFailedToGetUnpatchedAddressMapping(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"FailedToGetUnpatchedAddressMapping",
		"",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewFailedToGetAddressMapping(filename string, lineno int, err error) RookoutError {
	return newRookoutError(
		"FailedToGetAddressMapping",
		"",
		err,
		map[string]interface{}{
			"filename": filename,
			"lineno":   lineno,
		})
}

func NewFailedToStartCopyingFunction(err error) RookoutError {
	return newRookoutError(
		"FailedToStartCopyingFunction",
		"Unable to start copying original function",
		err,
		map[string]interface{}{})
}

func NewCompiledWithoutCGO() RookoutError {
	return newRookoutError("CompiledWithoutCGO", "Your project was built with CGO_ENABLED disabled", nil, map[string]interface{}{})
}

func NewFailedToGetDWARFTree(err error) RookoutError {
	return newRookoutError(
		"FailedToGetDWARFTree",
		"Unable to load DWARF tree",
		err,
		map[string]interface{}{})
}

func NewFlushTimedOut() RookoutError {
	return newRookoutError("FlushTimedOut", "Timed out during flush", nil, map[string]interface{}{})
}

func NewRookOutputQueueFull() RookoutError {
	return newRookoutError("RookOutputQueueFull",
		"Breakpoint triggered but output queue is full. Data collection will be disabled until the queue has emptied.",
		nil,
		map[string]interface{}{})
}

func NewInvalidDwarfRegister(dwarfReg uint64) RookoutError {
	return newRookoutError("InvalidDwarfRegister",
		"Tracked invalid dwarf register while locating variable",
		nil,
		map[string]interface{}{
			"dwarfReg": dwarfReg,
		})
}

func NewFailedToLocate(variableName string, externalErr error) RookoutError {
	return newRookoutError("FailedToLocate",
		"Failed to locate variable",
		externalErr,
		map[string]interface{}{
			"variableName": variableName,
		})
}

func NewFailedToAlignFunc(funcAddress, pclntableAddress, funcOffset uintptr) RookoutError {
	return newRookoutError(
		"FailedToAlignFunc",
		"Tried to align _func in moduledata pclntable but failed",
		nil,
		map[string]interface{}{
			"funcAddress":      funcAddress,
			"pclntableAddress": pclntableAddress,
			"funcOffset":       funcOffset,
		},
	)
}

func NewRookMessageSizeExceeded(messageSize int, maxMessageSize int) RookoutError {
	return newRookoutError("RookMessageSizeExceeded",
		fmt.Sprintf("Message size of %d exceeds max size limit of %d. "+
			"Change the depth of collection or change the default by setting ROOKOUT_MAX_MESSAGE_SIZE as environment variable or system property", messageSize, maxMessageSize),
		nil,
		map[string]interface{}{
			"messageSize":    messageSize,
			"maxMessageSize": maxMessageSize,
		})
}

func NewUnwrappedFuncNotFound(funcName string) RookoutError {
	return newRookoutError(
		"UnwrappedFuncNotFound",
		"Could not find unwrapped address of go assembly function",
		nil,
		map[string]interface{}{
			"funcName": funcName,
		})
}

func NewUnsupportedPlatform() RookoutError {
	var desc string
	if runtime.GOOS == "windows" {
		desc = "Your project was built for an unsupported platform - Windows"
	} else if runtime.GOARCH != "amd64" {
		desc = "Your project was built for an unsupported platform architecture - " + runtime.GOARCH + "-" + runtime.GOOS
	} else {
		desc = "You're building without CGO enabled, which is not supported"
	}
	return newRookoutError(
		"UnsupportedPlatform",
		desc,
		nil,
		map[string]interface{}{})
}

func NewFailedToExecuteBreakpoint(failedCount uint64) RookoutError {
	return newRookoutError(
		"FailedToExecuteBreakpoint",
		fmt.Sprintf("%d breakpoint executions failed because registers backup buffer was full", failedCount),
		nil,
		map[string]interface{}{
			"failedCount": failedCount,
		},
	)
}

func NewReadBuildFlagsError() RookoutError {
	return newRookoutError(
		"ReadBuildFlagsError",
		"Couldn't read the build flags. Verify the application was built with go build",
		nil,
		map[string]interface{}{})
}

func NewValidateBuildFlagsError(err error) RookoutError {
	return newRookoutError(
		"ValidateBuildFlagsError",
		"The application wasn't built with -gcflags all=-dwarflocationlists=true or it was built with either -ldflags -s or -w",
		err,
		map[string]interface{}{})
}

func NewMprotectFailed(address uintptr, size int, permissions int, err string) RookoutError {
	return newRookoutError(
		"MprotectFailed",
		"Tried to change permissions of memory area but failed",
		nil,
		map[string]interface{}{
			"address":     address,
			"size":        size,
			"permissions": permissions,
			"err":         err,
		})
}

func NewFailedToGetCurrentMemoryProtection(address uint64, size uint64) RookoutError {
	return newRookoutError(
		"FailedToGetCurrentMemoryProtection",
		"Failed to get current memory protection",
		nil,
		map[string]interface{}{
			"address": address,
			"size":    size,
		})
}

func NewDifferentNPCData(origNPCData uint32, newNPCData uint32) RookoutError {
	return newRookoutError(
		"DifferenceNPCData",
		"New module doesn't have the same number of PCData tables as original module",
		nil,
		map[string]interface{}{
			"origNPCData": origNPCData,
			"newNPCData":  newNPCData,
		})
}

func NewPCDataVerificationFailed(table uint32, origValue int32, origPC uintptr, newValue int32, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCDataVerificationFailed",
		"New module has a different value in pcdata table than the original module",
		nil,
		map[string]interface{}{
			"table":     table,
			"origValue": origValue,
			"origPC":    origPC,
			"newValue":  newValue,
			"newPC":     newPC,
		})
}

func NewPCDataAsyncUnsafePointVerificationFailed(newValue int32, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCDataAsyncUnsafePointVerificationFailed",
		"New module has a different value than the unsafe point for PCs within the patched code",
		nil,
		map[string]interface{}{
			"newValue": newValue,
			"newPC":    newPC,
		})
}

func NewPCSPInPatchedVerificationFailed(origValue int32, origPC uintptr, expectedNewValue, newValue int32, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCSPInPatchedVerificationFailed",
		"New module has a different value in pcsp table than the expected generated values",
		nil,
		map[string]interface{}{
			"origValue":        origValue,
			"origPC":           origPC,
			"expectedNewValue": expectedNewValue,
			"newValue":         newValue,
			"newPC":            newPC,
		})
}

func NewPCSPVerificationFailed(origValue int32, origPC uintptr, newValue int32, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCSPVerificationFailed",
		"New module has a different value in pcsp table than the original module",
		nil,
		map[string]interface{}{
			"origValue": origValue,
			"origPC":    origPC,
			"newValue":  newValue,
			"newPC":     newPC,
		})
}

func NewPCSPVerificationFailedMissingEntry(origValue int32, origPC uintptr, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCSPVerificationFailedMissingEntry",
		"New module has doesn't have a PCSP entry for a PC within the patched code",
		nil,
		map[string]interface{}{
			"origValue": origValue,
			"origPC":    origPC,
			"newPC":     newPC,
		})
}

func NewPCFileVerificationFailed(origFile string, origPC uintptr, newFile string, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCFileVerificationFailed",
		"New module has a different value in pcfile table than the original module",
		nil,
		map[string]interface{}{
			"origFile": origFile,
			"origPC":   origPC,
			"newFile":  newFile,
			"newPC":    newPC,
		})
}

func NewPCLineVerificationFailed(origLine int32, origPC uintptr, newLine int32, newPC uintptr) RookoutError {
	return newRookoutError(
		"PCLineVerificationFailed",
		"New module has a different value in pcline table than the original module",
		nil,
		map[string]interface{}{
			"origLine": origLine,
			"origPC":   origPC,
			"newLine":  newLine,
			"newPC":    newPC,
		})
}

func NewDifferentNFuncData(origNFuncData uint8, newNFuncData uint8) RookoutError {
	return newRookoutError(
		"DifferenceNFuncData",
		"New module doesn't have the same number of funcdata tables as original module",
		nil,
		map[string]interface{}{
			"origNFuncData": origNFuncData,
			"newNFuncData":  newNFuncData,
		})
}

func NewFuncDataVerificationFailed(table int, origValue uintptr, newValue uintptr) RookoutError {
	return newRookoutError(
		"FuncDataVerificationFailed",
		"New module has a different pointer to funcdata than the original module",
		nil,
		map[string]interface{}{
			"table":     table,
			"origValue": origValue,
			"newValue":  newValue,
		})
}

func NewModuleVerificationFailed(recovered interface{}) RookoutError {
	return newRookoutError(
		"ModuleVerificationFailed",
		"Panic occured while trying to verify new moduledata",
		nil,
		map[string]interface{}{
			"recovered": recovered,
		})
}

func NewIllegalAddressMappings() RookoutError {
	return newRookoutError(
		"BadAddressMapping",
		"Function address mapping must not contain patched code in the last two mappings",
		nil,
		nil)
}

func NewVariableCreationFailed(recovered interface{}) RookoutError {
	return newRookoutError(
		"VariableCreationFailed",
		"Panic occured while trying to create new variable",
		nil,
		map[string]interface{}{
			"recovered": recovered,
		})
}

func NewVariableLoadFailed(recovered interface{}) RookoutError {
	return newRookoutError(
		"VariableLoadFailed",
		"Panic occured while trying to load variable",
		nil,
		map[string]interface{}{
			"recovered": recovered,
		})
}

func NewArgIsNotRel(inst interface{}) RookoutError {
	return newRookoutError(
		"ArgIsNotRel",
		"Unable to calculate absolute dest because first arg is not Rel",
		nil,
		map[string]interface{}{
			"inst": inst,
		})
}

func NewInvalidJumpDest(jumpDest string) RookoutError {
	return newRookoutError(
		"InvalidJumpDest",
		"Created a jump with a nonexistant dest",
		nil,
		map[string]interface{}{
			"jumpDest": jumpDest,
		})
}

func NewFailedToAssemble(recovered interface{}) RookoutError {
	return newRookoutError(
		"FailedToAssemble",
		"Failed to assemble instructions",
		nil,
		map[string]interface{}{
			"recovered": recovered,
		})
}

func NewFailedToDecode(funcAsm []byte, err error) RookoutError {
	return newRookoutError(
		"FailedToDecode",
		"Failed to decode one instruction",
		err,
		map[string]interface{}{
			"inst": fmt.Sprintf("%x", funcAsm),
		})
}

func NewUnexpectedInstructionOp(inst interface{}) RookoutError {
	return newRookoutError(
		"UnexpectedInstructionOp",
		"Unable to calculate dest PC of instruction that isn't CALL or JMP",
		nil,
		map[string]interface{}{
			"inst": inst,
		})
}

func NewKeyNotInMap(mapName string, key string) RookoutError {
	return newRookoutError(
		"KeyNotInMap",
		"Given key does not exist in map",
		nil,
		map[string]interface{}{
			"mapName": mapName,
			"key":     key,
		})
}

func NewNoSuchMember(structName string, memberName string) RookoutError {
	return newRookoutError(
		"NoSuchChild",
		"Struct doesn't have member with given name",
		nil,
		map[string]interface{}{
			"structName": structName,
			"memberName": memberName,
		})
}

func NewVariableIsNotMap(name string, kind reflect.Kind) RookoutError {
	return newRookoutError(
		"VariableIsNotMap",
		"Tried to get map value of variable that is not of kind map",
		nil,
		map[string]interface{}{
			"name": name,
			"kind": kind,
		})
}

func NewVariableIsNotStruct(name string, kind reflect.Kind) RookoutError {
	return newRookoutError(
		"VariableIsNotStruct",
		"Tried to get struct value of variable that is not of kind struct",
		nil,
		map[string]interface{}{
			"name": name,
			"kind": kind,
		})
}

func NewVariableIsNotArray(name string, kind reflect.Kind) RookoutError {
	return newRookoutError(
		"VariableIsNotArray",
		"Tried to get array item of variable that is not of kind array",
		nil,
		map[string]interface{}{
			"name": name,
			"kind": kind,
		})
}

func NewLabelAlreadyExists(label string) RookoutError {
	return newRookoutError(
		"LabelAlreadyExists",
		"Unable to add label to assembly: the label already exists",
		nil,
		map[string]interface{}{
			"label": label,
		})
}

func NewInvalidBytes(bytes []byte) RookoutError {
	return newRookoutError(
		"InvalidBytes",
		"Cannot insert bytes: length of bytes is not a multiple of 4",
		nil,
		map[string]interface{}{
			"bytes": bytes,
		})
}

func NewUnexpectedInstruction(movGToR12 interface{}, ret interface{}) RookoutError {
	return newRookoutError(
		"UnexpectedInstruction",
		"Unexpected instructions in assembled getg",
		nil,
		map[string]interface{}{
			"movGToR12": movGToR12,
			"ret":       ret,
		})
}
