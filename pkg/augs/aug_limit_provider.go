package augs

import (
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type LimitProvider struct {
	config            *config.LocationsConfiguration
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
	return &LimitProvider{
		GlobalRateLimiter: nil,
	}
}

func (l *LimitProvider) UpdateConfig(config config.LocationsConfiguration) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&l.config)), unsafe.Pointer(&config))

	if l.config.RateLimiterConfiguration.GlobalRateLimit != "" {
		l.tryToCreateGlobalRateLimiter()
	} else {
		l.DeleteGlobalRateLimiter()
	}
}

func (l *LimitProvider) DeleteGlobalRateLimiter() {
	l.GlobalRateLimiter = nil
}

func (l *LimitProvider) GetLimitManager(configuration types.AugConfiguration, augId types.AugId, output com_ws.Output) (LimitsManager, rookoutErrors.RookoutError) {
	limitsManager := NewLimitsManager(augId, output)

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
	maxAugTime := l.config.MaxAugTime
	maxAugTimeStr, ok := configuration["maxAugTime"].(string)
	if ok {
		maxAugTimeMS, err := strconv.ParseInt(maxAugTimeStr, 10, 64)
		if err != nil {
			logger.Logger().WithError(err).Errorln("Failed to parse max aug time configuration")
		} else {
			maxAugTime = time.Duration(maxAugTimeMS) * time.Millisecond
		}
	}

	maxAugTime = maxAugTime * time.Duration(l.config.MaxAugTimeMultiplier)

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
	globalRateLimiter, err := l.createRateLimiter(l.config.RateLimiterConfiguration.GlobalRateLimit,
		0, 0, 0)
	if globalRateLimiter == nil {
		err = rookoutErrors.NewRookInvalidRateLimitConfiguration(l.config.RateLimiterConfiguration.GlobalRateLimit)
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

	return NewAugRateLimiter(quota, windowSize, activeLimit, l.config.RateLimiterConfiguration), nil
}
