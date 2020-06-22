package shell

import (
    "os"
    "os/exec"

    "github.com/creack/pty"

    "github.com/thecoderstudio/apollo-agent/client"
)

type PTYSession struct {
    sessionID string
    session *os.File
    Out *chan client.Message
}

func (ptySession *PTYSession) Execute(toBeExecuted string) {
    if toBeExecuted == "" {
        return
    }

    if ptySession.session == nil {
        ptySession.createNewSession()
    }
    ptySession.session.Write([]byte(toBeExecuted))
}

func (ptySession *PTYSession) createNewSession() {
    cmd := exec.Command("/bin/bash")
    session, _ := pty.Start(cmd)

    ptySession.session = session
    go ptySession.listen(ptySession.session)
}

func (ptySession *PTYSession) listen(session *os.File) {
    for {
        buf := make([]byte, 512)
        session.Read(buf)

        outMessage := client.Message {
            SessionID: ptySession.sessionID,
            Message: string(buf),
        }
        *ptySession.Out <- outMessage
    }
}

func CreateNewPTY(sessionID string) *PTYSession {
    out := make(chan client.Message)
    ptySession := PTYSession{
        sessionID: sessionID,
        Out: &out,
    }
    return &ptySession
}
