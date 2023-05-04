package rookout

import (
	"fmt"
	"os"

	"github.com/Rookout/GoSDK/pkg"
	"github.com/Rookout/GoSDK/pkg/information"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

//go:generate go generate ./trampoline/

func memberToString(memberName string, member interface{}) string {
	if member != nil {
		if member == "" {
			return fmt.Sprintf("%s:'' ,", memberName)
		}
		return fmt.Sprintf("%s:%v ,", memberName, member)
	}
	return ""
}

func printOptions(opts *RookOptions) {
	censoredToken := ""
	if len(opts.Token) > 5 {
		censoredToken = opts.Token[:5]
	}

	s := "RookOptions: " +
		memberToString("token", censoredToken) +
		memberToString("host", opts.Host) +
		memberToString("port", opts.Port) +
		memberToString("proxy", opts.Proxy) +
		memberToString("log_level", opts.LogLevel) +
		memberToString("log_to_stderr", opts.LogToStderr) +
		memberToString("log_file", opts.LogFile) +
		memberToString("git_commit", opts.GitCommit) +
		memberToString("git_origin", opts.GitOrigin) +
		memberToString("git_sources", opts.GitSources) +
		memberToString("live_tail", opts.LiveTail) +
		memberToString("labels", opts.Labels)

	println(s)
}

func start(opts RookOptions) {
	pkg.InitSingleton()
	obj := pkg.GetSingleton()

	err := obj.Start(&opts)
	if opts.Debug {
		logger.Logger().Debug("Rookout SDK for Go, Version: " + information.VERSION)
		printOptions(&opts)
	}
	if err != nil {
		logger.Logger().WithError(err).Errorln("Failed to start rook")
		if rookErr, ok := err.(rookoutErrors.RookoutError); ok {
			switch {
			case isErrorType(rookErr, rookoutErrors.NewRookInvalidOptions("")),
				isErrorType(rookErr, rookoutErrors.NewInvalidTokenError()),
				isErrorType(rookErr, rookoutErrors.NewRookMissingToken()),
				isErrorType(rookErr, rookoutErrors.NewInvalidLabelError("")),
				isErrorType(rookErr, rookoutErrors.NewWebSocketError()):
				_, _ = fmt.Fprintf(os.Stderr, "[Rookout] Failed to start rookout: %v\n", err)
			default:
				_, _ = fmt.Fprintf(os.Stderr, "[Rookout] Failed to connect to the controller - will continue attempting in the background: %v\n", err)
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "[Rookout] Failed to start rookout: %v\n", err)
		}
	}
}

func isErrorType(err rookoutErrors.RookoutError, errType rookoutErrors.RookoutError) bool {
	return err.GetType() == errType.GetType()
}

func stop() {
	obj := pkg.GetSingleton()
	obj.Stop()
}

func flush() {
	obj := pkg.GetSingleton()
	obj.Flush()
}
