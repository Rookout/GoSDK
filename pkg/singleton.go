package pkg

import (
	"github.com/Rookout/GoSDK/pkg/aug_manager"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/information"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation"
	"github.com/Rookout/GoSDK/pkg/utils"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

type singleton struct {
	config          config.DynamicConfiguration
	output          com_ws.Output
	agentCom        com_ws.AgentCom
	commandHandler  *aug_manager.CommandHandler
	augManager      aug_manager.AugManager
	triggerServices *instrumentation.TriggerServices

	opts *RookOptions

	started         bool
	servicesStarted bool
}

var initOnce sync.Once
var rookSingleton *singleton

func GetSingleton() *singleton {
	if rookSingleton == nil {
		InitSingleton()
	}

	return rookSingleton
}


func InitSingleton() {
	initOnce.Do(func() {
		initializedSingleton := createSingleton()
		rookSingleton = initializedSingleton
	})
}

func createSingleton() *singleton {
	return &singleton{
		servicesStarted: false,
	}
}

func initOptsFromEnv(opts *RookOptions) (err error) {
	if !opts.Debug {
		rookoutDebug, _ := os.LookupEnv("ROOKOUT_DEBUG")
		opts.Debug = utils.Contains(utils.TrueValues, rookoutDebug)
	}

	if !opts.LogToStderr {
		logToStderr, _ := os.LookupEnv("ROOKOUT_LOG_TO_STDERR")
		opts.LogToStderr = utils.Contains(utils.TrueValues, logToStderr)
	}

	if !opts.LogToFile {
		logToFile, _ := os.LookupEnv("ROOKOUT_LOG_TO_FILE")
		opts.LogToFile = utils.Contains(utils.TrueValues, logToFile)
	}

	if opts.LogFile == "" {
		opts.LogFile, _ = os.LookupEnv("ROOKOUT_LOG_FILE")
	}

	if opts.LogLevel == "" {
		opts.LogLevel, _ = os.LookupEnv("ROOKOUT_LOG_LEVEL")
	}

	if opts.Token == "" {
		opts.Token, _ = os.LookupEnv("ROOKOUT_TOKEN")
	}

	if opts.Host == "" {
		opts.Host, _ = os.LookupEnv("ROOKOUT_CONTROLLER_HOST")
	}

	if opts.GitOrigin == "" {
		opts.GitOrigin, _ = os.LookupEnv("ROOKOUT_REMOTE_ORIGIN")
	}
	information.GitConfig.RemoteOrigin = opts.GitOrigin

	if opts.GitCommit == "" {
		opts.GitCommit, _ = os.LookupEnv("ROOKOUT_COMMIT")
	}
	information.GitConfig.Commit = opts.GitCommit

	if !opts.LiveTail {
		liveTail, _ := os.LookupEnv("ROOKOUT_LIVE_LOGGER")
		opts.LiveTail = utils.Contains(utils.TrueValues, liveTail)
	}

	if opts.Proxy == "" {
		opts.Proxy, _ = os.LookupEnv("ROOKOUT_PROXY")
	}

	if !opts.Quiet {
		quiet, _ := os.LookupEnv("ROOKOUT_QUIET")
		opts.Quiet = utils.Contains(utils.TrueValues, quiet)
	}

	skipSslVerify, ok := os.LookupEnv("ROOKOUT_SKIP_SSL_VERIFY")
	opts.skipSslVerify = ok && utils.Contains(utils.TrueValues, skipSslVerify)

	if opts.Port == 0 {
		if port, ok := os.LookupEnv("ROOKOUT_CONTROLLER_PORT"); ok {
			if p, ok := strconv.Atoi(port); ok == nil {
				opts.Port = p
			}
		}
	}

	if len(opts.Labels) == 0 {
		if opts.Labels, err = getLabelsFromEnv(opts.Labels); err != nil {
			return err
		}
	}

	return
}


func normalizeOpts(opts *RookOptions) (err error) {
	Sanitize(opts)
	if opts.Token == "" && opts.Host == "" {
		return rookoutErrors.NewRookMissingToken()
	} else if opts.Token != "" {
		if err = validateToken(opts.Token); err != nil {
			return err
		}
	}

	if opts.Host == "" {
		opts.Host = ControllerAddress_HOST
	}

	if opts.Host == "staging.cloud.agent.rookout.com" || opts.Host == "cloud.agent.rookout.com" {
		opts.Host = "https://" + opts.Host
	}

	if opts.Host == "staging.control.rookout.com" || opts.Host == "control.rookout.com" {
		opts.Host = "wss://" + opts.Host
	}

	if opts.Port == 0 {
		opts.Port = ControllerAddress_PORT
	}

	if opts.LogLevel == "" {
		opts.LogLevel = "info"
	}

	for key := range opts.Labels {
		if err = validateLabel(key); err != nil {
			return err
		}
	}

	if opts.Debug {
		opts.LogToFile = true
		opts.LogToStderr = true
	}

	return
}

func (s *singleton) updateConfig() {
	if s.opts.LogLevel != "" {
		s.config.LoggingConfiguration.LogLevel = s.opts.LogLevel
	}

	if s.opts.LogFile != "" {
		s.config.LoggingConfiguration.FileName = s.opts.LogFile
	}

	s.config.LoggingConfiguration.LogToStderr = s.opts.LogToStderr

	if s.opts.Debug {
		s.config.LoggingConfiguration.Debug = true
		s.config.LoggingConfiguration.LogLevel = "DEBUG"
		s.config.LoggingConfiguration.LogToStderr = true
	}

	if s.opts.Quiet {
		s.config.LoggingConfiguration.Quiet = s.opts.Quiet
	}
}

func (s *singleton) Start(opts *RookOptions) (err error) {
	if s.started {
		return
	}

	s.config = config.GetDefaultConfiguration()
	config.UpdateObjectDumpConfigDefaults(s.config.ObjectDumpConfigDefaults)
	s.opts = opts

	s.started = true

	if err = initOptsFromEnv(s.opts); err != nil {
		return err
	}
	if err = normalizeOpts(s.opts); err != nil {
		return err
	}

	s.updateConfig()

	logger.Init(s.opts.Debug, s.opts.LogLevel, s.config.LoggingConfiguration)
	logger.InitHandlers(s.opts.LogToStderr, s.opts.LogToFile, s.opts.LogFile)
	utils.SetOnPanicFunc(func(err error) {
		logger.Logger().WithError(err).Fatalf("Caught panic in goroutine, stack trace: %s\n", string(debug.Stack()))
	})

	s.triggerServices, err = instrumentation.NewTriggerServices()
	if err != nil {
		return
	}

	output := com_ws.NewOutputWs(s.config.OutputWsConfiguration)
	s.output = output
	logger.SetLoggerOutput(output)
	utils.SetQuiet(s.config.LoggingConfiguration.Quiet)

	return s.connect()
}

func (s *singleton) Stop() {
	if !s.started {
		return
	}

	s.triggerServices.Close()
	return
}

func (s *singleton) Flush() {
	if !s.started || s.agentCom == nil {
		return
	}

	s.agentCom.Flush()
}

func (s *singleton) connect() (err error) {
	agentCom, err := com_ws.NewAgentComWs(
		com_ws.NewWebSocketClient,
		s.output,
		com_ws.NewBackoff(s.config.AgentComWsConfiguration.BackoffConfig),
		s.opts.Host,
		s.opts.Port,
		s.opts.Proxy,
		s.opts.Token,
		s.opts.Labels,
		true,
		s.opts.skipSslVerify,
		s.config.AgentComWsConfiguration)
	if err != nil {
		return err
	}

	s.output.SetAgentCom(agentCom)
	s.agentCom = agentCom
	s.augManager = aug_manager.NewAugManager(s.triggerServices, s.output, s.config.LocationsConfiguration)
	s.commandHandler = aug_manager.NewCommandHandler(s.agentCom, s.augManager)

	return agentCom.ConnectToAgent()
}

func (s *singleton) startServices() (err error) {
	s.triggerServices, err = instrumentation.NewTriggerServices()
	return err
}

func getLabelsFromEnv(labels map[string]string) (map[string]string, error) {
	if len(labels) == 0 {
		if labelsEnvVar, ok := os.LookupEnv("ROOKOUT_LABELS"); ok {
			labels = make(map[string]string)

			labelsPairs := strings.Split(labelsEnvVar, ",")
			for _, pair := range labelsPairs {
				k := strings.Split(pair, ":")
				if len(k) == 2 {
					if err := validateLabel(k[0]); err != nil {
						return nil, rookoutErrors.NewInvalidLabelError(k[0])
					}
					labels[k[0]] = k[1]
				}
			}
		}
	}

	return labels, nil
}

func validateToken(token string) error {
	if len(token) != 64 {
		return rookoutErrors.NewRookInvalidOptions("Rookout token should be 64 characters")
	}

	res, e := regexp.MatchString("^[0-9a-zA-Z]+$", token)
	if e != nil {
		return rookoutErrors.NewRuntimeError(e.Error())
	}

	if !res {
		return rookoutErrors.NewRookInvalidOptions("Rookout token must consist of only hexadecimal characters")
	}

	return nil
}

func validateLabel(label string) error {
	if strings.HasPrefix(label, "$") {
		return rookoutErrors.NewInvalidLabelError(label)
	}
	return nil
}