// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"unsafe"

	"github.com/grab/ego/ego/src/go/volatile"
)

// responseHeaderMap implements envoy.ResponseHeaderMap
//
type responseHeaderMap struct{ ptr unsafe.Pointer }

// AddCopy translates to
// Envoy::Http::ResponseHeaderMap::addCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) AddCopy(name, value string) {
	C.ResponseHeaderMap_add(h.ptr, GoStr(name), GoStr(value))
}

// SetCopy translates to
// Envoy::Http::ResponseHeaderMap::setCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) SetCopy(name, value string) {
	C.ResponseHeaderMap_set(h.ptr, GoStr(name), GoStr(value))
}

// AppendCopy translates to
// Envoy::Http::ResponseHeaderMap::appendCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) AppendCopy(name, value string) {
	C.ResponseHeaderMap_append(h.ptr, GoStr(name), GoStr(value))
}

// Remove translates to
// Envoy::Http::ResponseHeaderMap::remove(LowerCaseString).
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) Remove(name string) {
	C.ResponseHeaderMap_remove(h.ptr, GoStr(name))
}

// ContentType translates to
// Envoy::Http::ResponseHeaderMap::ContentType().
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) ContentType() volatile.String {
	var value C.GoStr
	C.ResponseHeaderMap_ContentType(h.ptr, &value)
	return CStrN(value.data, value.len)
}

// Status translates to
// Envoy::Http::ResponseHeaderMap::Status().
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) Status() volatile.String {
	var value C.GoStr
	C.ResponseHeaderMap_Status(h.ptr, &value)
	return CStrN(value.data, value.len)
}

// SetStatus translates to
// Envoy::Http::ResponseHeaderMap::setStatus(uint64_t).
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) SetStatus(status int) {
	C.ResponseHeaderMap_setStatus(h.ptr, C.int(status))

}

// Get translates to
// Envoy::Http::ResponseHeaderMap::Get(LowerCaseString).
//
// See //envoy/include/envoy/http/header_map.h
func (h responseHeaderMap) Get(name string) volatile.String {
	var value C.GoStr
	C.ResponseHeaderMap_get(h.ptr, GoStr(name), &value)
	return CStrN(value.data, value.len)
}
