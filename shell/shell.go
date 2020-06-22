package shell

import (
    "log"
    "io"
    "os"
    "os/exec"
    "strings"
    "syscall"

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

type chanReader struct {
	ch chan string
}

func newChanReader(ch chan string) *chanReader {
    return &chanReader{ch}
}

func (r *chanReader) Chan() <-chan string {
	return r.ch
}

func (r *chanReader) Read(p []byte) (int, error) {
    r.ch <- string(p)
	return len(p), nil
}

func (r *chanReader) Close() error {
	close(r.ch)
	return nil
}


type PTYSession struct {
    sessionID string
    session *os.File
    out chan string
    err chan string
    in *io.Reader
}

func (ptySession *PTYSession) Execute(toBeExecuted string) {
    log.Println("called")
    if ptySession.session != nil {
        log.Println("not nil")
        in := *ptySession.in
        in.Read([]byte(toBeExecuted))
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

    stdout := newChanWriter(ptySession.out)
    stderr := newChanWriter(ptySession.err)

    cmd.Stdout = stdout
    cmd.Stderr = stderr
    if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
	cmd.SysProcAttr.Setctty = true

    session, tty, err := pty.Open()

    cmd.Stdin = tty

    ptySession.session = session
    ptySession.in = &cmd.Stdin

    cmd.Start()

    if err != nil {
        log.Println(err)
    }
}

func (ptySession *PTYSession) listen() {
    for {
        select {
            case output := <-ptySession.out:
                log.Println("output")
                log.Println(string(output))
            case err := <-ptySession.err:
                log.Println("err")
                log.Println(string(err))
        }
    }
}

func CreateNewPTY(sessionID string) *PTYSession {
    ptySession := PTYSession{
        sessionID: sessionID,
        out: make(chan string),
        err: make(chan string),
    }
    go ptySession.listen()
    return &ptySession
}
