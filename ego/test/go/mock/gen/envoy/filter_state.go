// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	lifespan "github.com/grab/ego/ego/src/go/envoy/lifespan"
	mock "github.com/stretchr/testify/mock"

	statetype "github.com/grab/ego/ego/src/go/envoy/statetype"

	volatile "github.com/grab/ego/ego/src/go/volatile"
)

// FilterState is an autogenerated mock type for the FilterState type
type FilterState struct {
	mock.Mock
}

// GetDataReadOnly provides a mock function with given fields: name
func (_m *FilterState) GetDataReadOnly(name string) (volatile.String, bool) {
	ret := _m.Called(name)

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func(string) volatile.String); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// SetData provides a mock function with given fields: name, value, stateType, lifeSpan
func (_m *FilterState) SetData(name string, value string, stateType statetype.Type, lifeSpan lifespan.Type) {
	_m.Called(name, value, stateType, lifeSpan)
}