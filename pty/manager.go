package pty

import (
    "github.com/thecoderstudio/apollo-agent/client"
)

// Manager helps managing multiple PTY sessions by finding or creating the
// correct session based on Message.ConnectionID and handling execution.
type Manager struct {
    sessions map[string]*Session
    out *chan client.Message
}

// Execute executes the given command in a PTY session, reusing a session if
// if already exists.
func (manager *Manager) Execute(message client.Message) {
    pty := manager.sessions[message.ConnectionID]

    if pty == nil {
        pty = CreateSession(message.ConnectionID)
        manager.sessions[message.ConnectionID] = pty
        out := pty.Out()
        go manager.writeOutput(&out)
    }

    pty.Execute(message.Message)
}

func (manager *Manager) writeOutput(in *<-chan client.Message) {
    for {
        message := <-*in
        *manager.out <- message
    }
}

// Close closes all sessions.
func (manager *Manager) Close() {
    for _, pty := range manager.sessions {
        pty.Close()
    }
}

// CreateManager creates a Manager with the required out channel.
func CreateManager(out *chan client.Message) Manager {
    return Manager { 
        sessions: map[string]*Session{},
        out: out,
    }
}
