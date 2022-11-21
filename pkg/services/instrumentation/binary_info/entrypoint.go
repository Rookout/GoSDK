//go:build !arm64 || !darwin
// +build !arm64 !darwin

package binary_info

func GetEntrypoint(exeName string) uint64 {
	return 0
}
