// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	volatile "github.com/grab/ego/ego/src/go/volatile"
	mock "github.com/stretchr/testify/mock"
)

// RequestTrailerMapReadOnly is an autogenerated mock type for the RequestTrailerMapReadOnly type
type RequestTrailerMapReadOnly struct {
	mock.Mock
}

// Get provides a mock function with given fields: name
func (_m *RequestTrailerMapReadOnly) Get(name string) volatile.String {
	ret := _m.Called(name)

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func(string) volatile.String); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}
