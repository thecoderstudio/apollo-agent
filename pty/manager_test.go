package pty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/client"
	"github.com/thecoderstudio/apollo-agent/pty"
)

func TestCreateManager(t *testing.T) {
	out := make(chan client.Message)
	manager := pty.CreateManager(&out)

	assert.NotNil(t, manager)

	manager.Close()
}

func TestManagerExecute(t *testing.T) {
	out := make(chan client.Message)
	manager := pty.CreateManager(&out)

	manager.Execute(client.Message{
		ConnectionID: "test",
		Message:      "echo 1",
	})
	first := <-out

	manager.Execute(client.Message{
		ConnectionID: "test",
		Message:      "echo 2",
	})
	second := <-out

	assert.Contains(t, first.Message, "echo 1")
	assert.Contains(t, second.Message, "echo 2")

	manager.Close()
}
