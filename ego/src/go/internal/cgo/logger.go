// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"

import (
	"github.com/grab/ego/ego/src/go/envoy/loglevel"
)

// Log is a basic log function and base on that developers can extend or write their logger
func Log(level loglevel.Type, tag, message string) {
	C.Envoy_log_misc(C.uint32_t(level), GoStr(tag), GoStr(message))
}

type nativeLogger struct {
}

func (l nativeLogger) Log(level loglevel.Type, tag, message string) {
	Log(level, tag, message)
}
