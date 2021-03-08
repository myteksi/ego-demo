// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package statetype

// Envoy::StreamInfo::FilterState::StateType
// The constants have been chosen with arbtirary offsets to easier detect
// return value type mismatches.
//
// see //envoy/include/envoy/stream_info/filter_state.h
type Type int

const (
	ReadOnly Type = 1
	Mutable  Type = 2
)
