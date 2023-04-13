package logger

import (
	"fmt"
	"runtime"

	"github.com/Rookout/GoSDK/pkg/config"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/sirupsen/logrus"

	logrus_lumberjack "github.com/fallais/logrus-lumberjack-hook"
	"gopkg.in/natefinch/lumberjack.v2"

	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
)

func QuietPrintln(msg string) {
	if !config.LoggingConfig().Quiet {
		fmt.Println(msg)
	}
}

type LoggerOutput interface {
	SendLogMessage(level pb.LogMessage_LogLevel, time time.Time, filename string, lineno int, text string, args map[string]interface{}) error
}

type nilWriter struct{}

func (*nilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func getNilLogger(debug bool, level string) *logrus.Logger {
	l := logrus.New()
	if debug {
		l.SetOutput(&nilWriter{})
		l.Level = logrus.DebugLevel
	} else {
		l.Level, _ = logrus.ParseLevel(level)
	}

	l.SetReportCaller(true)
	l.ExitFunc = func(int) {}
	return l
}

var logger *logrus.Logger = nil
var isDebug = false

func Init(debug bool, level string) {
	isDebug = debug
	logger = getNilLogger(debug, level)
}

func Logger() *logrus.Logger {
	if logger == nil {
		Init(isDebug, "debug")
	}
	return logger
}

func NewFileHandler(file string) error {
	if file == "" {
		file = filepath.Join(getLogFolder(), "go-rook.log")
	}

	ljk := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    100, 
		MaxBackups: 5,   
		MaxAge:     28,  
	}
	if hook, err := logrus_lumberjack.NewLumberjackHook(ljk); err != nil {
		return err
	} else {
		logger.AddHook(hook)
	}

	return nil
}

func getLogFolder() string {
	switch runtime.GOOS {
	case "darwin":
		return os.Getenv("HOME")
	case "windows":
		return os.Getenv("USERPROFILE")
	default:
		return "/tmp/log/rookout"
	}
}

type outputHook struct {
	output      LoggerOutput
	levels      []logrus.Level
	rateLimiter *rate.Limiter
}

func InitHandlers(logToStderr, logToFile bool, logFile string) {
	if logToStderr {
		logger.SetOutput(os.Stderr)
	} else if isDebug {
		logger.SetOutput(os.Stdout)
	} else {
		logger.SetOutput(&nilWriter{})
	}

	if logToFile {
		_ = NewFileHandler(logFile)
	}
}

func newOutputHook(output LoggerOutput) *outputHook {
	hook := &outputHook{output: output, rateLimiter: rate.NewLimiter(3*rate.Every(time.Second), 10)}
	if isDebug {
		hook.levels = logrus.AllLevels
		return hook
	}

	actualLvls := make([]logrus.Level, 0)
	for _, level := range logrus.AllLevels {
		if level <= logrus.DebugLevel {
			actualLvls = append(actualLvls, level)
		}
	}

	hook.levels = actualLvls
	return hook
}

func (o *outputHook) Levels() []logrus.Level {
	return o.levels
}

func LevelToInt(level logrus.Level) pb.LogMessage_LogLevel {
	switch level {
	case logrus.TraceLevel:
		return pb.LogMessage_TRACE
	case logrus.DebugLevel:
		return pb.LogMessage_DEBUG
	case logrus.InfoLevel:
		return pb.LogMessage_INFO
	case logrus.WarnLevel:
		return pb.LogMessage_WARNING
	case logrus.ErrorLevel:
		return pb.LogMessage_ERROR
	case logrus.PanicLevel, logrus.FatalLevel:
		return pb.LogMessage_FATAL
	default:
		return pb.LogMessage_LogLevel(level)
	}
}

func (o *outputHook) Fire(e *logrus.Entry) error {
	var filename string
	var lineno int
	if e.HasCaller() {
		filename = e.Caller.File
		lineno = e.Caller.Line
	}

	
	_ = o.output.SendLogMessage(LevelToInt(e.Level), e.Time, filename, lineno, e.Message, e.Data)
	return nil
}

func SetLoggerOutput(output LoggerOutput) {
	logger.AddHook(newOutputHook(output))
}
