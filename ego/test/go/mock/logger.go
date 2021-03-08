// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package mock

import (
	"fmt"

	"github.com/grab/ego/ego/src/go/envoy/loglevel"
)

type NativeLogger struct{}

// Log ...
func (l NativeLogger) Log(level loglevel.Type, tag, message string) {
	fmt.Printf("[%v]: %v\n", tag, message)
}
