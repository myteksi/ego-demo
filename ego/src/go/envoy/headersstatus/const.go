// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package headersstatus

// Envoy::Http::FilterHeadersStatus
// The constants have been chosen with arbtirary offsets to easier detect
// return value type mismatches.
//
// see //envoy/include/envoy/http/filter.h
type Type int
const (
	Continue                     Type = 100
	StopIteration                Type = 101
	ContinueAndEndStream         Type = 102
	StopAllIterationAndBuffer    Type = 103
	StopAllIterationAndWatermark Type = 104
)
