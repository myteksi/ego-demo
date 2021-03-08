// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	envoy "github.com/grab/ego/ego/src/go/envoy"
	mock "github.com/stretchr/testify/mock"

	volatile "github.com/grab/ego/ego/src/go/volatile"
)

// GoHttpFilterConfig is an autogenerated mock type for the GoHttpFilterConfig type
type GoHttpFilterConfig struct {
	mock.Mock
}

// Scope provides a mock function with given fields:
func (_m *GoHttpFilterConfig) Scope() envoy.Scope {
	ret := _m.Called()

	var r0 envoy.Scope
	if rf, ok := ret.Get(0).(func() envoy.Scope); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(envoy.Scope)
		}
	}

	return r0
}

// Settings provides a mock function with given fields:
func (_m *GoHttpFilterConfig) Settings() volatile.Bytes {
	ret := _m.Called()

	var r0 volatile.Bytes
	if rf, ok := ret.Get(0).(func() volatile.Bytes); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(volatile.Bytes)
		}
	}

	return r0
}