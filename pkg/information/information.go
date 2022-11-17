package information

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
)

type collector func(information *AgentInformation) error

func collectors() map[string]collector {
	return map[string]collector{
		"version":           collectVersion,
		"network":           collectNetwork,
		"system":            collectSystem,
		"platform":          collectPlatform,
		"scm":               collectScm,
		"executable":        collectExecutable,
		"command_arguments": collectCommandArgs,
		"process_id":        collectProcessId,
		"k8s_namespace":     collectK8sNamespace,
		"serverless_info":   collectServerless,
	}
}

type AgentInformation struct {
	pb.AgentInformation
	DefaultK8sNamespaceFile string
	K8sNamespaceFileName    string
	K8sNamespace            string
}

func Collect(labels map[string]string, k8sNamespaceFile string) (*pb.AgentInformation, error) {
	k8sNamespaceLabel := "k8s_namespace"
	info := &AgentInformation{DefaultK8sNamespaceFile: "/var/run/secrets/kubernetes.io/serviceaccount/namespace"}

	info.Labels = labels
	if k8sNamespaceFile != "" {
		info.K8sNamespaceFileName = k8sNamespaceFile
	}
	for _, collector := range collectors() {
		err := collector(info)
		if err != nil {
			return nil, err
		}
	}

	if info.K8sNamespace != "" {
		if info.Labels == nil {
			info.Labels = make(map[string]string, 1)
		}
		if _, exists := info.Labels[k8sNamespaceLabel]; !exists {
			info.Labels[k8sNamespaceLabel] = info.K8sNamespace
		}
	}

	return &info.AgentInformation, nil
}
