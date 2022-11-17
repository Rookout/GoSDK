package information

import (
	"github.com/go-errors/errors"
	"os"
)

func collectProcessId(information *AgentInformation) error {
	information.ProcessId = uint32(os.Getpid())
	return nil
}

func collectCommandArgs(information *AgentInformation) error {
	information.CommandArguments = os.Args[1:]
	return nil
}

func collectExecutable(information *AgentInformation) error {
	exec, err := os.Executable()
	if err != nil {
		return errors.New(err)
	}
	information.Executable = exec
	return nil
}
