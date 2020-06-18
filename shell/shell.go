package shell

import (
    "bytes"
    "log"
    "strings"
    "os/exec"
)

// Execute allows you to run shell commands on the current system and get their
// result.
func Execute(toBeExecuted string) string {
    if len(toBeExecuted) == 0 {
        log.Println("empty")
        return ""
    }
    commandAndArgs := strings.Fields(toBeExecuted)
    command := commandAndArgs[0]
    commandAndArgs[0] = ""
    log.Println(command + "test")
    log.Println(commandAndArgs)
    cmd := exec.Command(command, commandAndArgs...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        log.Println(err)
        log.Println(stderr.String())
    }
    return out.String()
}
