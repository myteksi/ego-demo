// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// headerMapUpdatable is an autogenerated mock type for the headerMapUpdatable type
type headerMapUpdatable struct {
	mock.Mock
}

// AddCopy provides a mock function with given fields: name, value
func (_m *headerMapUpdatable) AddCopy(name string, value string) {
	_m.Called(name, value)
}

// AppendCopy provides a mock function with given fields: name, value
func (_m *headerMapUpdatable) AppendCopy(name string, value string) {
	_m.Called(name, value)
}

// Remove provides a mock function with given fields: name
func (_m *headerMapUpdatable) Remove(name string) {
	_m.Called(name)
}

// SetCopy provides a mock function with given fields: name, value
func (_m *headerMapUpdatable) SetCopy(name string, value string) {
	_m.Called(name, value)
}
