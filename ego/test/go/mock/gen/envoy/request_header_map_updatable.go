// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// requestHeaderMapUpdatable is an autogenerated mock type for the requestHeaderMapUpdatable type
type requestHeaderMapUpdatable struct {
	mock.Mock
}

// AddCopy provides a mock function with given fields: name, value
func (_m *requestHeaderMapUpdatable) AddCopy(name string, value string) {
	_m.Called(name, value)
}

// AppendCopy provides a mock function with given fields: name, value
func (_m *requestHeaderMapUpdatable) AppendCopy(name string, value string) {
	_m.Called(name, value)
}

// Remove provides a mock function with given fields: name
func (_m *requestHeaderMapUpdatable) Remove(name string) {
	_m.Called(name)
}

// SetCopy provides a mock function with given fields: name, value
func (_m *requestHeaderMapUpdatable) SetCopy(name string, value string) {
	_m.Called(name, value)
}

// SetPath provides a mock function with given fields: path
func (_m *requestHeaderMapUpdatable) SetPath(path string) {
	_m.Called(path)
}
