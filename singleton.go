package rookout

import (
	"fmt"
)

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
		memberToString("live_tail", opts.LiveTail) +
		memberToString("labels", opts.Labels)

	println(s)
}

func start(opts RookOptions) error {
	printOptions(&opts)
	println("Local stub rook, For Staging/Production you should use different module")
	return nil
}

func stop() error {
	return nil
}
