package suspender

type Suspender interface {
	StopAll()
	ResumeAll()
	Stopped() bool
}
