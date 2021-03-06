// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	volatile "github.com/grab/ego/ego/src/go/volatile"
	mock "github.com/stretchr/testify/mock"
)

// RequestHeaderMapReadOnly is an autogenerated mock type for the RequestHeaderMapReadOnly type
type RequestHeaderMapReadOnly struct {
	mock.Mock
}

// Authorization provides a mock function with given fields:
func (_m *RequestHeaderMapReadOnly) Authorization() volatile.String {
	ret := _m.Called()

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func() volatile.String); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}

// ContentType provides a mock function with given fields:
func (_m *RequestHeaderMapReadOnly) ContentType() volatile.String {
	ret := _m.Called()

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func() volatile.String); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}

// Get provides a mock function with given fields: name
func (_m *RequestHeaderMapReadOnly) Get(name string) volatile.String {
	ret := _m.Called(name)

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func(string) volatile.String); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}

// GetByPrefix provides a mock function with given fields: prefix
func (_m *RequestHeaderMapReadOnly) GetByPrefix(prefix string) map[string][]string {
	ret := _m.Called(prefix)

	var r0 map[string][]string
	if rf, ok := ret.Get(0).(func(string) map[string][]string); ok {
		r0 = rf(prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]string)
		}
	}

	return r0
}

// Method provides a mock function with given fields:
func (_m *RequestHeaderMapReadOnly) Method() volatile.String {
	ret := _m.Called()

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func() volatile.String); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}

// Path provides a mock function with given fields:
func (_m *RequestHeaderMapReadOnly) Path() volatile.String {
	ret := _m.Called()

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func() volatile.String); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}
