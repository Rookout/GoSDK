package information

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"io/ioutil"
	"os"
	"runtime"
)

func collectPlatform(info *AgentInformation) error {
	platformInfo := &pb.PlatformInformation{
		Platform: "golang",
		Version:  runtime.Version(),
		Variant:  "golang",
	}
	info.Platform = platformInfo
	return nil
}

func collectK8sNamespace(information *AgentInformation) error {
	defaultK8sNamespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	if information.K8sNamespaceFileName == "" {
		information.K8sNamespaceFileName = defaultK8sNamespaceFile
	}

	if namespace, err := ioutil.ReadFile(information.K8sNamespaceFileName); err == nil {
		information.K8sNamespace = string(namespace)
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
