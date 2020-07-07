package pty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/pty"
)

func TestStartWithError(t *testing.T) {
	session, err := pty.Start("/bin/fake")
	defer session.Close()
	assert.Nil(t, session)
	assert.EqualError(t, err, "fork/exec /bin/fake: no such file or directory")
}

func TestVerifyValid(t *testing.T) {
	err := pty.Verify("/bin/bash")
	assert.NoError(t, err)
}

func TestVerifyInvalid(t *testing.T) {
	err := pty.Verify("/bin/fake")
	assert.EqualError(t, err, "fork/exec /bin/fake: no such file or directory")
}
