package information

import pb "github.com/Rookout/GoSDK/pkg/protobuf"

const VERSION = "0.1.39"

func collectVersion(info *AgentInformation) error {
	info.Version = &pb.VersionInformation{
		Version: VERSION,
		Commit:  "",
	}
	return nil
}
