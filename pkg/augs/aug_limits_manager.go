package augs

import (
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type LimitsManager interface {
	AddLimiter(newLimiter Limiter)
	GetAllLimiters() []Limiter
	BeforeRun(executionId string) bool
	AfterRun(executionId string)
}

type Limiter interface {
	BeforeRun(executionId string) (types.AugStatus, rookoutErrors.RookoutError)
	CancelRun(executionId string)
	AfterRun(executionId string) (types.AugStatus, rookoutErrors.RookoutError)
}

type limitsManager struct {
	limiters  []Limiter
	augId     types.AugId
	augStatus types.AugStatus
	output    com_ws.Output
}

func NewLimitsManager(augId types.AugId, output com_ws.Output) LimitsManager {
	return &limitsManager{augId: augId, output: output, augStatus: types.Active}
}

func (l *limitsManager) AddLimiter(newLimiter Limiter) {
	l.limiters = append(l.limiters, newLimiter)
}

func (l *limitsManager) GetAllLimiters() []Limiter {
	return l.limiters
}

func (l *limitsManager) cancelAllLimiters(executionId string) {
	for _, limiter := range l.limiters {
		limiter.CancelRun(executionId)
	}
}

func (l *limitsManager) setupAllLimiters(executionId string) bool {
	for _, limiter := range l.limiters {
		status, err := limiter.BeforeRun(executionId)
		if status == types.Active {
			continue
		}

		if l.augStatus == types.Active {
			l.augStatus = status
			_ = l.output.SendRuleStatus(l.augId, l.augStatus, err)
		}

		return false
	}

	if l.augStatus != types.Active {
		l.augStatus = types.Active
		_ = l.output.SendRuleStatus(l.augId, l.augStatus, nil)
	}

	return true
}

func (l *limitsManager) BeforeRun(executionId string) bool {
	if ok := l.setupAllLimiters(executionId); !ok {
		l.cancelAllLimiters(executionId)
		return false
	}

	return true
}

func (l *limitsManager) AfterRun(executionId string) {
	for _, limiter := range l.limiters {
		status, err := limiter.AfterRun(executionId)
		if status == types.Active {
			continue
		}

		if l.augStatus == types.Active {
			l.augStatus = status
			_ = l.output.SendRuleStatus(l.augId, l.augStatus, err)
		}
	}
}
