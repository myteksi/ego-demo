// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"io"
	"unsafe"

	"github.com/grab/ego/ego/src/go/volatile"
)

// bufferInstance implements envoy.BufferInstance
//
type bufferInstance struct {
	ptr unsafe.Pointer
}

func (b bufferInstance) CopyOut(start uint64, dest []byte) int {
	return int(C.BufferInstance_copyOut(b.ptr, C.size_t(start), GoBuf(dest)))
}

func (b bufferInstance) GetRawSlices() []volatile.Bytes {

	// FIXME: since this is not thread safe, we could just buffer the result
	//        for BufferInstance_getRawSlices to pick it up without having to
	//        call getRawSlices a second time...
	max := int(C.BufferInstance_getRawSlicesCount(b.ptr))
	if 0 == max {
		return nil
	}

	temp := make([]C.GoBuf, max)
	count := int(C.BufferInstance_getRawSlices(b.ptr, C.uint64_t(max), &temp[0]))
	dest := make([]volatile.Bytes, count)
	for i := 0; i < count; i++ {
		dest[i] = CBytes(temp[i].data, temp[i].len, temp[i].cap)
	}
	return dest
}

func (b bufferInstance) Length() uint64 {
	return uint64(C.BufferInstance_length(b.ptr))
}

func (b bufferInstance) NewReader(start uint64) io.Reader {
	return &bufferInstanceReader{bufferInstance: b, pos: start}
}

type bufferInstanceReader struct {
	bufferInstance
	pos uint64
}

func (r *bufferInstanceReader) Read(p []byte) (n int, err error) {
	if 0 < len(p) {
		if n = r.CopyOut(r.pos, p); n <= 0 {
			err = io.EOF
		} else {
			r.pos += uint64(n)
		}
	}
	return
}
