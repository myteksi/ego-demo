// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package lifespan

// Envoy::StreamInfo::FilterState::LifeSpan
// The constants have been chosen with arbtirary offsets to easier detect
// return value type mismatches.
//
// see //envoy/include/envoy/stream_info/filter_state.h
type Type int

const (
	FilterChain          Type = 1
	DownstreamRequest    Type = 2
	DownstreamConnection Type = 3
	TopSpan              Type = 4
)
