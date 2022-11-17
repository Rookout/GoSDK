package augs

import (
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"sync"
	"time"
)

type AugTimeLimiter struct {
	disableAug bool
	maxAugTime time.Duration

	executionStartTimesLock sync.Mutex
	executionStartTimes     map[string]time.Time
}

func NewAugTimeLimiter(maxAugTime time.Duration) *AugTimeLimiter {
	return &AugTimeLimiter{
		disableAug:          false,
		maxAugTime:          maxAugTime,
		executionStartTimes: make(map[string]time.Time),
	}
}

func (t *AugTimeLimiter) BeforeRun(executionId string) (types.AugStatus, rookoutErrors.RookoutError) {
	t.executionStartTimesLock.Lock()
	defer t.executionStartTimesLock.Unlock()

	if t.disableAug {
		return types.Error, rookoutErrors.NewRookRuleMaxExecutionTimeReached()
	}

	t.executionStartTimes[executionId] = time.Now()
	return types.Active, nil
}

func (t *AugTimeLimiter) CancelRun(executionId string) {
	t.executionStartTimesLock.Lock()
	defer t.executionStartTimesLock.Unlock()

	delete(t.executionStartTimes, executionId)
}

func (t *AugTimeLimiter) AfterRun(executionId string) (types.AugStatus, rookoutErrors.RookoutError) {
	t.executionStartTimesLock.Lock()
	defer t.executionStartTimesLock.Unlock()

	augTime := time.Now().Sub(t.executionStartTimes[executionId])
	delete(t.executionStartTimes, executionId)

	if t.maxAugTime > 0 && augTime > t.maxAugTime {
		t.disableAug = true
		return types.Error, rookoutErrors.NewRookRuleMaxExecutionTimeReached()
	}

	return types.Active, nil
}
