package information

import (
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"net"
)

func resolveHostIp() string {

	netInterfaceAddresses, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, addr := range netInterfaceAddresses {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if !ip.IsLoopback() && ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}

func collectNetwork(info *AgentInformation) error {
	info.Network = &pb.NetworkInformation{
		IpAddr: resolveHostIp(),
	}
	return nil
}
