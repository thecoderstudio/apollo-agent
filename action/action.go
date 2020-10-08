package action

import "github.com/thecoderstudio/apollo-agent/websocket"

// Action is an interface that allows for the definition of
// pre-defined shell commands.
type Action interface {
	Run() *chan websocket.Command
	execute()
}
