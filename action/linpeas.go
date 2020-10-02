package action

import (
	"strings"

	"github.com/thecoderstudio/apollo-agent/logging"
	"github.com/thecoderstudio/apollo-agent/pty"
)

// LinPeas allows for the execution of LinPEAS. LinPEAS is a script
// to search for possible local privilege escalation paths
// https://github.com/carlospolop/privilege-escalation-awesome-scripts-suite
type LinPeas struct {
	Session *pty.Session
}

// Run runs LinPEAS on the machine
func (linPeas LinPeas) Run() {
	go func() {
		started := false
		for {
			output := <-linPeas.Session.Out()
			if started && strings.Contains(output.Message, "linPEAS done") {
				logging.Critical("DONEE")
			} else if strings.Contains(output.Message, "ADVISORY") {
				logging.Critical("true")
				started = true
			}
		}
	}()
	go linPeas.Session.Execute(
		"curl https://raw.githubusercontent.com/carlospolop/" +
			"privilege-escalation-awesome-scripts-suite/master/linPEAS/" +
			"linpeas.sh | sh && echo 'linPEAS done\n'\n",
	)
}
