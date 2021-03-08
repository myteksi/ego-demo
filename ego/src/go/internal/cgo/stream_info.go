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
	"github.com/grab/ego/ego/src/go/volatile"
)

type streamInfo struct {
	filter  unsafe.Pointer
	encoder bool
}

func (i streamInfo) FilterState() envoy.FilterState {
	return filterState{i.filter, i.encoder}
}

func (i streamInfo) LastDownstreamTxByteSent() int64 {
	return CLong(C.GoHttpFilter_StreamFilterCallbacks_StreamInfo_lastDownstreamTxByteSent(i.filter, GoBool(i.encoder)))
}

func (i streamInfo) GetRequestHeaders() envoy.RequestHeaderMapReadOnly {
	ptr := C.GoHttpFilter_StreamFilterCallbacks_StreamInfo_getRequestHeaders(i.filter, GoBool(i.encoder))
	if ptr == nil {
		return nil
	}
	return &requestHeaderMap{ptr}
}

func (i streamInfo) ResponseCode() int {
	return int(C.GoHttpFilter_StreamFilterCallbacks_StreamInfo_responseCode(i.filter, GoBool(i.encoder)))
}

func (i streamInfo) ResponseCodeDetails() volatile.String {
	var value C.GoStr
	C.GoHttpFilter_StreamFilterCallbacks_StreamInfo_responseCodeDetails(i.filter, GoBool(i.encoder), &value)
	return CStrN(value.data, value.len)
}
