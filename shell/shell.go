package shell

import (
    "os"
    "os/exec"

    "github.com/creack/pty"

    "github.com/thecoderstudio/apollo-agent/client"
)

// PTYSession is a PTY that allows command execution through PTYSession.Execute while sending output
// through the PTYSession.Out channel.
type PTYSession struct {
    SessionID string
    session *os.File
    out *chan client.Message
}

// Session returns the inner pty.
func (ptySession *PTYSession) Session() *os.File {
    return ptySession.session
}

// Out returns a read-only channel used for communicating output to command
// execution in the PTY.
func (ptySession *PTYSession) Out() <-chan client.Message {
    return *ptySession.out
}

// Execute executes toBeExecuted in the pty. Output is written to PTYSession.Out.
func (ptySession *PTYSession) Execute(toBeExecuted string) error {
    var err error

    if toBeExecuted == "" {
        return err
    }

    if ptySession.session == nil {
        err = ptySession.createNewSession()
    }
    _, err = ptySession.session.Write([]byte(toBeExecuted))
    return err
}

func (ptySession *PTYSession) createNewSession() error {
    cmd := exec.Command("/bin/bash")
    session, err := pty.Start(cmd)

    ptySession.session = session
    go ptySession.listen(ptySession.session)

    return err
}

func (ptySession *PTYSession) listen(session *os.File) {
    for {
        buf := make([]byte, 512)
        session.Read(buf)

        outMessage := client.Message {
            SessionID: ptySession.SessionID,
            Message: string(buf),
        }
        *ptySession.out <- outMessage
    }
}

// Close closes out channel and pty
func (ptySession *PTYSession) Close() {
    if ptySession.session != nil {
        ptySession.session.Close()
    }
    close(*ptySession.out)
}

// CreateNewPTY creates a new PTYSession injected with the given sessionID and an output channel.
func CreateNewPTY(sessionID string) *PTYSession {
    out := make(chan client.Message)
    ptySession := PTYSession{
        SessionID: sessionID,
        out: &out,
    }
    return &ptySession
}
