// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
	websocket "github.com/thecoderstudio/apollo-agent/websocket"
)

// Dialer is an autogenerated mock type for the Dialer type
type Dialer struct {
	mock.Mock
}

// Dial provides a mock function with given fields: _a0, _a1
func (_m *Dialer) Dial(_a0 string, _a1 http.Header) (websocket.Connection, *http.Response, error) {
	ret := _m.Called(_a0, _a1)

	var r0 websocket.Connection
	if rf, ok := ret.Get(0).(func(string, http.Header) websocket.Connection); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(websocket.Connection)
		}
	}

	var r1 *http.Response
	if rf, ok := ret.Get(1).(func(string, http.Header) *http.Response); ok {
		r1 = rf(_a0, _a1)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*http.Response)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, http.Header) error); ok {
		r2 = rf(_a0, _a1)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
