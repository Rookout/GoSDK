//go:build (!amd64 && !arm64) || !go1.16 || go1.23
// +build !amd64,!arm64 !go1.16 go1.23

package callstack

func (s *StackTraceBuffer) FillStackTraces() (int, bool) {
	callstackStubPanic()
	return 0, false
}

func callstackStubPanic() {
	panic("Callstack doesn't support this go version!!!!")
}
