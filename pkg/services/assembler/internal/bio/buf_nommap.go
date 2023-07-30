// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !(solaris && go1.20)
// +build !aix
// +build !darwin
// +build !dragonfly
// +build !freebsd
// +build !linux
// +build !netbsd
// +build !openbsd
// +build !solaris !go1.20

package bio

func (r *Reader) sliceOS(length uint64) ([]byte, bool) {
	return nil, false
}
