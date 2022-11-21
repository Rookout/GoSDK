//go:build !arm64 || !darwin
// +build !arm64 !darwin

package safe_hook_installer

func setWritable(address uintptr, size int) int {
	
	return 0
}

func setExecutable(address uintptr, size int) int {
	
	return 0
}
