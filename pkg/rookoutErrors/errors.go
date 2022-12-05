package rookoutErrors

import (
	"fmt"
	"runtime"

	"github.com/go-errors/errors"
)

type RookoutError interface {
	error

	StackFrames() []errors.StackFrame
	Stack() []byte

	GetType() string

	GetArguments() map[interface{}]interface{}
}

type RookoutErrorImpl struct {
	ExternalError error
	Type          string `json:"Type"`
	Arguments     map[string]interface{}
}

func (r RookoutErrorImpl) Error() string {
	errorString := r.Type

	if nil != r.ExternalError {
		errorString += ": " + r.ExternalError.Error()
	}

	if len(r.Arguments) != 0 {
		errorString = fmt.Sprintf(errorString+" | %v", r.Arguments)
	}

	return errorString
}

func (r RookoutErrorImpl) GetType() string {
	return r.Type
}

func (r RookoutErrorImpl) GetArguments() map[interface{}]interface{} {
	outputMap := make(map[interface{}]interface{})
	for key, value := range r.Arguments {
		outputMap[key] = value
	}
	return outputMap
}

func (r RookoutErrorImpl) StackFrames() []errors.StackFrame {
	switch e := r.ExternalError.(type) {
	case *errors.Error:
		return e.StackFrames()
	case *RookoutErrorImpl:
		return e.StackFrames()
	default:
		return errors.New(e).StackFrames()
	}

}

func (r RookoutErrorImpl) Stack() []byte {
	switch e := r.ExternalError.(type) {
	case *errors.Error:
		return e.Stack()
	case *RookoutErrorImpl:
		return e.Stack()
	default:
		return errors.New(e).Stack()
	}

}

func newRookoutError(errorType string, description string, externalError error, arguments map[string]interface{}) RookoutError {
	if externalError == nil {
		externalError = errors.Wrap(description, 2)
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

func NewRookPanicInGoroutine(recovered interface{}) RookoutError {
	err, _ := recovered.(error)

	return newRookoutError(
		"RookPanicInGoroutine",
		"Caught panic in goroutine",
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

func NewFailedToGetStackUsageMap(reason string) RookoutError {
	return newRookoutError(
		"FailedToGetStackUsageMap",
		"Failed to get stack usage map from native",
		nil,
		map[string]interface{}{
			"reason": reason,
		})
}

func NewFailedToParseStackUsageMap(buffer string, externalErr error) RookoutError {
	return newRookoutError(
		"FailedToParseStackUsageMap",
		"Failed to parse stack usage map from native",
		externalErr,
		map[string]interface{}{
			"buffer": buffer,
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
