package config

type RookOptions struct {
	Token       string
	Host        string
	Port        int
	Proxy       string
	Debug       bool
	LogLevel    string
	LogToStderr bool
	LogToFile   bool
	LogFile     string
	GitCommit   string
	GitOrigin   string
	LiveTail    bool
	Labels      map[string]string
	Quiet       bool
}

func UpdateFromOpts(opts RookOptions) {
	UpdateConfig(func(config *DynamicConfiguration) {
		if opts.LogLevel != "" {
			config.LoggingConfiguration.LogLevel = opts.LogLevel
		}

		if opts.LogFile != "" {
			config.LoggingConfiguration.FileName = opts.LogFile
		}

		config.LoggingConfiguration.LogToStderr = opts.LogToStderr

		if opts.Debug {
			config.LoggingConfiguration.Debug = true
			config.LoggingConfiguration.LogLevel = "DEBUG"
			config.LoggingConfiguration.LogToStderr = true
		}

		if opts.Quiet {
			config.LoggingConfiguration.Quiet = opts.Quiet
		}
	})
}
