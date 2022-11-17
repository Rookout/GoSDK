package utils

import "time"

type Backoff struct {
	currentBackoff time.Duration
	maxBackoff     time.Duration
}

type BackoffOption interface {
	Apply(*Backoff)
}

type MaxBackoffOption struct {
	maxBackoff time.Duration
}

func (m MaxBackoffOption) Apply(backoff *Backoff) {
	backoff.maxBackoff = m.maxBackoff
}

func WithMaxBackoff(maxBackoff time.Duration) MaxBackoffOption {
	return MaxBackoffOption{maxBackoff: maxBackoff}
}

type InitialBackoff struct {
	initialBackoff time.Duration
}

func (i InitialBackoff) Apply(backoff *Backoff) {
	backoff.currentBackoff = i.initialBackoff
}

func WithInitialBackoff(initialBackoff time.Duration) InitialBackoff {
	return InitialBackoff{initialBackoff: initialBackoff}
}

const defaultBackoff = 1 * time.Second


func NewBackoff(opts ...BackoffOption) *Backoff {
	b := &Backoff{}
	for _, opt := range opts {
		opt.Apply(b)
	}
	return b
}

func (b *Backoff) NextBackoff() time.Duration {
	if b.currentBackoff == 0 {
		b.currentBackoff = defaultBackoff
	}

	newBackoff := b.currentBackoff * 2
	if b.maxBackoff != 0 {
		if newBackoff > b.maxBackoff {
			newBackoff = b.maxBackoff
		}
	}

	b.currentBackoff = newBackoff
	return b.currentBackoff
}

func (b *Backoff) MaxBackoff() time.Duration {
	return b.maxBackoff
}
