package pty

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// Start starts and returns an open PTY session with the given shell command.
func Start(shell string) (*os.File, error) {
	cmd := exec.Command(shell)
	return pty.Start(cmd)
}

// Verify verifies the given shell command, returning an error if invalid.
func Verify(shell string) error {
	session, err := Start(shell)
	session.Close()
	return err
}
