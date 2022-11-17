package augs

import (
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"sync"
	"time"
)

type AugRateLimiter struct {
	config config.RateLimiterConfiguration

	WindowSize   time.Duration
	Quota        time.Duration
	activeWeight int64

	windowUsages map[int64]time.Duration

	dataLock    sync.Mutex
	currentData map[string]executionData
}

type executionData struct {
	windowIndex int64
	startTime   time.Time
}

func NewAugRateLimiter(quota time.Duration, windowSize time.Duration, activeLimit int, config config.RateLimiterConfiguration) *AugRateLimiter {
	if windowSize <= 0 {
		windowSize = 1
	}

	var activeWeight int64
	if activeLimit != 0 {
		activeWeight = int64(quota) / int64(activeLimit)
	} else {
		activeWeight = 0
	}

	return &AugRateLimiter{
		config:       config,
		WindowSize:   windowSize,
		Quota:        quota,
		activeWeight: activeWeight,
		windowUsages: make(map[int64]time.Duration),
		currentData:  make(map[string]executionData),
	}
}

func (r *AugRateLimiter) activeCount() int64 {
	return int64(len(r.currentData))
}

func (r *AugRateLimiter) currentUsage(startTime time.Time, currentWindowIndex int64) int64 {
	
	currentWindowUsage, ok := r.windowUsages[currentWindowIndex]
	if !ok {
		
		r.windowUsages[currentWindowIndex] = 0
		currentWindowUsage = 0
	}

	
	prevWindowUsage, ok := r.windowUsages[currentWindowIndex-1]
	if !ok {
		prevWindowUsage = 0
	}
	
	timeInWindow := float64(startTime.UnixNano() % r.WindowSize.Nanoseconds())
	
	
	prevWeight := 1 - (timeInWindow / float64(r.WindowSize))

	
	prevCount := int64(float64(prevWindowUsage.Nanoseconds()) * prevWeight)
	activeCount := r.activeCount() * r.activeWeight
	return prevCount + activeCount + currentWindowUsage.Nanoseconds()
}

func (r *AugRateLimiter) cleanup(currentWindowIndex int64) {
	if len(r.windowUsages) > 10 {
		for windowIndex := range r.windowUsages {
			if windowIndex < currentWindowIndex-5 {
				delete(r.windowUsages, windowIndex)
			}
		}
	}
}

func (r *AugRateLimiter) BeforeRun(executionId string) (types.AugStatus, rookoutErrors.RookoutError) {
	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	startTime := time.Now()
	currentWindowIndex := startTime.UnixNano() / r.WindowSize.Nanoseconds() 
	r.currentData[executionId] = executionData{
		startTime:   startTime,
		windowIndex: currentWindowIndex,
	}

	r.cleanup(currentWindowIndex)

	if r.Quota > 0 && r.currentUsage(startTime, currentWindowIndex) > int64(r.Quota) {
		return types.Warning, rookoutErrors.NewRookRuleRateLimited()
	}

	return types.Active, nil
}

func (r *AugRateLimiter) CancelRun(executionId string) {
	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	delete(r.currentData, executionId)
}

func (r *AugRateLimiter) AfterRun(executionId string) (types.AugStatus, rookoutErrors.RookoutError) {
	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	data := r.currentData[executionId]
	duration := time.Now().Sub(data.startTime)
	if duration < r.config.MinRateLimitValue {
		duration = r.config.MinRateLimitValue
	}
	if totalUsage, ok := r.windowUsages[data.windowIndex]; ok {
		r.windowUsages[data.windowIndex] = totalUsage + duration
	}

	delete(r.currentData, executionId)

	return types.Active, nil
}
