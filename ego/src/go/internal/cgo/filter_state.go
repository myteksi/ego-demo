// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"unsafe"

	"github.com/grab/ego/ego/src/go/envoy/lifespan"
	"github.com/grab/ego/ego/src/go/envoy/statetype"
	"github.com/grab/ego/ego/src/go/volatile"
)

type filterState struct {
	filter  unsafe.Pointer
	encoder bool
}

func (s filterState) SetData(name, value string, stateType statetype.Type, lifeSpan lifespan.Type) {
	C.GoHttpFilter_DecoderCallbacks_StreamInfo_FilterState_setData(s.filter, GoStr(name), GoStr(value), C.int(stateType), C.int(lifeSpan))
}

func (s filterState) GetDataReadOnly(name string) (volatile.String, bool) {
	var value C.GoStr
	ok := C.GoHttpFilter_StreamFilterCallbacks_StreamInfo_FilterState_getDataReadOnly(s.filter, GoBool(s.encoder), GoStr(name), &value)
	return CStrN(value.data, value.len), ok != 0
}
