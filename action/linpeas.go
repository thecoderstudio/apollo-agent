package action

import (
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
	go linPeas.Session.Execute("curl https://raw.githubusercontent.com/carlospolop/privilege-escalation-awesome-scripts-suite/master/linPEAS/linpeas.sh | sh\n")
}
