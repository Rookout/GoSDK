// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package abi



const (
	MapBucketCountBits = 3 
	MapBucketCount     = 1 << MapBucketCountBits
	MapMaxKeyBytes     = 128 
	MapMaxElemBytes    = 128 
)
