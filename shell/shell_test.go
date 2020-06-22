package shell_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/thecoderstudio/apollo-agent/shell"
)

func TestCreateNewPTY(t *testing.T) {
    pty := shell.CreateNewPTY("test")
    assert.Equal(t, pty.SessionID, "test")
    assert.NotNil(t, pty.Out)
}

func TestExecuteEmptyCommand(t *testing.T) {
    pty := shell.CreateNewPTY("test")
    pty.Execute("")
    assert.Nil(t, pty.Session())
}
