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

func TestNewConnectionCommand(t *testing.T) {
	out := make(chan websocket.ShellIO)
	manager := pty.CreateManager(&out)

	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      pty.NewConnection,
	})

	assert.NotNil(t, manager.GetSession("test"))

	manager.Close()
}

func TestGetSession(t *testing.T) {
	out := make(chan websocket.ShellIO)
	manager := pty.CreateManager(&out)

	session := manager.CreateNewSession("test")

	assert.Equal(t, manager.GetSession("test"), session)

	manager.Close()

}

func TestGetSessionNotFound(t *testing.T) {
	out := make(chan websocket.ShellIO)
	manager := pty.CreateManager(&out)

	assert.Nil(t, manager.GetSession("test"))
}

func TestCreateNewSession(t *testing.T) {
	out := make(chan websocket.ShellIO)
	manager := pty.CreateManager(&out)

	assert.NotNil(t, manager.CreateNewSession("test"))
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
