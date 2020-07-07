package pty

import (
	"log"

	"github.com/thecoderstudio/apollo-agent/websocket"
)

// NewConnection command to open a new connection and PTY session
const NewConnection = "new connection"

// Manager helps managing multiple PTY sessions by finding or creating the
// correct session based on ShellIO.ConnectionID and handling execution.
type Manager struct {
	Shell    string
	sessions map[string]*Session
	out      *chan websocket.ShellIO
	done      *chan bool
}

// ExecutePredefinedCommand executes the pre-defined command if it exists.
func (manager *Manager) ExecutePredefinedCommand(command websocket.Command) {
	if command.Command == NewConnection {
		manager.CreateNewSession(command.ConnectionID)
	}
}

// Execute send the given input to the PTY session, reusing a session if
// if already exists.
func (manager *Manager) Execute(shellIO websocket.ShellIO) error {
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
func (manager *Manager) GetSession(sessionID string) *Session {
	return manager.sessions[sessionID]
}

// CreateNewSession creates a new PTY session for the given ID,
// overwriting the existing session for this ID if present.
func (manager *Manager) CreateNewSession(sessionID string) (*Session, error) {
	pty, err := CreateSession(sessionID, manager.Shell)
	if err != nil {
        manager.writeError(sessionID, err)
		log.Println(err)
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
	*manager.out <- errMessage
}

func (manager *Manager) writeOutput(in *<-chan websocket.ShellIO) {
	for {
        select {
        case <-*manager.done:
            return
        default:
            message := <-*in
            *manager.out <- message
        }
	}
}

// Close closes all sessions.
func (manager *Manager) Close() {
    close(*manager.done)
	for _, pty := range manager.sessions {
		pty.Close()
	}
}

// CreateManager creates a Manager with the required out channel. All sessions will get created
// with the given shell.
func CreateManager(out *chan websocket.ShellIO, shell string) (Manager, error) {
	err := Verify(shell)
	var manager Manager
	if err != nil {
		return manager, err
	}

    done := make(chan bool)
	manager = Manager{
		Shell:    shell,
		sessions: map[string]*Session{},
		out:      out,
        done: &done,
	}

	return manager, err
}
