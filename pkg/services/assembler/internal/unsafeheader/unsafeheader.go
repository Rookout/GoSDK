// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

// Package unsafeheader contains header declarations for the Go runtime's slice
// and string implementations.
//
// This package allows packages that cannot import "reflect" to use types that
// are tested to be equivalent to reflect.SliceHeader and reflect.StringHeader.
package unsafeheader

import (
	"unsafe"
)







type Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}







type String struct {
	Data unsafe.Pointer
	Len  int
}
