// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	volatile "github.com/grab/ego/ego/src/go/volatile"
	mock "github.com/stretchr/testify/mock"
)

// GenericSecretConfigProvider is an autogenerated mock type for the GenericSecretConfigProvider type
type GenericSecretConfigProvider struct {
	mock.Mock
}

// Secret provides a mock function with given fields:
func (_m *GenericSecretConfigProvider) Secret() volatile.String {
	ret := _m.Called()

	var r0 volatile.String
	if rf, ok := ret.Get(0).(func() volatile.String); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(volatile.String)
	}

	return r0
}
