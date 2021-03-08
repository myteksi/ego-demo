// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package stats

// see envoy/include/envoy/stats/stats.h

// ImportMode for Gauge Metric
type ImportMode int

const (
	// Uninitialized means Gauge was discovered during hot-restart transfer.
	Uninitialized ImportMode = 1
	// NeverImport means On hot-restart, each process starts with gauge at 0.
	NeverImport ImportMode = 2
	// Accumulate means Transfers gauge state on hot-restart.
	Accumulate ImportMode = 3
)

// see envoy/include/envoy/stats/histogram.h

// Unit for Histogram Metric
type Unit int

const (
	// Null means The histogram has been rejected, i.e. it's a null histogram
	// and is not recording anything.
	Null Unit = 1
	// Unspecified means Measured quantity does not require a unit, e.g. "items".
	Unspecified Unit = 2
	// Bytes ...
	Bytes Unit = 3
	// Microseconds ...
	Microseconds Unit = 4
	// Milliseconds ...
	Milliseconds Unit = 5
)
