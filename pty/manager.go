package pty

import (
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// Manager helps managing multiple PTY sessions by finding or creating the
// correct session based on ShellCommunication.ConnectionID and handling execution.
type Manager struct {
	sessions map[string]*Session
	out      *chan websocket.ShellCommunication
}

// Execute executes the given command in a PTY session, reusing a session if
// if already exists.
func (manager *Manager) Execute(shellComm websocket.ShellCommunication) {
	pty := manager.sessions[shellComm.ConnectionID]

	if pty == nil {
        pty = manager.CreateNewSession(shellComm.ConnectionID)
	}

	go pty.Execute(shellComm.Message)
}

// CreateNewSession creates a new PTY session for the given ID,
// overwriting the existing session for this ID if present.
func (manager *Manager) CreateNewSession(sessionID string) *Session {
    pty := CreateSession(sessionID)
    manager.sessions[sessionID] = pty
    out := pty.Out()
    go manager.writeOutput(&out)
    return pty
}

func (manager *Manager) writeOutput(in *<-chan websocket.ShellCommunication) {
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
func CreateManager(out *chan websocket.ShellCommunication) Manager {
	return Manager{
		sessions: map[string]*Session{},
		out:      out,
	}
}
