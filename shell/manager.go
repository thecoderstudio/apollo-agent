package shell

import (
    "github.com/thecoderstudio/apollo-agent/client"
)

// PTYManager helps managing multiple PTY sessions by finding or creating the
// correct session based on Message.SessionID and handling execution.
type PTYManager struct {
    sessions map[string]*PTYSession
    out *chan client.Message
}

// Execute executes the given command in a PTY session, reusing a session if
// if already exists.
func (manager *PTYManager) Execute(message client.Message) {
    pty := manager.sessions[message.SessionID]

    if pty == nil {
        pty = CreateNewPTY(message.SessionID)
        manager.sessions[message.SessionID] = pty
        go manager.writeOutput(pty.Out)
    }

    pty.Execute(message.Message)
}

func (manager *PTYManager) writeOutput(in *chan client.Message) {
    for {
        message := <-*in
        *manager.out <- message
    }
}

// Close closes all sessions.
func (manager *PTYManager) Close() {
    for _, pty := range manager.sessions {
        pty.Close()
    }
}

// CreateManager creates a PTYManager with the required out channel.
func CreateManager(out *chan client.Message) PTYManager {
    return PTYManager { out: out }
}
