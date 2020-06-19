package shell

import (
    "bytes"
    "log"
    "os"
    "os/exec"
    "strings"

    "github.com/creack/pty"

    "github.com/thecoderstudio/apollo-agent/client"
)

type chanWriter struct {
	Chan chan string
}

func newChanWriter() *chanWriter {
	return &chanWriter{make(chan string)}
}

func (w *chanWriter) Write(p []byte) (int, error) {
    w.ch <- string(p)
	return len(p), nil
}

func (w *chanWriter) Close() error {
	close(w.ch)
	return nil
}

var sessions = map[string]*os.File{}

// Execute allows you to run shell commands on the current system and get their
// result.
func Execute(commandMessage client.Message) {
    toBeExecuted := commandMessage.Message
    if len(toBeExecuted) == 0 {
        return
    }
    commandAndArgs := strings.Fields(toBeExecuted)
    log.Println(commandAndArgs)
    cmd := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)

    out := newChanWriter()
    go func() {
        for {
            output := <-out.Chan
            log.Println(string(output))
        }
    }()


	var stderr bytes.Buffer
	cmd.Stdout = out
	cmd.Stderr = &stderr

    ptySession, err := pty.Start(cmd)
    ptySession.Write([]byte{4})

    if err != nil {
        log.Println(err)
        log.Println(stderr.String())
    }
}
