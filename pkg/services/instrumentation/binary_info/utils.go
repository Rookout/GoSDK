// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
