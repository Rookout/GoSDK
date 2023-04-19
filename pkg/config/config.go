package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Rookout/GoSDK/pkg/utils"
)

var configLock sync.Mutex
var config atomic.Value

func init() {
	config.Store(GetDefaultConfiguration())
}

func AgentComWsConfig() AgentComWsConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.AgentComWsConfiguration
}
func LocationsConfig() LocationsConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.LocationsConfiguration
}
func LoggingConfig() LoggingConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.LoggingConfiguration
}
func OutputWsConfig() OutputWsConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.OutputWsConfiguration
}
func RateLimiterConfig() RateLimiterConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.RateLimiterConfiguration
}
func BackoffConfig() BackoffConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.BackoffConfiguration
}
func WebSocketClientConfig() WebSocketClientConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.WebSocketClientConfiguration
}
func SizeLimitedChannelConfig() SizeLimitedChannelConfiguration {
	c := config.Load().(DynamicConfiguration)
	return c.SizeLimitedChannelConfiguration
}

type LocationsConfiguration struct {
	MaxAugTime           time.Duration
	MaxAugTimeMultiplier float64
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
	PingTimeout       time.Duration
	ConnectTimeout    time.Duration
	ConnectionTimeout time.Duration
}

type BackoffConfiguration struct {
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
	ProtobufVersion2  bool
}

type WebSocketClientConfiguration struct {
	PingTimeout   time.Duration
	PingInterval  time.Duration
	WriteTimeout  time.Duration
	SkipSSLVerify bool
}

type DynamicConfiguration struct {
	AgentComWsConfiguration         AgentComWsConfiguration
	LocationsConfiguration          LocationsConfiguration
	LoggingConfiguration            LoggingConfiguration
	OutputWsConfiguration           OutputWsConfiguration
	RateLimiterConfiguration        RateLimiterConfiguration
	BackoffConfiguration            BackoffConfiguration
	WebSocketClientConfiguration    WebSocketClientConfiguration
	SizeLimitedChannelConfiguration SizeLimitedChannelConfiguration

	ObjectDumpConfigDefaults ObjectDumpConfigDefaults
	onUpdate                 []func()
}

func GetDefaultConfiguration() DynamicConfiguration {
	return DynamicConfiguration{
		LocationsConfiguration: LocationsConfiguration{
			MaxAugTime:           400 * time.Millisecond,
			MaxAugTimeMultiplier: 1,
		},
		RateLimiterConfiguration: RateLimiterConfiguration{
			MinRateLimitValue:           20000 * time.Nanosecond,
			GlobalRateLimit:             os.Getenv("ROOKOUT_GLOBAL_RATE_LIMIT"),
			GlobalRateLimitQuotaMS:      "",
			GlobalRateLimitWindowSizeMS: "",
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
		},
		SizeLimitedChannelConfiguration: SizeLimitedChannelConfiguration{
			FlushTimeout:      2 * time.Second,
			MaxQueueLength:    250,
			MaxBytesInChannel: 15 * 1024 * 1024,
			MaxMessageSize:    1024 * 1024,
		},
		BackoffConfiguration: BackoffConfiguration{
			DefaultBackoff:      200 * time.Millisecond,
			MaxBackoff:          60 * time.Second,
			ResetBackoffTimeout: 3 * time.Minute,
		},
		WebSocketClientConfiguration: WebSocketClientConfiguration{
			PingTimeout:   30 * time.Second,
			PingInterval:  10 * time.Second,
			WriteTimeout:  5 * time.Second,
			SkipSSLVerify: utils.IsTrue(os.Getenv("ROOKOUT_SKIP_SSL_VERIFY")),
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
			ProtobufVersion2:  utils.IsTrue(os.Getenv("ROOKOUT_Protobuf_Version2")),
		},
	}
}

func updateGlobalRateLimitConfig(newConfig *DynamicConfiguration) {
	if os.Getenv("ROOKOUT_GLOBAL_RATE_LIMIT") != "" {
		return
	}

	config := RateLimiterConfig()
	if config.GlobalRateLimit == "" {
		if config.GlobalRateLimitQuotaMS != "" &&
			config.GlobalRateLimitWindowSizeMS != "" {
			newConfig.RateLimiterConfiguration.GlobalRateLimit = fmt.Sprintf("%s/%s",
				config.GlobalRateLimitQuotaMS,
				config.GlobalRateLimitWindowSizeMS,
			)
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

		config.SizeLimitedChannelConfiguration.MaxMessageSize = maxMessageSize
		config.SizeLimitedChannelConfiguration.MaxQueueLength = maxMessageSize * 10
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
		if config.RateLimiterConfiguration.GlobalRateLimit == "" {
			config.RateLimiterConfiguration.GlobalRateLimitQuotaMS = value

			if config.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS != "" {
				config.RateLimiterConfiguration.GlobalRateLimit = fmt.Sprintf("%s/%s",
					config.RateLimiterConfiguration.GlobalRateLimitQuotaMS,
					config.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS)
			}
		}
	},
	"GOLANG_GLOBAL_RATE_LIMIT_WINDOW_SIZE_MS": func(value string, config *DynamicConfiguration) {
		if config.RateLimiterConfiguration.GlobalRateLimit == "" {
			config.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS = value

			if config.RateLimiterConfiguration.GlobalRateLimitQuotaMS != "" {
				config.RateLimiterConfiguration.GlobalRateLimit = fmt.Sprintf("%s/%s",
					config.RateLimiterConfiguration.GlobalRateLimitQuotaMS,
					config.RateLimiterConfiguration.GlobalRateLimitWindowSizeMS)
			}
		}
	},
	"GOLANG_PROTOBUF_VERSION_2": func(value string, config *DynamicConfiguration) {
		if value == "" {
			return
		}

		config.OutputWsConfiguration.ProtobufVersion2 = config.OutputWsConfiguration.ProtobufVersion2 || utils.Contains(utils.TrueValues, value)
	},
}



func UpdateConfig(update func(config *DynamicConfiguration)) {
	configLock.Lock()
	defer configLock.Unlock()

	c := config.Load().(DynamicConfiguration)
	update(&c)
	config.Store(c)

	for _, f := range c.onUpdate {
		f()
	}
}

func Update(configMap map[string]string) {
	UpdateConfig(func(config *DynamicConfiguration) {
		for key, f := range configParsers {
			if value, ok := configMap[key]; ok {
				f(value, config)
			}
		}
		updateGlobalRateLimitConfig(config)
	})
}

func OnUpdate(f func()) {
	UpdateConfig(func(config *DynamicConfiguration) {
		config.onUpdate = append(config.onUpdate, f)
	})
}
