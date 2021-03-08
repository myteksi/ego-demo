// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	envoy "github.com/grab/ego/ego/src/go/envoy"
	mock "github.com/stretchr/testify/mock"
)

// Span is an autogenerated mock type for the Span type
type Span struct {
	mock.Mock
}

// FinishSpan provides a mock function with given fields:
func (_m *Span) FinishSpan() {
	_m.Called()
}

// GetContext provides a mock function with given fields:
func (_m *Span) GetContext() map[string][]string {
	ret := _m.Called()

	var r0 map[string][]string
	if rf, ok := ret.Get(0).(func() map[string][]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]string)
		}
	}

	return r0
}

// SpawnChild provides a mock function with given fields: name
func (_m *Span) SpawnChild(name string) envoy.Span {
	ret := _m.Called(name)

	var r0 envoy.Span
	if rf, ok := ret.Get(0).(func(string) envoy.Span); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(envoy.Span)
		}
	}

	return r0
}