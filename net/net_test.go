package net_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/net"
)

func TestGetHostFromURLString(t *testing.T) {
	urls := []string{
		"https://localhost:1234/test/123",
		"http://localhost:1234",
		"wss://localhost:1234/abc",
		"ws://localhost:1234/abc",
		"fake://localhost:1234//",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			host := net.GetHostFromURLString(url)
			assert.Equal(t, host, "localhost:1234")
		})
	}
}
