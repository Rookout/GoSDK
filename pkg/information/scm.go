package information

import (
	"os"

	pb "github.com/Rookout/GoSDK/pkg/protobuf"
)

var GitConfig = struct {
	Commit       string
	RemoteOrigin string
	Sources      map[string]string
}{}

func collectSources(info *AgentInformation) {
	if len(GitConfig.Sources) == 0 {
		return
	}

	info.Scm.Sources = make([]*pb.SCMInformation_SourceInfo, len(GitConfig.Sources))
	i := 0
	for remoteOriginUrl, commit := range GitConfig.Sources {
		source := &pb.SCMInformation_SourceInfo{RemoteOriginUrl: remoteOriginUrl, Commit: commit}
		info.Scm.Sources[i] = source
		i += 1
	}
}

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
	collectSources(info)
	return nil
}
