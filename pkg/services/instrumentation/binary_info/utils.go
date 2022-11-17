package binary_info

import (
	"fmt"
	"runtime"
)

func uniq(s []string) []string {
	if len(s) <= 0 {
		return s
	}
	src, dst := 1, 1
	for src < len(s) {
		if s[src] != s[dst-1] {
			s[dst] = s[src]
			dst++
		}
		src++
	}
	return s[:dst]
}


func GoVersionAfterOrEqual(major int, minor int) bool {
	var realMajor, realMinor int
	n, err := fmt.Sscanf(runtime.Version(), "go%d.%d", &realMajor, &realMinor)
	if err != nil || n != 2 {
		return false
	}

	return realMajor > major || (realMajor == major && realMinor >= minor)
}

func complexType(typename string) bool {
	for _, ch := range typename {
		switch ch {
		case '*', '[', '<', '{', '(', ' ':
			return true
		}
	}
	return false
}
