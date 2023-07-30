// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package objabi

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)




func WorkingDir() string {
	var path string
	path, _ = os.Getwd()
	if path == "" {
		path = "/???"
	}
	return filepath.ToSlash(path)
}










func AbsFile(dir, file, rewrites string) string {
	abs := file
	if dir != "" && !filepath.IsAbs(file) {
		abs = filepath.Join(dir, file)
	}

	abs, rewritten := ApplyRewrites(abs, rewrites)
	if !rewritten && buildcfg.GOROOT != "" && hasPathPrefix(abs, buildcfg.GOROOT) {
		abs = "$GOROOT" + abs[len(buildcfg.GOROOT):]
	}

	
	
	
	if runtime.GOOS == "windows" {
		abs = strings.ReplaceAll(abs, `\`, "/")
	}

	if abs == "" {
		abs = "??"
	}
	return abs
}








func ApplyRewrites(file, rewrites string) (string, bool) {
	start := 0
	for i := 0; i <= len(rewrites); i++ {
		if i == len(rewrites) || rewrites[i] == ';' {
			if new, ok := applyRewrite(file, rewrites[start:i]); ok {
				return new, true
			}
			start = i + 1
		}
	}

	return file, false
}




func applyRewrite(path, rewrite string) (string, bool) {
	prefix, replace := rewrite, ""
	if j := strings.LastIndex(rewrite, "=>"); j >= 0 {
		prefix, replace = rewrite[:j], rewrite[j+len("=>"):]
	}

	if prefix == "" || !hasPathPrefix(path, prefix) {
		return path, false
	}
	if len(path) == len(prefix) {
		return replace, true
	}
	if replace == "" {
		return path[len(prefix)+1:], true
	}
	return replace + path[len(prefix):], true
}








func hasPathPrefix(s string, t string) bool {
	if len(t) > len(s) {
		return false
	}
	var i int
	for i = 0; i < len(t); i++ {
		cs := int(s[i])
		ct := int(t[i])
		if 'A' <= cs && cs <= 'Z' {
			cs += 'a' - 'A'
		}
		if 'A' <= ct && ct <= 'Z' {
			ct += 'a' - 'A'
		}
		if cs == '\\' {
			cs = '/'
		}
		if ct == '\\' {
			ct = '/'
		}
		if cs != ct {
			return false
		}
	}
	return i >= len(s) || s[i] == '/' || s[i] == '\\'
}
