package action

import (
	"fmt"

	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

const initialisationIndication = "Green"
const completionIndication = "linPEAS done"
const commandFormat = "curl https://raw.githubusercontent.com/carlospolop/" +
	"privilege-escalation-awesome-scripts-suite/master/linPEAS/" +
	"linpeas.sh | sh && echo '%s\n'\n"

// LinPeasCommand is a constant that holds the expected string for linpeas execution in
// a websocket.Command message.
const LinPeasCommand = "linpeas"

// LinPeas allows for the execution of LinPEAS. LinPEAS is a script
// to search for possible local privilege escalation paths
// https://github.com/carlospolop/privilege-escalation-awesome-scripts-suite
type LinPeas struct {
	Session         *pty.Session
	commandObserver CommandObserver
}

// Run runs LinPEAS on the machine
func (linPeas LinPeas) Run() *chan websocket.Command {
	go linPeas.commandObserver.WaitForCompletion(linPeas.Session)
	go linPeas.execute()
	return linPeas.commandObserver.CommandOutput()
}

func (linPeas LinPeas) execute() {
	linPeas.Session.Execute(
		fmt.Sprintf(commandFormat, completionIndication),
	)
}

// CreateLinPeas create and returns a fully initialised LinPeas action.
func CreateLinPeas(session *pty.Session) LinPeas {
	return LinPeas{
		Session:         session,
		commandObserver: CreateCommandObserver(initialisationIndication, completionIndication),
	}
}
