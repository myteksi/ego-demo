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
	"github.com/grab/ego/ego/src/go/volatile"
)

// requestheadermap implements envoy.RequestHeaderMap
//
type requestHeaderMap struct{ ptr unsafe.Pointer }

// AddCopy translates to
// Envoy::Http::RequestHeaderMap::addCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h requestHeaderMap) AddCopy(name, value string) {
	C.RequestHeaderMap_add(h.ptr, GoStr(name), GoStr(value))
}

// SetCopy translates to
// Envoy::Http::RequestHeaderMap::setCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h requestHeaderMap) SetCopy(name, value string) {
	C.RequestHeaderMap_set(h.ptr, GoStr(name), GoStr(value))
}

// AppendCopy translates to
// Envoy::Http::RequestHeaderMap::appendCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h requestHeaderMap) AppendCopy(name, value string) {
	C.RequestHeaderMap_append(h.ptr, GoStr(name), GoStr(value))
}

func (h requestHeaderMap) GetByPrefix(prefix string) map[string][]string {
	defaultBufferSize := uint64(100)
	data := make([]byte, defaultBufferSize)
	size := uint64(C.RequestHeaderMap_getByPrefix(h.ptr, GoStr(prefix), GoBuf(data)))
	if size > defaultBufferSize {
		newBufferSize := size
		data = make([]byte, newBufferSize)
		size = uint64(C.RequestHeaderMap_getByPrefix(h.ptr, GoStr(prefix), GoBuf(data)))
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

// Remove translates to
// Envoy::Http::RequestHeaderMap::remove(LowerCaseString).
//
// See //envoy/include/envoy/http/header_map.h
func (h requestHeaderMap) Remove(name string) {
	C.RequestHeaderMap_remove(h.ptr, GoStr(name))
}

func (h requestHeaderMap) Path() volatile.String {
	var value C.GoStr
	C.RequestHeaderMap_Path(h.ptr, &value)
	return CStrN(value.data, value.len)
}

func (h requestHeaderMap) SetPath(path string) {
	C.RequestHeaderMap_setPath(h.ptr, GoStr(path))
}

func (h requestHeaderMap) Method() volatile.String {
	var value C.GoStr
	C.RequestHeaderMap_Method(h.ptr, &value)
	return CStrN(value.data, value.len)
}

func (h requestHeaderMap) Authorization() volatile.String {
	var value C.GoStr
	C.RequestHeaderMap_Authorization(h.ptr, &value)
	return CStrN(value.data, value.len)
}

func (h requestHeaderMap) ContentType() volatile.String {
	var value C.GoStr
	C.RequestHeaderMap_ContentType(h.ptr, &value)
	return CStrN(value.data, value.len)
}

func (h requestHeaderMap) Get(name string) volatile.String {
	var value C.GoStr
	C.RequestHeaderMap_get(h.ptr, GoStr(name), &value)
	return CStrN(value.data, value.len)
}
