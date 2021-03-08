// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"unsafe"

	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/envoy/stats"
)

type scope struct {
	ptr unsafe.Pointer
}

func (s scope) CounterFromStatName(name string) envoy.Counter {
	ptr := C.Stats_Scope_counterFromStatName(s.ptr, GoStr(name))
	if ptr == nil {
		return nil
	}
	return counter{ptr}
}

func (s scope) GaugeFromStatName(name string, importMode stats.ImportMode) envoy.Gauge {
	ptr := C.Stats_Scope_gaugeFromStatName(s.ptr, GoStr(name), C.int(importMode))
	if ptr == nil {
		return nil
	}
	return gauge{ptr}
}

func (s scope) HistogramFromStatName(name string, unit stats.Unit) envoy.Histogram {
	ptr := C.Stats_Scope_histogramFromStatName(s.ptr, GoStr(name), C.int(unit))
	if ptr == nil {
		return nil
	}
	return histogram{ptr}
}

type counter struct {
	ptr unsafe.Pointer
}

func (c counter) Add(amount uint64) {
	C.Stats_Counter_add(c.ptr, C.uint64_t(amount))
}

func (c counter) Inc() {
	C.Stats_Counter_inc(c.ptr)
}

func (c counter) Latch() uint64 {
	return uint64(C.Stats_Counter_latch(c.ptr))
}

func (c counter) Reset() {
	C.Stats_Counter_reset(c.ptr)
}

func (c counter) Value() uint64 {
	return uint64(C.Stats_Counter_value(c.ptr))

}

type gauge struct {
	ptr unsafe.Pointer
}

func (g gauge) Add(amount uint64) {
	C.Stats_Gauge_add(g.ptr, C.uint64_t(amount))
}

func (g gauge) Dec() {
	C.Stats_Gauge_dec(g.ptr)
}

func (g gauge) Inc() {
	C.Stats_Gauge_inc(g.ptr)
}

func (g gauge) Set(value uint64) {
	C.Stats_Gauge_set(g.ptr, C.uint64_t(value))
}

func (g gauge) Sub(amount uint64) {
	C.Stats_Gauge_sub(g.ptr, C.uint64_t(amount))
}

func (g gauge) Value() uint64 {
	return uint64(C.Stats_Gauge_value(g.ptr))
}

type histogram struct {
	ptr unsafe.Pointer
}

func (h histogram) Unit() stats.Unit {
	return stats.Unit(C.Stats_Histogram_unit(h.ptr))
}

func (h histogram) RecordValue(value uint64) {
	C.Stats_Histogram_recordValue(h.ptr, C.uint64_t(value))
}
