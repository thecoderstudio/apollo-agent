package shell

import (
    "os"
    "os/exec"

    "github.com/creack/pty"
)

type PTYSession struct {
    sessionID string
    session *os.File
    Out chan string
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
        ptySession.Out <- string(buf)
    }
}

func CreateNewPTY(sessionID string) *PTYSession {
    ptySession := PTYSession{
        sessionID: sessionID,
        Out: make(chan string),
    }
    return &ptySession
}
