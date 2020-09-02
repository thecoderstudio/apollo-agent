package pty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestCreateManager(t *testing.T) {
	manager := pty.CreateManager("/bin/bash")

	assert.NotNil(t, manager)

	manager.Close()
}

func TestNewConnectionCommand(t *testing.T) {
	manager := pty.CreateManager("/bin/bash")

	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      pty.NewConnection,
	})

	assert.NotNil(t, manager.GetSession("test"))

	manager.Close()
}

func TestGetSession(t *testing.T) {
	manager := pty.CreateManager("/bin/bash")

	session := manager.CreateNewSession("test")

	assert.Equal(t, manager.GetSession("test"), session)

	manager.Close()
}

func TestGetSessionNotFound(t *testing.T) {
	manager := pty.CreateManager("/bin/bash")

	assert.Nil(t, manager.GetSession("test"))
}

func TestCreateNewSession(t *testing.T) {
	manager := pty.CreateManager("/bin/bash")

	assert.NotNil(t, manager.CreateNewSession("test"))
}

func TestManagerExecute(t *testing.T) {
	manager := pty.CreateManager("/bin/bash")

	manager.Execute(websocket.ShellIO{
		ConnectionID: "test",
		Message:      "echo 1",
	})
	first := <-manager.Out()

	manager.Execute(websocket.ShellIO{
		ConnectionID: "test",
		Message:      "echo 2",
	})
	second := <-manager.Out()

	assert.Contains(t, first.Message, "echo 1")
	assert.Contains(t, second.Message, "echo 2")

	manager.Close()
}
