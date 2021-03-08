// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package datastatus

// Envoy::Http::FilterDataStatus
// The constants have been chosen with arbtirary offsets to easier detect
// return value type mismatches.
//
// see //envoy/include/envoy/http/filter.h

type Type int
const (
	Continue                  Type = 300
	StopIterationAndBuffer    Type = 301
	StopIterationAndWatermark Type = 302
	StopIterationNoBuffer     Type = 303
)
