// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

// This file defines flags attached to various functions
// and data objects. The compilers, assemblers, and linker must
// all agree on these values.

package obj

const (
	
	
	
	NOPROF = 1

	
	
	DUPOK = 2

	
	NOSPLIT = 4

	
	RODATA = 8

	
	NOPTR = 16

	
	
	WRAPPER = 32

	
	NEEDCTXT = 64

	
	LOCAL = 128

	
	
	TLSBSS = 256

	
	
	
	NOFRAME = 512

	
	REFLECTMETHOD = 1024

	
	
	TOPFRAME = 2048

	
	ABIWRAPPER = 4096

	
	PKGINIT = 8192
)
