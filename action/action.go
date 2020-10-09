package action

import (
	"fmt"

	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// Action is an interface that allows for the definition of
// pre-defined shell commands.
type Action interface {
	Run() *chan websocket.Command
	execute()
}

// Execute executes the action for the given command and returns its output. If no action is found Execute will
// return an error.
func Execute(session pty.BaseSession, command websocket.Command) (*chan websocket.Command, error) {
	var action Action
	var err error

	switch command.Command {
	case LinPeasCommand:
		action = CreateLinPeas(session)
	default:
		return nil, fmt.Errorf("action not found for given command: %s", command.Command)
	}

	return action.Run(), err
}
