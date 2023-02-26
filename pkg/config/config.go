package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type LocationsConfiguration struct {
	MaxAugTime               time.Duration
	MaxAugTimeMultiplier     float64
	RateLimiterConfiguration RateLimiterConfiguration
}

type RateLimiterConfiguration struct {
	MinRateLimitValue           time.Duration
	GlobalRateLimitQuotaMS      string
	GlobalRateLimitWindowSizeMS string
	GlobalRateLimit             string
}

type LoggingConfiguration struct {
	FileName    string
	LogToStderr bool
	LogLevel    string
	Debug       bool
	MaxLogSize  int
	Quiet       bool
}

type AgentComWsConfiguration struct {
	PingTimeout                     time.Duration
	ConnectTimeout                  time.Duration
	ConnectionTimeout               time.Duration
	SizeLimitedChannelConfiguration SizeLimitedChannelConfiguration
	BackoffConfig                   BackoffConfig
	WebSocketClientConfig           WebSocketClientConfig
}

type BackoffConfig struct {
	DefaultBackoff      time.Duration
	MaxBackoff          time.Duration
	ResetBackoffTimeout time.Duration
}

type ObjectDumpConfigDefaults struct {
	defaultConfig   ObjectDumpConfig
	tolerantConfig  ObjectDumpConfig
	strictConfig    ObjectDumpConfig
	unlimitedConfig ObjectDumpConfig
}

type SizeLimitedChannelConfiguration struct {
	FlushTimeout      time.Duration
	MaxQueueLength    int
	MaxBytesInChannel int
	MaxMessageSize    int
}

type OutputWsConfiguration struct {
	MaxStatusUpdates  int
	BucketRefreshRate int
	MaxAugMessages    int
	MaxLogItems       int
}

type WebSocketClientConfig struct {
	PingTimeout  time.Duration
	PingInterval time.Duration
	WriteTimeout time.Duration
}

type DynamicConfiguration struct {
	AgentComWsConfiguration  AgentComWsConfiguration
	LocationsConfiguration   LocationsConfiguration
	LoggingConfiguration     LoggingConfiguration
	ObjectDumpConfigDefaults ObjectDumpConfigDefaults
	OutputWsConfiguration    OutputWsConfiguration
}

func GetDefaultConfiguration() DynamicConfiguration {
	return DynamicConfiguration{
		LocationsConfiguration: LocationsConfiguration{
			MaxAugTime:           400 * time.Millisecond,
			MaxAugTimeMultiplier: 1,
			RateLimiterConfiguration: RateLimiterConfiguration{
				MinRateLimitValue:           20000 * time.Nanosecond,
				GlobalRateLimit:             os.Getenv("ROOKOUT_GLOBAL_RATE_LIMIT"),
				GlobalRateLimitQuotaMS:      "",
				GlobalRateLimitWindowSizeMS: "",
			},
		},
		LoggingConfiguration: LoggingConfiguration{
			FileName:    "",
			LogToStderr: false,
			LogLevel:    "INFO",
			Debug:       false,
			Quiet:       false,
			MaxLogSize:  100 * 1024 * 1024, 
		},
		AgentComWsConfiguration: AgentComWsConfiguration{
			ConnectTimeout:    10 * time.Minute,
			PingTimeout:       10 * time.Second,
			ConnectionTimeout: 8 * time.Second,
			SizeLimitedChannelConfiguration: SizeLimitedChannelConfiguration{
				FlushTimeout:      2 * time.Second,
				MaxQueueLength:    250,
				MaxBytesInChannel: 15 * 1024 * 1024,
				MaxMessageSize:    1024 * 1024,
			},
			BackoffConfig: BackoffConfig{
				DefaultBackoff:      200 * time.Millisecond,
				MaxBackoff:          60 * time.Second,
				ResetBackoffTimeout: 3 * time.Minute,
			},
			WebSocketClientConfig: WebSocketClientConfig{
				PingTimeout:  30 * time.Second,
				PingInterval: 10 * time.Second,
				WriteTimeout: 5 * time.Second,
			},
		},
		ObjectDumpConfigDefaults: ObjectDumpConfigDefaults{
			unlimitedConfig: ObjectDumpConfig{
				MaxDepth:           0,
				MaxWidth:           100,
				MaxCollectionDepth: 0,
				MaxString:          64 * 1024,
			},
			defaultConfig: ObjectDumpConfig{
				MaxDepth:           4,
				MaxWidth:           15,
				MaxCollectionDepth: 4,
				MaxString:          512,
			},
			tolerantConfig: ObjectDumpConfig{
				MaxDepth:           5,
				MaxWidth:           25,
				MaxCollectionDepth: 5,
				MaxString:          4 * 1024,
			},
			strictConfig: ObjectDumpConfig{
				MaxDepth:           2,
				MaxWidth:           10,
				MaxCollectionDepth: 2,
				MaxString:          128,
			},
		},
		OutputWsConfiguration: OutputWsConfiguration{
			MaxStatusUpdates:  200,
			BucketRefreshRate: 10,
			MaxAugMessages:    250,
			MaxLogItems:       200,
		},
	}
}

func UpdateGlobalRateLimitConfig(config *DynamicConfiguration) {
	if os.Getenv("ROOKOUT_GLOBAL_RATE_LIMIT") != "" {
		return
	}

	if config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimit == "" {
		if config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitQuotaMS != "" &&
			config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS != "" {
			config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimit = fmt.Sprintf("%s/%s",
				config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitQuotaMS,
				config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS)
		}
	}
}

var configParsers = map[string]func(string, *DynamicConfiguration){
	"GOLANG_DEFAULT_MAX_DEPTH": func(value string, config *DynamicConfiguration) {
		defaultMaxDepth, err := strconv.Atoi(value)
		if err != nil {
			return
		}

		config.ObjectDumpConfigDefaults.defaultConfig.MaxDepth = defaultMaxDepth
	},
	"GOLANG_DEFAULT_MAX_COLLECTION_DEPTH": func(value string, config *DynamicConfiguration) {
		defaultMaxCollectionDepth, err := strconv.Atoi(value)
		if err != nil {
			return
		}

		config.ObjectDumpConfigDefaults.defaultConfig.MaxCollectionDepth = defaultMaxCollectionDepth
	},
	"GOLANG_TOLERANT_MAX_DEPTH": func(value string, config *DynamicConfiguration) {
		tolerantMaxDepth, err := strconv.Atoi(value)
		if err != nil {
			return
		}

		config.ObjectDumpConfigDefaults.tolerantConfig.MaxDepth = tolerantMaxDepth
	},
	"GOLANG_TOLERANT_MAX_WIDTH": func(value string, config *DynamicConfiguration) {
		tolerantMaxWidth, err := strconv.Atoi(value)
		if err != nil {
			return
		}

		config.ObjectDumpConfigDefaults.tolerantConfig.MaxCollectionDepth = tolerantMaxWidth
	},
	"GOLANG_MAX_MESSAGE_SIZE": func(value string, config *DynamicConfiguration) {
		maxMessageSize, err := strconv.Atoi(value)
		if err != nil {
			return
		}

		config.AgentComWsConfiguration.SizeLimitedChannelConfiguration.MaxMessageSize = maxMessageSize
		config.AgentComWsConfiguration.SizeLimitedChannelConfiguration.MaxQueueLength = maxMessageSize * 10
	},
	"GOLANG_MAX_AUG_TIME_MULTIPLIER": func(value string, config *DynamicConfiguration) {
		maxAugTimeMultiplier, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return
		}

		if maxAugTimeMultiplier > 2 || maxAugTimeMultiplier < 1 {
			
			return
		}

		config.LocationsConfiguration.MaxAugTimeMultiplier = maxAugTimeMultiplier
	},
	"GOLANG_GLOBAL_RATE_LIMIT_QUOTA_MS": func(value string, config *DynamicConfiguration) {
		if config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimit == "" {
			config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitQuotaMS = value

			if config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS != "" {
				config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimit = fmt.Sprintf("%s/%s",
					config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitQuotaMS,
					config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS)
			}
		}
	},
	"GOLANG_GLOBAL_RATE_LIMIT_WINDOW_SIZE_MS": func(value string, config *DynamicConfiguration) {
		if config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimit == "" {
			config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS = value

			if config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitQuotaMS != "" {
				config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimit = fmt.Sprintf("%s/%s",
					config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitQuotaMS,
					config.LocationsConfiguration.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS)
			}
		}
	},
}

func ParseConfig(configMap map[string]string) DynamicConfiguration {
	config := GetDefaultConfiguration()

	for key, f := range configParsers {
		if value, ok := configMap[key]; ok {
			f(value, &config)
		}
	}

	return config
}
