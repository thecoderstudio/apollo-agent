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
    Out *chan client.Message
}

// Execute executes toBeExecuted in the pty. Output is written to PTYSession.Out
func (ptySession *PTYSession) Execute(toBeExecuted string) error {
    if toBeExecuted == "" {
        return nil
    }

    var err error

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
        *ptySession.Out <- outMessage
    }
}

// CreateNewPTY creates a new PTYSession injected with the given sessionID and an output channel.
func CreateNewPTY(sessionID string) *PTYSession {
    out := make(chan client.Message)
    ptySession := PTYSession{
        SessionID: sessionID,
        Out: &out,
    }
    return &ptySession
}
