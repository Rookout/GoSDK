package pkg

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
}
