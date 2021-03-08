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

// requestheadermap implements envoy.RequestHeaderMap
//
type requestTrailerMap struct{ ptr unsafe.Pointer }

// AddCopy translates to
// Envoy::Http::RequestHeaderMap::addCopy(LowerCaseString, absl::string_view).
//
// See //envoy/include/envoy/http/header_map.h
func (h requestTrailerMap) AddCopy(name, value string) {
	C.RequestTrailerMap_add(h.ptr, GoStr(name), GoStr(value))
}

func (h requestTrailerMap) SetCopy(name, value string) {
	panic("Not implemented yet")
}

func (h requestTrailerMap) AppendCopy(name, value string) {
	panic("Not implemented yet")
}

func (h requestTrailerMap) Remove(name string) {
	panic("Not implemented yet")
}

func (h requestTrailerMap) Get(name string) volatile.String {
	panic("Not implemented yet")
	return volatile.String("")
}

func (h requestTrailerMap) GetByPrefix(prefix string) map[string][]string {
	panic("Not implemented yet")
	return map[string][]string{}
}
