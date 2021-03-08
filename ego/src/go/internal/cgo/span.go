// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"

import (
	"unsafe"

	"github.com/golang/protobuf/proto"

	pb "github.com/grab/ego/ego/src/cc/goc/proto"
	"github.com/grab/ego/ego/src/go/envoy"
)

type span struct {
	filter unsafe.Pointer
	spanID int64
}

func (s span) GetContext() map[string][]string {
	defaultBufferSize := uint64(100)
	data := make([]byte, defaultBufferSize)
	size := uint64(C.GoHttpFilter_Span_getContext(s.filter, C.long(s.spanID), GoBuf(data)))
	if size > defaultBufferSize {
		newBufferSize := size
		data = make([]byte, newBufferSize)
		size = uint64(C.GoHttpFilter_Span_getContext(s.filter, C.long(s.spanID), GoBuf(data)))
		if size > newBufferSize {
			// it's not supposed to happen.
			panic("Invalid logic")
		}
	}
	headers := pb.RequestHeaderMap{}
	if err := proto.Unmarshal(data[:size], &headers); err != nil {
		// TODO: handle error
	}

	result := make(map[string][]string)
	for _, h := range headers.Headers {
		if _, existing := result[h.Key]; !existing {
			result[h.Key] = []string{h.Value}
		}
	}

	return result
}

func (s span) SpawnChild(name string) envoy.Span {
	spanID := C.GoHttpFilter_Span_spawnChild(s.filter, C.long(s.spanID), GoStr(name))
	return span{filter: s.filter, spanID: int64(spanID)}
}

func (s span) FinishSpan() {
	C.GoHttpFilter_Span_finishSpan(s.filter, C.long(s.spanID))
}
