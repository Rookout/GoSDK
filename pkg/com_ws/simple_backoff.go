package com_ws

import (
	"context"
	"time"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
)

type Backoff interface {
	AfterConnect()
	AfterDisconnect(context.Context)
}

type backoff struct {
	connectTime time.Time
	connected   bool
	nextBackoff time.Duration
}

func NewBackoff() *backoff {
	return &backoff{
		nextBackoff: config.BackoffConfig().DefaultBackoff,
	}
}

func (b *backoff) AfterConnect() {
	b.connectTime = time.Now()
	b.connected = true
}

func (b *backoff) AfterDisconnect(ctx context.Context) {
	if b.connected && time.Since(b.connectTime) > config.BackoffConfig().ResetBackoffTimeout {
		b.reset()
	}

	b.connected = false
	logger.Logger().Debugf("Backing off %g seconds\n", b.nextBackoff.Seconds())
	timer := time.NewTimer(b.nextBackoff)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
		return
	}

	b.nextBackoff *= 2
	maxBackoff := config.BackoffConfig().MaxBackoff
	if maxBackoff != 0 && b.nextBackoff > maxBackoff {
		b.nextBackoff = maxBackoff
	}
}

func (b *backoff) reset() {
	b.nextBackoff = config.BackoffConfig().DefaultBackoff
}
