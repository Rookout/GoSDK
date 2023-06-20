// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bio

import (
	"io"
	"log"
)


func MustClose(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Fatal(err)
	}
}



func MustWriter(w io.Writer) io.Writer {
	return mustWriter{w}
}

type mustWriter struct {
	w io.Writer
}

func (w mustWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		log.Fatal(err)
	}
	return n, nil
}

func (w mustWriter) WriteString(s string) (int, error) {
	n, err := io.WriteString(w.w, s)
	if err != nil {
		log.Fatal(err)
	}
	return n, nil
}
