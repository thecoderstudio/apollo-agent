package action

import (
	"strings"

	"github.com/thecoderstudio/apollo-agent/logging"
	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// CommandObserver observes a command channel and tracks its command initialisation and completion.
type CommandObserver struct {
	InitialisationIndication string
	CompletionIndication     string
	commandOutput            chan websocket.Command
}

// CommandOutput returns all stdout output of the observer command.
func (commandObserver CommandObserver) CommandOutput() *chan websocket.Command {
	return &commandObserver.commandOutput
}

// WaitForCompletion observes the given PTY session and sends a finished command on completion.
func (commandObserver CommandObserver) WaitForCompletion(session pty.BaseSession) {
	out := make(chan interface{})
	broadcaster := *session.Out()
	broadcaster.Register(out)

	commandObserver.waitForInitialisation(out)
	for {
		if commandObserver.outputContains(out, commandObserver.CompletionIndication) {
			commandObserver.commandOutput <- websocket.Command{
				ConnectionID: session.SessionID(),
				Command:      "finished",
			}
			broadcaster.Unregister(out)
		}
	}
}

func (commandObserver CommandObserver) waitForInitialisation(out chan interface{}) {
	logging.Critical("waiting")
	for {
		if commandObserver.outputContains(out, commandObserver.InitialisationIndication) {
			return
		}
	}
}

func (commandObserver CommandObserver) outputContains(out chan interface{}, substring string) bool {
	outputGeneric := <-out
	output := outputGeneric.(websocket.ShellIO)
	logging.Critical(output.Message)
	return strings.Contains(output.Message, substring)
}

// CreateCommandObserver creates a CommandObserver initialised with the given arguments and a command output
// channel.
func CreateCommandObserver(initialisationIndication, completionIndication string) CommandObserver {
	commandOutput := make(chan websocket.Command)
	return CommandObserver{
		InitialisationIndication: initialisationIndication,
		CompletionIndication:     completionIndication,
		commandOutput:            commandOutput,
	}
}
