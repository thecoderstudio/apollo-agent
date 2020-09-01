// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	oauth "github.com/thecoderstudio/apollo-agent/oauth"

	url "net/url"

	websocket "github.com/thecoderstudio/apollo-agent/websocket"
)

// ShellInterface is an autogenerated mock type for the ShellInterface type
type ShellInterface struct {
	mock.Mock
}

// Commands provides a mock function with given fields:
func (_m *ShellInterface) Commands() <-chan websocket.Command {
	ret := _m.Called()

	var r0 <-chan websocket.Command
	if rf, ok := ret.Get(0).(func() <-chan websocket.Command); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan websocket.Command)
		}
	}

	return r0
}

// Errs provides a mock function with given fields:
func (_m *ShellInterface) Errs() <-chan error {
	ret := _m.Called()

	var r0 <-chan error
	if rf, ok := ret.Get(0).(func() <-chan error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan error)
		}
	}

	return r0
}

// Listen provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *ShellInterface) Listen(_a0 url.URL, _a1 oauth.AccessToken, _a2 *chan websocket.ShellIO, _a3 *chan struct{}) <-chan struct{} {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func(url.URL, oauth.AccessToken, *chan websocket.ShellIO, *chan struct{}) <-chan struct{}); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// Out provides a mock function with given fields:
func (_m *ShellInterface) Out() <-chan websocket.ShellIO {
	ret := _m.Called()

	var r0 <-chan websocket.ShellIO
	if rf, ok := ret.Get(0).(func() <-chan websocket.ShellIO); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan websocket.ShellIO)
		}
	}

	return r0
}
