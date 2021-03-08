// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"unsafe"

	"github.com/grab/ego/ego/src/go/envoy"
)

type encoderCallbacks struct {
	filter unsafe.Pointer
}

// This section is implementation downcalls for StreamEncoderFilterCallbacks
//
// See //envoy/include/envoy/http/filter.h
func (c encoderCallbacks) EncodingBuffer() envoy.BufferInstance {
	ptr := C.GoHttpFilter_EncoderCallbacks_encodingBuffer(c.filter)
	if ptr == nil {
		return nil
	}

	return bufferInstance{ptr}
}

// See //envoy/include/envoy/http/filter.h
func (c encoderCallbacks) AddEncodedData(buffer envoy.BufferInstance, streamingFilter bool) {
	if buffer == nil {
		return
	}

	var streamingFilterInt int8
	if streamingFilter {
		streamingFilterInt = 1
	}

	b := buffer.(bufferInstance)
	C.GoHttpFilter_EncoderCallbacks_addEncodedData(c.filter, b.ptr, C.int(streamingFilterInt))
}

// See //envoy/include/envoy/http/filter.h
func (c encoderCallbacks) ContinueEncoding() {
	C.GoHttpFilter_EncoderCallbacks_continueEncoding(c.filter)
}

func (c encoderCallbacks) StreamInfo() envoy.StreamInfo {
	return &streamInfo{c.filter, true}
}

func (c encoderCallbacks) RouteExisting() bool {
	existing := C.GoHttpFilter_StreamFilterCallbacks_routeExisting(c.filter, GoBool(true))
	return existing != 0
}

func (c encoderCallbacks) Route() envoy.Route {
	return &route{c.filter, true}
}

func (c encoderCallbacks) ActiveSpan() envoy.Span {
	return span{filter: c.filter, spanID: 0}
}
