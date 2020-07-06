package pty

import (
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"

	"github.com/thecoderstudio/apollo-agent/websocket"
)

// Session is a PTY that allows command execution through Session.Execute while sending output
// through the Session.Out channel.
type Session struct {
	SessionID string
	shell     string
	session   *os.File
	out       *chan websocket.ShellIO
	done      *chan bool
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
	cmd := exec.Command(ptySession.shell)
	session, err := pty.Start(cmd)

	ptySession.session = session
	go ptySession.listen(ptySession.session)

	return err
}

func (ptySession *Session) listen(session *os.File) {
	ticker := time.NewTicker(100 * time.Nanosecond)
	defer ticker.Stop()
	for {
		select {
		case <-*ptySession.done:
			ptySession.closeSession()
			return
		case <-ticker.C:
			buf := make([]byte, 512)
			session.Read(buf)

			outComm := websocket.ShellIO{
				ConnectionID: ptySession.SessionID,
				Message:      string(buf),
			}
			*ptySession.out <- outComm
		}
	}
}

// Close schedules the session for closure
func (ptySession *Session) Close() {
	ptySession.closed = true
	*ptySession.done <- true
}

func (ptySession *Session) closeSession() {
	if ptySession.session != nil {
		ptySession.session.Close()
	}
	close(*ptySession.out)
}

// CreateSession creates a new Session injected with the given sessionID, the given shell and defaults.
func CreateSession(sessionID, shell string) *Session {
	out := make(chan websocket.ShellIO)
	done := make(chan bool)
	ptySession := Session{
		SessionID: sessionID,
		shell:     shell,
		out:       &out,
		done:      &done,
		closed:    false,
	}
	ptySession.createNewSession()
	return &ptySession
}
