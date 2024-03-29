// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	pty "github.com/thecoderstudio/apollo-agent/pty"

	websocket "github.com/thecoderstudio/apollo-agent/websocket"
)

// ManagerInterface is an autogenerated mock type for the ManagerInterface type
type ManagerInterface struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *ManagerInterface) Close() {
	_m.Called()
}

// CreateNewSession provides a mock function with given fields: _a0
func (_m *ManagerInterface) CreateNewSession(_a0 string) (*pty.Session, error) {
	ret := _m.Called(_a0)

	var r0 *pty.Session
	if rf, ok := ret.Get(0).(func(string) *pty.Session); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pty.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Execute provides a mock function with given fields: _a0
func (_m *ManagerInterface) Execute(_a0 websocket.ShellIO) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(websocket.ShellIO) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecutePredefinedCommand provides a mock function with given fields: _a0
func (_m *ManagerInterface) ExecutePredefinedCommand(_a0 websocket.Command) {
	_m.Called(_a0)
}

// GetSession provides a mock function with given fields: _a0
func (_m *ManagerInterface) GetSession(_a0 string) *pty.Session {
	ret := _m.Called(_a0)

	var r0 *pty.Session
	if rf, ok := ret.Get(0).(func(string) *pty.Session); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pty.Session)
		}
	}

	return r0
}

// Out provides a mock function with given fields:
func (_m *ManagerInterface) Out() <-chan websocket.Message {
	ret := _m.Called()

	var r0 <-chan websocket.Message
	if rf, ok := ret.Get(0).(func() <-chan websocket.Message); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan websocket.Message)
		}
	}

	return r0
}
