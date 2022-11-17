package namespaces

import (
	"github.com/Rookout/GoSDK/pkg/config"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)



func GetErrorAsProtobuf(err rookoutErrors.RookoutError) *pb.Error {
	if err == nil {
		return nil
	}

	stacktrace := string(err.Stack())
	traceback := NewGoObjectNamespace(stacktrace)
	traceback.SetObjectDumpConfig(config.GetTailoredLimits(stacktrace)) 
	return &pb.Error{
		Message:    err.Error(),
		Type:       err.GetType(),
		Parameters: NewGoObjectNamespace(err.GetArguments()).ToProtobuf(false),
		Exc:        NewGoObjectNamespace(err.StackFrames()).ToProtobuf(false),
		Traceback:  traceback.ToProtobuf(false)}
}
