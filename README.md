# Apollo Agent
Agent meant to be deployed on a compromised machines to connect back to Apollo.

Apollo is a post-exploitation tool for managing, enumerating and pivotting on
compromised machines.

This app is only meant to be ethically used. Only use Apollo on systems you're
authorized to use.

## Installation & usage

### Development
During development you can use `go run` which compiles your code and runs the resulting binary.
Refer to *Running* to see the requirements arguments for the agent.

### Testing
I prefer to run tests while checking for race conditions and collecting coverage.
```
go test --race --cover --coverprofile cover.out ./...
```

### (Cross-)Compilation
To compile for your current OS:
```
go build -o build/apollo .
```

`-o` is optional and is used to set an output directory and filename.

To compile for a different OS and/or architecture you can use the `GOOS` and `GOARCH` env vars:
```
GOOS=linux GOARCH=amd64 go build -o build/apollo .
```

### Run
Depending on whether you're using `go run` or you're executing a manually compiled binary:
```
go run main.go --host <apollo_host> --agent-id <your_agent_id> --secret <your_client_secret>
```

or (on Linux and MacOS):
```
build/apollo --host <apollo_API_host> --agent-id <your_agent_id> --secret <your_client_secret>
```

The agent id and secret given out by the Apollo API.

Use `--help` for more info about the required and optional arguments.
