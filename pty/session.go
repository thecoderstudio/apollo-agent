package pty

import (
	"bufio"
	"errors"
	"os"

	"github.com/dustin/go-broadcast"

	"github.com/thecoderstudio/apollo-agent/websocket"
)

// BaseSession allows communcation with a PTY session.
type BaseSession interface {
	SessionID() string
	Session() *os.File
	Out() *broadcast.Broadcaster
	Execute(string) error
	Close()
}

// Session is a PTY that allows command execution through Session.Execute while sending output
// through the Session.Out channel.
type Session struct {
	sessionID string
	shell     string
	session   *os.File
	out       broadcast.Broadcaster
	done      *chan bool
	closed    bool
}

// SessionID returns the sessionID
func (ptySession *Session) SessionID() string {
	return ptySession.sessionID
}

// Session returns the inner pty.
func (ptySession *Session) Session() *os.File {
	return ptySession.session
}

// Out returns a broadcaster used for communicating output to command
// execution in the PTY.
func (ptySession *Session) Out() *broadcast.Broadcaster {
	return &ptySession.out
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
	session, err := Start(ptySession.shell)

	ptySession.session = session
	go ptySession.listen(ptySession.session)

	return err
}

func (ptySession *Session) listen(session *os.File) {
	reader := bufio.NewReader(session)
	for {
		select {
		case <-*ptySession.done:
			ptySession.closeSession()
			return
		default:
			buf := make([]byte, 512)
			reader.Read(buf)

			outComm := websocket.ShellIO{
				ConnectionID: ptySession.SessionID(),
				Message:      string(buf),
			}
			ptySession.out.Submit(outComm)
		}
	}
}

// Close schedules the session for closure
func (ptySession *Session) Close() {
	ptySession.closed = true
	close(*ptySession.done)
}

func (ptySession *Session) closeSession() {
	if ptySession.session != nil {
		ptySession.session.Close()
	}
}

// CreateSession creates a new Session injected with the given sessionID, the given shell and defaults.
func CreateSession(sessionID, shell string) (*Session, error) {
	out := broadcast.NewBroadcaster(512)
	done := make(chan bool)
	ptySession := Session{
		sessionID: sessionID,
		shell:     shell,
		out:       out,
		done:      &done,
		closed:    false,
	}
	err := ptySession.createNewSession()
	return &ptySession, err
}
