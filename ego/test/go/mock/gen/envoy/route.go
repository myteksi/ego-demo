// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	envoy "github.com/grab/ego/ego/src/go/envoy"
	mock "github.com/stretchr/testify/mock"
)

// Route is an autogenerated mock type for the Route type
type Route struct {
	mock.Mock
}

// RouteEntry provides a mock function with given fields:
func (_m *Route) RouteEntry() envoy.RouteEntry {
	ret := _m.Called()

	var r0 envoy.RouteEntry
	if rf, ok := ret.Get(0).(func() envoy.RouteEntry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(envoy.RouteEntry)
		}
	}

	return r0
}
