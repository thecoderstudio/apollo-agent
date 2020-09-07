package pty

import (
	"github.com/thecoderstudio/apollo-agent/logging"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// NewConnection command to open a new connection and PTY session
const NewConnection = "new connection"

// ShellManager is an interface that allows for PTY session
// management and command execution.
type ShellManager interface {
	Out() <-chan websocket.ShellIO
	ExecutePredefinedCommand(websocket.Command)
	Execute(websocket.ShellIO) error
	GetSession(string) *Session
	CreateNewSession(string) (*Session, error)
	Close()
}

// Manager helps managing multiple PTY sessions by finding or creating the
// correct session based on ShellIO.ConnectionID and handling execution.
type Manager struct {
	Shell    string
	sessions map[string]*Session
	out      chan websocket.ShellIO
}

// Out returns all output of the PTY session(s) through a channel.
func (manager Manager) Out() <-chan websocket.ShellIO {
	return manager.out
}

// ExecutePredefinedCommand executes the pre-defined command if it exists.
func (manager Manager) ExecutePredefinedCommand(command websocket.Command) {
	if command.Command == NewConnection {
		manager.CreateNewSession(command.ConnectionID)
	}
}

// Execute send the given input to the PTY session, reusing a session if
// if already exists.
func (manager Manager) Execute(shellIO websocket.ShellIO) error {
	pty := manager.GetSession(shellIO.ConnectionID)

	if pty == nil {
		newPty, err := manager.CreateNewSession(shellIO.ConnectionID)
		if err != nil {
			return err
		}
		pty = newPty
	}

	go pty.Execute(shellIO.Message)
	return nil
}

// GetSession returns the session for the given ID or nil if no such
// session exists.
func (manager Manager) GetSession(sessionID string) *Session {
	return manager.sessions[sessionID]
}

// CreateNewSession creates a new PTY session for the given ID,
// overwriting the existing session for this ID if present.
func (manager Manager) CreateNewSession(sessionID string) (*Session, error) {
	pty, err := CreateSession(sessionID, manager.Shell)
	if err != nil {
		manager.writeError(sessionID, err)
		logging.Err(err)
		pty.Close()
		return nil, err
	}

	manager.sessions[sessionID] = pty
	out := pty.Out()
	go manager.writeOutput(&out)
	return pty, err
}

func (manager *Manager) writeError(sessionID string, err error) {
	errMessage := websocket.ShellIO{
		ConnectionID: sessionID,
		Message:      err.Error(),
	}
	manager.out <- errMessage
}

func (manager *Manager) writeOutput(in *<-chan websocket.ShellIO) {
	for {
		message := <-*in
		manager.out <- message
	}
}

// Close closes all sessions.
func (manager Manager) Close() {
	for _, pty := range manager.sessions {
		pty.Close()
	}
}

// CreateManager creates a Manager with the required out channel. All sessions will get created
// with the given shell.
func CreateManager(shell string) (Manager, error) {
	err := Verify(shell)
	var manager Manager
	if err != nil {
		return manager, err
	}

	out := make(chan websocket.ShellIO)
	manager = Manager{
		Shell:    shell,
		sessions: map[string]*Session{},
		out:      out,
	}

	return manager, err
}
