// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86



const argListMax int = 6

type argList [argListMax]uint8

type ytab struct {
	zcase   uint8
	zoffset uint8

	
	
	
	args argList
}







func (yt *ytab) match(args []int) bool {
	
	
	
	if len(args) < len(yt.args) && yt.args[len(args)] != Yxxx {
		return false
	}

	for i := range args {
		if ycover[args[i]+int(yt.args[i])] == 0 {
			return false
		}
	}

	return true
}
