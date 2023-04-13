package augs

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
)

type LimitProvider struct {
	GlobalRateLimiter *AugRateLimiter
}

var initOnce sync.Once
var rookLimitProvider *LimitProvider

func GetLimitProvider() *LimitProvider {
	if rookLimitProvider == nil {
		InitLimitProvider()
	}

	return rookLimitProvider
}


func InitLimitProvider() {
	initOnce.Do(func() {
		initializedSingleton := createLimitProvider()
		rookLimitProvider = initializedSingleton
	})
}

func createLimitProvider() *LimitProvider {
	l := &LimitProvider{
		GlobalRateLimiter: nil,
	}
	config.OnUpdate(l.updateConfig)
	return l
}

func (l *LimitProvider) updateConfig() {
	if config.RateLimiterConfig().GlobalRateLimit != "" {
		l.tryToCreateGlobalRateLimiter()
	} else {
		l.DeleteGlobalRateLimiter()
	}
}

func (l *LimitProvider) DeleteGlobalRateLimiter() {
	l.GlobalRateLimiter = nil
}

func (l *LimitProvider) GetLimitManager(configuration types.AugConfiguration, augID types.AugID, output com_ws.Output) (LimitsManager, rookoutErrors.RookoutError) {
	limitsManager := NewLimitsManager(augID, output)

	rateLimiter, err := l.getRateLimiter(configuration)
	if err != nil {
		return nil, err
	}

	if rateLimiter != nil {
		limitsManager.AddLimiter(rateLimiter)
	}

	limitsManager.AddLimiter(l.getAugTimeLimiter(configuration))

	return limitsManager, nil
}

func (l *LimitProvider) getAugTimeLimiter(configuration types.AugConfiguration) *AugTimeLimiter {
	maxAugTime := config.LocationsConfig().MaxAugTime
	maxAugTimeStr, ok := configuration["maxAugTime"].(string)
	if ok {
		maxAugTimeMS, err := strconv.ParseInt(maxAugTimeStr, 10, 64)
		if err != nil {
			logger.Logger().WithError(err).Errorln("Failed to parse max aug time configuration")
		} else {
			maxAugTime = time.Duration(maxAugTimeMS) * time.Millisecond
		}
	}

	maxAugTime = maxAugTime * time.Duration(config.LocationsConfig().MaxAugTimeMultiplier)

	return NewAugTimeLimiter(maxAugTime)
}

func (l *LimitProvider) getRateLimiter(configuration types.AugConfiguration) (*AugRateLimiter, rookoutErrors.RookoutError) {
	
	if l.GlobalRateLimiter != nil {
		return l.GlobalRateLimiter, nil
	} else {
		windowQuota := utils.MSToNS(200)
		windowSize := utils.MSToNS(500)
		rateLimitSpec, ok := configuration["rateLimit"].(string)

		var err error
		rateLimitModifier := 0
		rateLimitModifierStr, ok := configuration["rateLimitModifier"].(string)
		if ok {
			rateLimitModifier, err = strconv.Atoi(rateLimitModifierStr)
			if err != nil {
				logger.Logger().WithError(err).Errorln("Failed to parse rate limit configuration")
			}
		}

		rateLimiter, rookErr := l.createRateLimiter(rateLimitSpec, windowQuota, windowSize, rateLimitModifier)
		if rookErr != nil {
			return nil, rookErr
		}

		return rateLimiter, nil
	}
}

func (l *LimitProvider) tryToCreateGlobalRateLimiter() {
	globalRateLimit := config.RateLimiterConfig().GlobalRateLimit
	globalRateLimiter, err := l.createRateLimiter(globalRateLimit, 0, 0, 0)
	if globalRateLimiter == nil {
		err = rookoutErrors.NewRookInvalidRateLimitConfiguration(globalRateLimit)
	}

	if err != nil {
		logger.Logger().WithError(err).Warningln("Failed to create global rate limiter")
		return
	}

	rookoutErrors.UsingGlobalRateLimiter = true
	l.GlobalRateLimiter = globalRateLimiter
}

func (l *LimitProvider) createRateLimiter(limitsSpec string, defaultQuota time.Duration, defaultWindowSize time.Duration, activeLimit int) (*AugRateLimiter, rookoutErrors.RookoutError) {
	quota := defaultQuota
	windowSize := defaultWindowSize

	if limitsSpec != "" {
		limits := strings.Split(limitsSpec, "/")

		if len(limits) == 2 {
			var err error

			quota, err = utils.StringMSToNS(limits[0])
			if err != nil {
				quota = defaultQuota
			} else {
				windowSize, err = utils.StringMSToNS(limits[1])
				if err != nil {
					quota = defaultQuota
					windowSize = defaultWindowSize
				}
			}
		}
	}

	if quota == 0 {
		return nil, nil
	}

	if quota >= windowSize {
		return nil, rookoutErrors.NewRookInvalidRateLimitConfiguration(limitsSpec)
	}

	return NewAugRateLimiter(quota, windowSize, activeLimit, config.RateLimiterConfig()), nil
}
