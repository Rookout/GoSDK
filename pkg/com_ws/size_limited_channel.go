package com_ws

import (
	"context"
	"sync"
	"time"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)



type SizeLimitedChannel struct {
	channel        chan []byte
	bytesInChannel int
	channelLock    sync.Mutex
	doneChannel    chan []struct{}
	flushing       bool
}

func NewSizeLimitedChannel() *SizeLimitedChannel {
	return &SizeLimitedChannel{
		channel:     make(chan []byte, config.SizeLimitedChannelConfig().MaxQueueLength),
		doneChannel: make(chan []struct{}, 1),
	}
}
func (s *SizeLimitedChannel) Offer(message []byte) rookoutErrors.RookoutError {
	s.channelLock.Lock()
	defer s.channelLock.Unlock()

	maxMessageSize := config.SizeLimitedChannelConfig().MaxMessageSize
	if len(message) > maxMessageSize {
		
		
		return rookoutErrors.NewRookMessageSizeExceeded(len(message), maxMessageSize)
	}
	if s.bytesInChannel+len(message) > config.SizeLimitedChannelConfig().MaxBytesInChannel {
		return rookoutErrors.NewRookOutputQueueFull()
	}

	select {
	case s.channel <- message:
		s.bytesInChannel += len(message)
		return nil
	default:
		return rookoutErrors.NewRookOutputQueueFull()
	}
}

func (s *SizeLimitedChannel) Poll(ctx context.Context) []byte {
	select {
	case message := <-s.channel:
		s.channelLock.Lock()
		defer s.channelLock.Unlock()

		s.bytesInChannel -= len(message)

		if s.bytesInChannel == 0 && s.flushing {
			select {
			case s.doneChannel <- nil:
			default:
			}
		}

		return message
	case <-ctx.Done():
		return nil
	}
}

func (s *SizeLimitedChannel) setFlushing(state bool) {
	s.channelLock.Lock()
	defer s.channelLock.Unlock()
	s.flushing = state
}

func (s *SizeLimitedChannel) Flush() rookoutErrors.RookoutError {
	if s.bytesInChannel == 0 {
		return nil
	}

	s.setFlushing(true)
	defer func() { s.setFlushing(false) }()

	timeout := config.SizeLimitedChannelConfig().FlushTimeout

	select {
	case <-s.doneChannel:
		return nil
	case <-time.After(timeout):
		return rookoutErrors.NewFlushTimedOut()
	}
}
