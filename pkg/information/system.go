package information

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/go-errors/errors"
	"os"
	"runtime"
)

func collectSystem(information *AgentInformation) error {
	hostname, err := os.Hostname()
	if err != nil {
		return errors.New(err)
	}

	information.System = &pb.SystemInformation{
		Hostname: hostname,
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}

	return nil

}
