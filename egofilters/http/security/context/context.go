// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package context

import "github.com/grab/ego/ego/src/go/envoy"

type Context interface {
	ActiveSpan() envoy.Span
}
