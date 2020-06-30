package websocket

// ShellIO is used for communicating shell input, output and error streams.
type ShellIO struct {
	ConnectionID string `json:"connection_id"`
	Message      string `json:"message"`
}

// Command is used to instruct the agent to execute a pre-defined command.
type Command struct {
    ConnectionID    string `json:"connection_id"`
    Command         string `json:"command"`
}

