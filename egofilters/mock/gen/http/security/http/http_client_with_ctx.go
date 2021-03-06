// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	http "net/http"

	context "github.com/grab/ego/egofilters/http/security/context"

	mock "github.com/stretchr/testify/mock"
)

// HttpClientWithCtx is an autogenerated mock type for the HttpClientWithCtx type
type HttpClientWithCtx struct {
	mock.Mock
}

// Do provides a mock function with given fields: req
func (_m *HttpClientWithCtx) Do(req *http.Request) (*http.Response, error) {
	ret := _m.Called(req)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Request) *http.Response); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DoWithTracing provides a mock function with given fields: ctx, req, spanName
func (_m *HttpClientWithCtx) DoWithTracing(ctx context.Context, req *http.Request, spanName string) (*http.Response, error) {
	ret := _m.Called(ctx, req, spanName)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, string) *http.Response); ok {
		r0 = rf(ctx, req, spanName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *http.Request, string) error); ok {
		r1 = rf(ctx, req, spanName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
