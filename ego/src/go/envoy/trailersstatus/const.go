// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package trailersstatus

// Envoy::Http::FilterTrailerStatus
// The constants have been chosen with arbtirary offsets to easier detect
// return value type mismatches.
//
// see //envoy/include/envoy/http/filter.h
type Type int
const (
	Continue      Type = 201
	StopIteration Type = 202
)
