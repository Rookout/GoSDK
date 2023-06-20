// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objabi

import (
	"fmt"
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/buildcfg"
)

const (
	ElfRelocOffset   = 256
	MachoRelocOffset = 2048    
	GlobalDictPrefix = ".dict" 
)





func HeaderString() string {
	archExtra := ""
	if k, v := buildcfg.GOGOARCH(); k != "" && v != "" {
		archExtra = " " + k + "=" + v
	}
	return fmt.Sprintf("go object %s %s %s%s X:%s\n",
		buildcfg.GOOS, buildcfg.GOARCH,
		buildcfg.Version, archExtra,
		strings.Join(buildcfg.Experiment.Enabled(), ","))
}
