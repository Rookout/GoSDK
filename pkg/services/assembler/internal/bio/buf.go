// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bio implements common I/O abstractions used within the Go toolchain.
package bio

import (
	"bufio"
	"io"
	"log"
	"os"
)


type Reader struct {
	f *os.File
	*bufio.Reader
}


type Writer struct {
	f *os.File
	*bufio.Writer
}



func Create(name string) (*Writer, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &Writer{f: f, Writer: bufio.NewWriter(f)}, nil
}


func Open(name string) (*Reader, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return NewReader(f), nil
}


func NewReader(f *os.File) *Reader {
	return &Reader{f: f, Reader: bufio.NewReader(f)}
}

func (r *Reader) MustSeek(offset int64, whence int) int64 {
	if whence == 1 {
		offset -= int64(r.Buffered())
	}
	off, err := r.f.Seek(offset, whence)
	if err != nil {
		log.Fatalf("seeking in output: %v", err)
	}
	r.Reset(r.f)
	return off
}

func (w *Writer) MustSeek(offset int64, whence int) int64 {
	if err := w.Flush(); err != nil {
		log.Fatalf("writing output: %v", err)
	}
	off, err := w.f.Seek(offset, whence)
	if err != nil {
		log.Fatalf("seeking in output: %v", err)
	}
	return off
}

func (r *Reader) Offset() int64 {
	off, err := r.f.Seek(0, 1)
	if err != nil {
		log.Fatalf("seeking in output [0, 1]: %v", err)
	}
	off -= int64(r.Buffered())
	return off
}

func (w *Writer) Offset() int64 {
	if err := w.Flush(); err != nil {
		log.Fatalf("writing output: %v", err)
	}
	off, err := w.f.Seek(0, 1)
	if err != nil {
		log.Fatalf("seeking in output [0, 1]: %v", err)
	}
	return off
}

func (r *Reader) Close() error {
	return r.f.Close()
}

func (w *Writer) Close() error {
	err := w.Flush()
	err1 := w.f.Close()
	if err == nil {
		err = err1
	}
	return err
}

func (r *Reader) File() *os.File {
	return r.f
}

func (w *Writer) File() *os.File {
	return w.f
}






func (r *Reader) Slice(length uint64) ([]byte, bool, error) {
	if length == 0 {
		return []byte{}, false, nil
	}

	data, ok := r.sliceOS(length)
	if ok {
		return data, true, nil
	}

	data = make([]byte, length)
	_, err := io.ReadFull(r, data)
	if err != nil {
		return nil, false, err
	}
	return data, false, nil
}





func (r *Reader) SliceRO(length uint64) []byte {
	data, ok := r.sliceOS(length)
	if ok {
		return data
	}
	return nil
}
