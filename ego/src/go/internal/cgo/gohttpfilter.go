// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"fmt"
	"unsafe"

	ego "github.com/grab/ego/ego/src/go"
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/envoy/datastatus"
	"github.com/grab/ego/ego/src/go/envoy/headersstatus"
	"github.com/grab/ego/ego/src/go/envoy/loglevel"
	"github.com/grab/ego/ego/src/go/envoy/trailersstatus"
	"github.com/grab/ego/ego/src/go/volatile"
)

type goHttpFilter struct {
	filter unsafe.Pointer
}

func newGoHttpFilter(ptr unsafe.Pointer) envoy.GoHttpFilter {
	return goHttpFilter{ptr}
}

func (f goHttpFilter) DecoderCallbacks() envoy.DecoderFilterCallbacks {
	return decoderCallbacks{f.filter}
}

func (f goHttpFilter) EncoderCallbacks() envoy.EncoderFilterCallbacks {
	return encoderCallbacks{f.filter}
}

func (f goHttpFilter) ResolveMostSpecificPerGoFilterConfig(name string, route envoy.Route) interface{} {
	cgoTag := C.GoHttpFilter_ResolveMostSpecificPerGoFilterConfig(f.filter, GoStr(name))
	return GetRouteSpecificFilterConfig(uint64(cgoTag))
}

func (f goHttpFilter) GenericSecretProvider() envoy.GenericSecretConfigProvider {
	return genericSecretConfigProvider{f.filter}
}

func (f goHttpFilter) Post(tag uint64) {
	C.GoHttpFilter_post(f.filter, C.uint64_t(tag))
}

func (f goHttpFilter) Pin() {
	C.GoHttpFilter_pin(f.filter)
}

func (f goHttpFilter) Unpin() {
	C.GoHttpFilter_unpin(f.filter)
}

// Log with two simple paramters level & message, we can extend it with
// keyvals ...interface{} & logstring := l.getLogstring(keyvals) from structured log wrapper it
// It will not optimize for performance such as don't build the message if loglevel isn't match
func (f goHttpFilter) Log(logLevel loglevel.Type, message string) {
	C.GoHttpFilter_log(f.filter, C.uint32_t(logLevel), GoStr(message))
}

type genericSecretConfigProvider struct {
	filter unsafe.Pointer
}

func (p genericSecretConfigProvider) Secret() volatile.String {
	var value C.GoStr
	C.GoHttpFilter_GenericSecretConfigProvider_secret(p.filter, &value)
	return CStrN(value.data, value.len)
}

// Cgo_GoHttpFilter_DecodeHeaders is the entry point for
// Envoy::Http::GoHttpFilter::decodeHeaders().
// See //src/cc/filters/http/go/filter-cgo.cc
//
//export Cgo_GoHttpFilter_DecodeHeaders
func Cgo_GoHttpFilter_DecodeHeaders(filterTag uint64, headers unsafe.Pointer, end_stream C.int) int {
	return int(cgo_GoHttpFilter_DecodeHeaders(filterTag, headers, end_stream))
}

func cgo_GoHttpFilter_DecodeHeaders(filterTag uint64, headers unsafe.Pointer, end_stream C.int) (result headersstatus.Type) {
	const tag = "cgo_GoHttpFilter_DecodeHeaders"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
			// FIXME: emit 500
			// TODO(tien.nguyen): think about how can call to C side without filter point to call sendLocalReply
			result = headersstatus.StopIteration
		}
	}()
	filter := GetHttpFilter(filterTag)
	if nil == filter {
		Log(loglevel.Error, tag, "nil filter")
		// FIXME: emit 500
		return headersstatus.StopIteration
	}
	return filter.DecodeHeaders(requestHeaderMap{headers}, end_stream != 0)
}

// Cgo_GoHttpFilter_DecodeData is the entry point for
// Envoy::Http::GoHttpFilter::decodeData().
// See //src/cc/filters/http/go/filter-cgo.cc
//
//export Cgo_GoHttpFilter_DecodeData
func Cgo_GoHttpFilter_DecodeData(filterTag uint64, buffer unsafe.Pointer, end_stream C.int) int {
	return int(cgo_GoHttpFilter_DecodeData(filterTag, buffer, end_stream))
}

func cgo_GoHttpFilter_DecodeData(filterTag uint64, buffer unsafe.Pointer, end_stream C.int) (result datastatus.Type) {
	const tag = "cgo_GoHttpFilter_DecodeData"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
			result = datastatus.StopIterationNoBuffer
		}
	}()
	filter := GetHttpFilter(filterTag)
	if nil == filter {
		Log(loglevel.Error, tag, "nil filter")
		// FIXME: emit 500
		return datastatus.StopIterationNoBuffer
	}
	return filter.DecodeData(bufferInstance{buffer}, end_stream != 0)
}

// Cgo_GoHttpFilter_DecodeTrailers is the entry point for
// Envoy::Http::GoHttpFilter::decodeTrailers().
// See //src/cc/filters/http/go/filter-cgo.cc
//
//export Cgo_GoHttpFilter_DecodeTrailers
func Cgo_GoHttpFilter_DecodeTrailers(filterTag uint64, trailers unsafe.Pointer) int {
	return int(cgo_GoHttpFilter_DecodeTrailers(filterTag, trailers))
}

func cgo_GoHttpFilter_DecodeTrailers(filterTag uint64, trailers unsafe.Pointer) (result trailersstatus.Type) {
	const tag = "cgo_GoHttpFilter_DecodeTrailers"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
			// FIXME: emit 500
			result = trailersstatus.StopIteration
		}
	}()
	filter := GetHttpFilter(filterTag)
	if nil == filter {
		Log(loglevel.Error, tag, "nil filter")
		// FIXME: emit 500
		return trailersstatus.StopIteration
	}
	return filter.DecodeTrailers(requestTrailerMap{trailers})
}

// Cgo_GoHttpFilter_OnPost is the entry point for
// Envoy::Http::GoHttpFilter::onPost().
// See //src/cc/filters/http/go/filter-cgo.cc
//
//export Cgo_GoHttpFilter_OnPost
func Cgo_GoHttpFilter_OnPost(filterTag, postTag uint64) {
	const tag = "Cgo_GoHttpFilter_OnPost"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
		}
	}()
	f := GetHttpFilter(filterTag)
	if nil == f {
		Log(loglevel.Error, tag, "nil filter")
		return
	}
	f.OnPost(postTag)
}

// Cgo_GoHttpFilter_OnDestroy is the entry point for
// Envoy::Http::GoHttpFilter::onDestroy().
// See //src/cc/filters/http/go/filter-cgo.cc
//
//export Cgo_GoHttpFilter_OnDestroy
func Cgo_GoHttpFilter_OnDestroy(filterTag uint64) {
	const tag = "Cgo_GoHttpFilter_OnDestroy"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
		}
	}()
	f := RemoveHttpFilter(filterTag)
	if nil == f {
		Log(loglevel.Error, tag, "nil filter")
		return
	}
	f.OnDestroy()
}

//export Cgo_GoHttpFilter_Create
func Cgo_GoHttpFilter_Create(native unsafe.Pointer, factoryTag uint64, filterSlot uint64) (result uint64) {
	const tag = "Cgo_GoHttpFilter_Create"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
			result = 0
		}
	}()

	// NOTE: we are not sure if we are running on the same thread as the
	// filter factory creation. But we know the factory _is_ alive right
	// now, so this is safe.
	filterFactory := GetHttpFilterFactory(factoryTag)
	if nil == filterFactory {
		Log(loglevel.Error, tag, "nil filterFactory")
		return 0
	}

	filter := filterFactory(newGoHttpFilter(native))
	if nil == filter {
		Log(loglevel.Error, tag, "nil filter")
		return 0
	}

	return TagHttpFilter(filterSlot, filter)
}

//export Cgo_GoHttpFilter_EncodeHeaders
func Cgo_GoHttpFilter_EncodeHeaders(filterTag uint64, headers unsafe.Pointer, end_stream C.int) int {
	return int(cgo_GoHttpFilter_EncodeHeaders(filterTag, headers, end_stream))
}

func cgo_GoHttpFilter_EncodeHeaders(filterTag uint64, headers unsafe.Pointer, end_stream C.int) (result headersstatus.Type) {
	const tag = "cgo_GoHttpFilter_EncodeHeaders"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
			// FIXME: emit 500
			result = headersstatus.StopIteration
		}
	}()
	filter := GetHttpFilter(filterTag)
	if nil == filter {
		Log(loglevel.Error, tag, "nil filter")
		// FIXME: emit 500
		return headersstatus.StopIteration
	}

	return filter.EncodeHeaders(responseHeaderMap{headers}, end_stream != 0)
}

//export Cgo_GoHttpFilter_EncodeData
func Cgo_GoHttpFilter_EncodeData(filterTag uint64, buffer unsafe.Pointer, end_stream C.int) int {
	return int(cgo_GoHttpFilter_EncodeData(filterTag, buffer, end_stream))
}

func cgo_GoHttpFilter_EncodeData(filterTag uint64, buffer unsafe.Pointer, end_stream C.int) (result datastatus.Type) {
	const tag = "cgo_GoHttpFilter_EncodeData"
	defer func() {
		if err := recover(); err != nil {
			Log(loglevel.Error, tag, fmt.Sprintf("%v", err))
			result = datastatus.StopIterationNoBuffer
		}
	}()
	filter := GetHttpFilter(filterTag)
	if nil == filter {
		Log(loglevel.Error, tag, "nil filter")
		// FIXME: emit 500
		return datastatus.StopIterationNoBuffer
	}

	return filter.EncodeData(bufferInstance{buffer}, end_stream != 0)
}

// httpFilters is a clutch to bridge the "air gap" between the C++ filter object
// and the go filter state. We share this among all filters, but in case the 16M
// clutch entries turn out to be insufficient, we can create one clutch per
// filter name (and stealing a few bits from the tag mark for the clutch ID).
//
var httpFilters = &clutch{}

// Cgo_AcquireHttpFilterSlot is the public proxy for httpFilters.AcquireSlot
//
//export Cgo_AcquireHttpFilterSlot
func Cgo_AcquireHttpFilterSlot() uint64 {
	return httpFilters.AcquireSlot()
}

//Cgo_ReleaseHttpFilterSlot is the public proxy for httpFilters.ReleaseSlot
//
//export Cgo_ReleaseHttpFilterSlot
func Cgo_ReleaseHttpFilterSlot(id uint64) {
	httpFilters.ReleaseSlot(id)
}

// TagHttpFilter is the public proxy for httpFilters.TagItem
//
func TagHttpFilter(slot uint64, filter ego.HttpFilter) uint64 {
	return httpFilters.TagItem(slot, filter)
}

// GetHttpFilter is the public proxy for httpFilters.GetItem
//
func GetHttpFilter(tag uint64) ego.HttpFilter {
	filter, _ := httpFilters.GetItem(tag).(ego.HttpFilter)
	return filter
}

// RemoveHttpFilter is the public proxy for httpFilters.RemoveItem
//
func RemoveHttpFilter(tag uint64) ego.HttpFilter {
	filter, _ := httpFilters.RemoveItem(tag).(ego.HttpFilter)
	return filter
}
