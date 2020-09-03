package pty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestCreateManager(t *testing.T) {
	manager, err := pty.CreateManager("/bin/bash")

	assert.NotNil(t, manager)
	assert.NoError(t, err)

	manager.Close()
}

func TestCreateManagerInvalidShell(t *testing.T) {
	_, err := pty.CreateManager("/bin/fake")

	assert.EqualError(t, err, "fork/exec /bin/fake: no such file or directory")
}

func TestNewConnectionCommand(t *testing.T) {
	manager, _ := pty.CreateManager("/bin/bash")

	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      pty.NewConnection,
	})

	assert.NotNil(t, manager.GetSession("test"))

	manager.Close()
}

func TestGetSession(t *testing.T) {
	manager, _ := pty.CreateManager("/bin/bash")

	session, _ := manager.CreateNewSession("test")

	assert.Equal(t, manager.GetSession("test"), session)

	manager.Close()
}

func TestGetSessionNotFound(t *testing.T) {
	manager, _ := pty.CreateManager("/bin/bash")

	assert.Nil(t, manager.GetSession("test"))
}

func TestCreateNewSession(t *testing.T) {
	manager, _ := pty.CreateManager("/bin/bash")
	session, err := manager.CreateNewSession("test")

	assert.NotNil(t, session)
	assert.NoError(t, err)

	manager.Close()
}

func TestCreateNewSessionInvalidShell(t *testing.T) {
	expectedErrMessage := "fork/exec /bin/fake: no such file or directory"
	manager, _ := pty.CreateManager("/bin/bash")
	manager.Shell = "/bin/fake"

	go func() {
		session, err := manager.CreateNewSession("test")
		assert.Nil(t, session)
		assert.EqualError(t, err, expectedErrMessage)
	}()

	writtenErr := <-manager.Out()
	assert.Equal(t, writtenErr, websocket.ShellIO{
		ConnectionID: "test",
		Message:      expectedErrMessage,
	})

	manager.Close()
}

func TestManagerExecute(t *testing.T) {
	manager, _ := pty.CreateManager("/bin/bash")

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

func TestManagerExecuteInvalidShell(t *testing.T) {
	expectedErrMessage := "fork/exec /bin/fake: no such file or directory"
	manager, _ := pty.CreateManager("/bin/bash")
	manager.Shell = "/bin/fake"

	go func() {
		err := manager.Execute(websocket.ShellIO{
			ConnectionID: "test",
			Message:      "echo 1",
		})
		assert.EqualError(t, err, expectedErrMessage)
	}()
	writtenErr := <-manager.Out()

	assert.Equal(t, writtenErr, websocket.ShellIO{
		ConnectionID: "test",
		Message:      expectedErrMessage,
	})

	manager.Close()
}
