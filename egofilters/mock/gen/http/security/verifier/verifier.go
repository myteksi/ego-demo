// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	context "github.com/grab/ego/egofilters/http/security/context"
	mock "github.com/stretchr/testify/mock"
)

// Verifier is an autogenerated mock type for the Verifier type
type Verifier struct {
	mock.Mock
}

// Verify provides a mock function with given fields: _a0
func (_m *Verifier) Verify(_a0 context.RequestContext) {
	_m.Called(_a0)
}

// WithBody provides a mock function with given fields:
func (_m *Verifier) WithBody() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}