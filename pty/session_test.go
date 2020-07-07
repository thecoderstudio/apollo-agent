package pty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/pty"
)

func TestCreateSession(t *testing.T) {
	pty, err := pty.CreateSession("test", "/bin/bash")
	defer pty.Close()

    assert.NoError(t, err)
	assert.Equal(t, pty.SessionID, "test")
	assert.NotNil(t, pty.Session())
	assert.NotNil(t, pty.Out())
}

func TestCreateSessionInvalidShell(t *testing.T) {
	pty, err := pty.CreateSession("test", "/bin/fake")
	defer pty.Close()

    assert.EqualError(t, err, "fork/exec /bin/fake: no such file or directory")
}

func TestExecuteEmptyCommand(t *testing.T) {
	pty, _ := pty.CreateSession("test", "/bin/bash")
	defer pty.Close()

	pty.Execute("")
	assert.Empty(t, pty.Out())
}

func TestExecute(t *testing.T) {
	shellsForTesting := []string{"/bin/bash", "/bin/zsh"}

	for _, shell := range shellsForTesting {
		t.Run(shell, func(t *testing.T) {
			pty, _ := pty.CreateSession("test", shell)
			defer pty.Close()

			pty.Execute("echo 1")

			outChan := pty.Out()
			output := <-outChan
			assert.Contains(t, output.Message, "echo 1")
			assert.NotNil(t, pty.Session())

			pty.Execute("echo 2")
			output = <-outChan
			assert.Contains(t, output.Message, "echo 2")
		})
	}
}

func TestExecuteOnClosed(t *testing.T) {
	pty, _ := pty.CreateSession("test", "/bin/bash")
	pty.Close()
	err := pty.Execute("echo 1")
	assert.EqualError(t, err, "session is closed, please create a new session")
}
