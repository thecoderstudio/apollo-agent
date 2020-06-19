package shell

import (
    "bytes"
    "log"
    "os"
    "os/exec"
    "strings"

    "github.com/creack/pty"
)

type chanWriter struct {
	ch chan string
}

func newChanWriter(ch chan string) *chanWriter {
    return &chanWriter{ch}
}

func (w *chanWriter) Chan() <-chan string {
	return w.ch
}

func (w *chanWriter) Write(p []byte) (int, error) {
    w.ch <- string(p)
	return len(p), nil
}

func (w *chanWriter) Close() error {
	close(w.ch)
	return nil
}

type PTYSession struct {
    sessionID string
    session *os.File
    out chan string
}

func (ptySession *PTYSession) Execute(toBeExecuted string) {
    log.Println("called")
    if ptySession.session != nil {
        log.Println("not nil")
        ptySession.session.Write([]byte(toBeExecuted))
    } else {
        log.Println("nil")
        ptySession.createNewSession(toBeExecuted)
    }
}

func (ptySession *PTYSession) createNewSession(toBeExecuted string) {
    if len(toBeExecuted) == 0 {
        log.Println("0")
        return
    }
    commandAndArgs := strings.Fields(toBeExecuted)
    log.Println(commandAndArgs)
    cmd := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)

    chanWriter := newChanWriter(ptySession.out)

    var stderr bytes.Buffer
    cmd.Stdout = chanWriter
    cmd.Stderr = &stderr

    session, err := pty.Start(cmd)
    ptySession.session = session

    if err != nil {
        log.Println(err)
        log.Println(stderr.String())
    }
}

func (ptySession *PTYSession) listen() {
    for {
        output := <-ptySession.out
        log.Println(string(output))
    }
}

func CreateNewPTY(sessionID string) *PTYSession {
    ptySession := PTYSession{
        sessionID: sessionID,
        out: make(chan string),
    }
    go ptySession.listen()
    return &ptySession
}
