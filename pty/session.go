package pty

import (
    "errors"
    "os"
    "os/exec"

    "github.com/creack/pty"

    "github.com/thecoderstudio/apollo-agent/client"
)

// Session is a PTY that allows command execution through Session.Execute while sending output
// through the Session.Out channel.
type Session struct {
    SessionID string
    session *os.File
    out *chan client.Message
    closed bool
}

// Session returns the inner pty.
func (ptySession *Session) Session() *os.File {
    return ptySession.session
}

// Out returns a read-only channel used for communicating output to command
// execution in the PTY.
func (ptySession *Session) Out() <-chan client.Message {
    return *ptySession.out
}

// Execute executes toBeExecuted in the pty. Output is written to Session.Out.
func (ptySession *Session) Execute(toBeExecuted string) error {
    var err error

    if ptySession.closed {
        return errors.New("Session is closed, please create a new session.")
    }

    if toBeExecuted == "" {
        return err
    }

    if ptySession.session == nil {
        err = ptySession.createNewSession()
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
    buf := make([]byte, 512)

    for {
        session.Read(buf)

        outMessage := client.Message {
            ConnectionID: ptySession.SessionID,
            Message: string(buf),
        }
        if !ptySession.closed {
            *ptySession.out <- outMessage
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
    out := make(chan client.Message)
    ptySession := Session{
        SessionID: sessionID,
        out: &out,
        closed: false,
    }
    return &ptySession
}
