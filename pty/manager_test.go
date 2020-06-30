package pty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestCreateManager(t *testing.T) {
	out := make(chan websocket.ShellIO)
	manager := pty.CreateManager(&out)

	assert.NotNil(t, manager)

	manager.Close()
}

func TestManagerExecute(t *testing.T) {
	out := make(chan websocket.ShellIO)
	manager := pty.CreateManager(&out)

	manager.Execute(websocket.ShellIO{
		ConnectionID: "test",
		Message:      "echo 1",
	})
	first := <-out

	manager.Execute(websocket.ShellIO{
		ConnectionID: "test",
		Message:      "echo 2",
	})
	second := <-out

	assert.Contains(t, first.Message, "echo 1")
	assert.Contains(t, second.Message, "echo 2")

	manager.Close()
}
