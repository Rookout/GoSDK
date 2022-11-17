package information

import (
	"os"

	pb "github.com/Rookout/GoSDK/pkg/protobuf"
)

var GitConfig = struct {
	Commit       string
	RemoteOrigin string
}{}

func collectScm(info *AgentInformation) error {
	info.Scm = &pb.SCMInformation{}

	var gitRoot string
	if GitConfig.Commit == "" || GitConfig.RemoteOrigin == "" {
		if cwd, err := os.Getwd(); err == nil {
			gitRoot = FindRoot(cwd)
		}
	}

	if GitConfig.Commit == "" {
		GitConfig.Commit = GetRevision(gitRoot)
	}
	if GitConfig.RemoteOrigin == "" {
		GitConfig.RemoteOrigin = GetRemoteOrigin(gitRoot)
	}

	info.Scm.Commit = GitConfig.Commit
	info.Scm.Origin = GitConfig.RemoteOrigin
	return nil
}
