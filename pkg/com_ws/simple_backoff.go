package com_ws

import (
	"context"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"sync/atomic"
	"time"
	"unsafe"
)

type Backoff interface {
	UpdateConfig(config.BackoffConfig)
	AfterConnect()
	AfterDisconnect(context.Context)
}

type backoff struct {
	config      *config.BackoffConfig
	connectTime time.Time
	connected   bool
	nextBackoff time.Duration
}

func NewBackoff(config config.BackoffConfig) *backoff {
	return &backoff{
		nextBackoff: config.DefaultBackoff,
		config:      &config,
	}
}

func (b *backoff) UpdateConfig(config config.BackoffConfig) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&b.config)), unsafe.Pointer(&config))
}

func (b *backoff) AfterConnect() {
	b.connectTime = time.Now()
	b.connected = true
}

func (b *backoff) AfterDisconnect(ctx context.Context) {
	if b.connected && time.Now().Sub(b.connectTime) > b.config.ResetBackoffTimeout {
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
	if b.config.MaxBackoff != 0 && b.nextBackoff > b.config.MaxBackoff {
		b.nextBackoff = b.config.MaxBackoff
	}
}

func (b *backoff) reset() {
	b.nextBackoff = b.config.DefaultBackoff
}
