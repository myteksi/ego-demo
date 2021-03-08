// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

// #include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/volatile"
)

type route struct {
	filter  unsafe.Pointer
	encoder bool
}

func (r *route) RouteEntry() envoy.RouteEntry {
	return &routeEntry{r.filter, r.encoder}
}

type routeEntry struct {
	filter  unsafe.Pointer
	encoder bool
}

func (e *routeEntry) PathMatchCriterion() envoy.PathMatchCriterion {
	return &pathMatchCriterion{e.filter, e.encoder}
}

type pathMatchCriterion struct {
	filter  unsafe.Pointer
	encoder bool
}

func (c *pathMatchCriterion) MatchType() (envoy.PathMatchType, error) {
	matchType := C.GoHttpFilter_StreamFilterCallbacks_route_routeEntry_pathMatchCriterion_matchType(c.filter, GoBool(c.encoder))
	switch matchType {
	case 0:
		return envoy.PathMatchNone, nil
	case 1:
		return envoy.PathMatchPrefix, nil
	case 2:
		return envoy.PathMatchExact, nil
	case 3:
		return envoy.PathMatchRegex, nil
	}
	return envoy.PathMatchNone, fmt.Errorf("invalid match type of %v", matchType)
}

func (c *pathMatchCriterion) Matcher() (volatile.String, error) {
	var value C.GoStr
	errCode := C.GoHttpFilter_StreamFilterCallbacks_route_routeEntry_pathMatchCriterion_matcher(c.filter, GoBool(c.encoder), &value)
	if errCode != 0 {
		return volatile.String(""), fmt.Errorf("can't get matcher. error code %d", errCode)
	}

	return CStrN(value.data, value.len), nil
}
