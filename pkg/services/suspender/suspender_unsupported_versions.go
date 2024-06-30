//go:build !go1.16 || go1.23
// +build !go1.16 go1.23

package suspender

type suspenderStub struct {
}

func (s *suspenderStub) StopAll() {
	suspenderStubPanic()
}

func (s *suspenderStub) ResumeAll() {
	suspenderStubPanic()
}

func (s *suspenderStub) Stopped() bool {
	suspenderStubPanic()
	return false
}

func GetSuspender() Suspender {
	suspenderStubPanic()
	return nil
}

func suspenderStubPanic() {
	panic("Suspender doesn't support this go version!!!!")
}
