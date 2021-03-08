// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"unsafe"

	pb "github.com/grab/ego/ego/src/cc/goc/proto"
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/envoy/loglevel"

	"github.com/golang/protobuf/proto"
)

type decoderCallbacks struct {
	filter unsafe.Pointer
}

// This section is implementation downcalls for StreamDecoderFilterCallbacks

func (c decoderCallbacks) DecodingBuffer() envoy.BufferInstance {
	ptr := C.GoHttpFilter_DecoderCallbacks_decodingBuffer(c.filter)
	if ptr == nil {
		return nil
	}

	return bufferInstance{ptr}
}

// See //envoy/include/envoy/http/filter.h
// Continue iterating through the filter chain with buffered headers and body data. This routine
// can only be called if the filter has previously returned StopIteration from decodeHeaders()
// AND one of StopIterationAndBuffer, StopIterationAndWatermark, or StopIterationNoBuffer
// from each previous call to decodeData().
func (c decoderCallbacks) ContinueDecoding() {
	C.GoHttpFilter_DecoderCallbacks_continueDecoding(c.filter)
}

// See //envoy/include/envoy/http/filter.h
// sendLocalReply only available for StreamDecoderFilterCallbacks
// StreamEncoderFilterCallbacks doesn't have
func (c decoderCallbacks) SendLocalReply(responseCode int, body string, headers map[string]string, details string) {
	headerMap := pb.RequestHeaderMap{}
	var entries []*pb.HeaderEntry
	for k, v := range headers {
		entry := &pb.HeaderEntry{Key: k, Value: v}
		entries = append(entries, entry)
	}
	headerMap.Headers = entries
	headerBytes, err := proto.Marshal(&headerMap)
	if err != nil {
		Log(loglevel.Error, "decoderCallbacks", "can't marshall headers. "+err.Error())
		// Continue to send response with empty header
	}
	if 0 != C.GoHttpFilter_DecoderCallbacks_sendLocalReply(c.filter, C.int(responseCode), GoStr(body), GoBuf(headerBytes), GoStr(details)) {
		Log(loglevel.Error, "decoderCallbacks", "can't sendLocalReply")
	}
}

func (c decoderCallbacks) AddDecodedData(buffer envoy.BufferInstance, streamingFilter bool) {
	if buffer == nil {
		return
	}

	var streamingFilterInt int8
	if streamingFilter {
		streamingFilterInt = 1
	}

	b := buffer.(bufferInstance)
	C.GoHttpFilter_DecoderCallbacks_addDecodedData(c.filter, b.ptr, C.int(streamingFilterInt))
}

func (c decoderCallbacks) StreamInfo() envoy.StreamInfo {
	return &streamInfo{c.filter, false}
}

func (c decoderCallbacks) RouteExisting() bool {
	existing := C.GoHttpFilter_StreamFilterCallbacks_routeExisting(c.filter, GoBool(false))
	return existing != 0
}

func (c decoderCallbacks) Route() envoy.Route {
	return &route{c.filter, false}
}

func (c decoderCallbacks) EncodeHeaders(responseCode int, headers *pb.ResponseHeaderMap, endStream bool) {
	headerBytes, err := proto.Marshal(headers)
	if err != nil {
		Log(loglevel.Error, "decoderCallbacks", "can't marshall headers. "+err.Error())
		return
	}

	endStreamVal := 0
	if endStream {
		endStreamVal = 1
	}

	if 0 != C.GoHttpFilter_DecoderCallbacks_encodeHeaders(c.filter, C.int(responseCode), GoBuf(headerBytes), C.int(endStreamVal)) {
		Log(loglevel.Error, "decoderCallbacks", "can't encodeHeaders")
	}
}

func (c decoderCallbacks) ActiveSpan() envoy.Span {
	return span{filter: c.filter, spanID: -1}
}
