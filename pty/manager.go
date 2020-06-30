package pty

import (
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// NewConnection command to open a new connection and PTY session
const NewConnection = "new connection"

// Manager helps managing multiple PTY sessions by finding or creating the
// correct session based on ShellIO.ConnectionID and handling execution.
type Manager struct {
	sessions map[string]*Session
	out      *chan websocket.ShellIO
}

// ExecutePredefinedCommand executes the pre-defined command if it exists.
func (manager *Manager) ExecutePredefinedCommand(command websocket.Command) {
    if command.Command == NewConnection {
        manager.createNewSession(command.ConnectionID)
    }
}

// Execute send the given input to the PTY session, reusing a session if
// if already exists.
func (manager *Manager) Execute(shellIO websocket.ShellIO) {
	pty := manager.GetSession(shellIO.ConnectionID)

	if pty == nil {
        pty = manager.createNewSession(shellIO.ConnectionID)
	}

	go pty.Execute(shellIO.Message)
}

// GetSession returns the session for the given ID or nil if no such
// session exists.
func (manager *Manager) GetSession(sessionID string) *Session {
    return manager.sessions[sessionID]
}

func (manager *Manager) createNewSession(sessionID string) *Session {
    pty := CreateSession(sessionID)
    manager.sessions[sessionID] = pty
    out := pty.Out()
    go manager.writeOutput(&out)
    return pty
}

func (manager *Manager) writeOutput(in *<-chan websocket.ShellIO) {
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
func CreateManager(out *chan websocket.ShellIO) Manager {
	return Manager{
		sessions: map[string]*Session{},
		out:      out,
	}
}
