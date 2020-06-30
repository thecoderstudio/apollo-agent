package pty

import (
	"errors"
    "log"
	"os"
	"os/exec"

	"github.com/creack/pty"

	"github.com/thecoderstudio/apollo-agent/websocket"
)

// Session is a PTY that allows command execution through Session.Execute while sending output
// through the Session.Out channel.
type Session struct {
	SessionID string
	session   *os.File
	out       *chan websocket.ShellIO
	closed    bool
}

// Session returns the inner pty.
func (ptySession *Session) Session() *os.File {
	return ptySession.session
}

// Out returns a read-only channel used for communicating output to command
// execution in the PTY.
func (ptySession *Session) Out() <-chan websocket.ShellIO {
	return *ptySession.out
}

// Execute executes toBeExecuted in the pty. Output is written to Session.Out.
func (ptySession *Session) Execute(toBeExecuted string) error {
	var err error

	if ptySession.closed {
		return errors.New("session is closed, please create a new session")
	}

	if toBeExecuted == "" {
		return err
	}

	_, err = ptySession.session.Write([]byte(toBeExecuted))
	return err
}

func (ptySession *Session) createNewSession() error {
	cmd := exec.Command("/bin/bash")
	session, err := pty.Start(cmd)

	ptySession.session = session
	go ptySession.listen(ptySession.session)

	return err
}

func (ptySession *Session) listen(session *os.File) {
	for {
		buf := make([]byte, 512)
		session.Read(buf)

		outComm := websocket.ShellIO{
			ConnectionID: ptySession.SessionID,
			Message:      string(buf),
		}
        log.Println(outComm.ConnectionID)
        log.Println(outComm.Message)
		if !ptySession.closed {
			*ptySession.out <- outComm
		}
	}
}

// Close closes out channel and pty
func (ptySession *Session) Close() {
	ptySession.closed = true
	if ptySession.session != nil {
		ptySession.session.Close()
	}
	close(*ptySession.out)
}

// CreateSession creates a new Session injected with the given sessionID and defaults.
func CreateSession(sessionID string) *Session {
	out := make(chan websocket.ShellIO)
	ptySession := Session{
		SessionID: sessionID,
		out:       &out,
		closed:    false,
	}
    ptySession.createNewSession()
	return &ptySession
}
