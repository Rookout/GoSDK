//go:build (!amd64 && !arm64) || !go1.15 || go1.22
// +build !amd64,!arm64 !go1.15 go1.22

package callstack

func (s *StackTraceBuffer) FillStackTraces() (int, bool) {
	callstackStubPanic()
	return 0, false
}

func callstackStubPanic() {
	panic("Callstack doesn't support this go version!!!!")
}
