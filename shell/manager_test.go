package shell_test

import (
	"testing"

	"github.com/eapache/channels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/thecoderstudio/apollo-agent/action"
	"github.com/thecoderstudio/apollo-agent/mocks"
	"github.com/thecoderstudio/apollo-agent/shell"
	"github.com/thecoderstudio/apollo-agent/testutil"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestCreateManager(t *testing.T) {
	manager, err := shell.CreateManager("/bin/bash", action.Executor{})

	assert.NotNil(t, manager)
	assert.NoError(t, err)

	manager.Close()
}

func TestCreateManagerInvalidShell(t *testing.T) {
	_, err := shell.CreateManager("/bin/fake", action.Executor{})

	assert.EqualError(t, err, "fork/exec /bin/fake: no such file or directory")
}

func TestNewConnectionCommand(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})

	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      shell.NewConnection,
	})

	assert.NotNil(t, manager.GetSession("test"))

	manager.Close()
}

func TestGetSession(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})

	session, _ := manager.CreateNewSession("test")

	assert.Equal(t, manager.GetSession("test"), session)

	manager.Close()
}

func TestGetSessionNotFound(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})

	assert.Nil(t, manager.GetSession("test"))
}

func TestCreateNewSession(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	session, err := manager.CreateNewSession("test")

	assert.NotNil(t, session)
	assert.NoError(t, err)

	manager.Close()
}

func TestCreateNewSessionInvalidShell(t *testing.T) {
	expectedErrMessage := "fork/exec /bin/fake: no such file or directory"
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	manager.Shell = "/bin/fake"

	go func() {
		session, err := manager.CreateNewSession("test")
		assert.Nil(t, session)
		assert.EqualError(t, err, expectedErrMessage)
	}()

	writtenErr := <-manager.Out()
	assert.Equal(t, writtenErr, websocket.ShellIO{
		ConnectionID: "test",
		Message:      expectedErrMessage,
	})

	manager.Close()
}

func TestManagerExecute(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})

	manager.Execute(websocket.ShellIO{
		ConnectionID: "test",
		Message:      "echo 1\n",
	})

	// Depending on the test environment there may be garbage before the initial echo.
	testutil.BlockUntilContains(
		channels.Wrap(manager.Out()).Out(),
		func(output interface{}) string { return output.(websocket.ShellIO).Message },
		"echo 1",
	)

	manager.Execute(websocket.ShellIO{
		ConnectionID: "test",
		Message:      "echo 2\n",
	})
	testutil.BlockUntilContains(
		channels.Wrap(manager.Out()).Out(),
		func(output interface{}) string { return output.(websocket.ShellIO).Message },
		"echo 2",
	)

	manager.Close()
}

func TestManagerExecuteInvalidShell(t *testing.T) {
	expectedErrMessage := "fork/exec /bin/fake: no such file or directory"
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	manager.Shell = "/bin/fake"

	go func() {
		err := manager.Execute(websocket.ShellIO{
			ConnectionID: "test",
			Message:      "echo 1",
		})
		assert.EqualError(t, err, expectedErrMessage)
	}()
	writtenErr := <-manager.Out()

	assert.Equal(t, writtenErr.(websocket.ShellIO), websocket.ShellIO{
		ConnectionID: "test",
		Message:      expectedErrMessage,
	})
}

func TestManagerCancelSession(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      shell.NewConnection,
	})

	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      shell.Cancel,
	})

	// TODO assert closure
	session := manager.GetSession("test")
	assert.Nil(t, session)
}

func TestManagerCancelFakeSession(t *testing.T) {
	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      shell.Cancel,
	})
}

func TestManagerExecuteAction(t *testing.T) {
	successCommand := websocket.Command{
		ConnectionID: "test",
		Command:      "success",
	}
	out := make(chan websocket.Command)
	mockExecutor := new(mocks.CommandExecutor)
	mockExecutor.On(
		"Execute",
		mock.Anything,
		mock.MatchedBy(func(command websocket.Command) bool { return command.Command == "fake" }),
	).Return(&out, nil).Run(func(args mock.Arguments) {
		go func() {
			out <- successCommand
		}()
	})

	manager, _ := shell.CreateManager("/bin/bash", mockExecutor)
	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      shell.NewConnection,
	})

	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      "fake",
	})

	for {
		receivedCommand := <-manager.Out()
		if receivedCommand == successCommand {
			return
		}
	}
}

func TestManagerExecuteUnknownAction(t *testing.T) {
	unknownActionMessage := websocket.ShellIO{
		ConnectionID: "test",
		Message:      "action not found for given command: fake",
	}

	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	manager.ExecutePredefinedCommand(websocket.Command{
		ConnectionID: "test",
		Command:      shell.NewConnection,
	})

	go func() {
		manager.ExecutePredefinedCommand(websocket.Command{
			ConnectionID: "test",
			Command:      "fake",
		})
	}()

	for {
		message := <-manager.Out()
		if message.(websocket.ShellIO).Message == unknownActionMessage.Message {
			return
		}
	}
}

func TestManagerExecuteOnFakeSession(t *testing.T) {
	notFoundMessage := websocket.ShellIO{
		ConnectionID: "test",
		Message:      "PTYSession not found",
	}

	manager, _ := shell.CreateManager("/bin/bash", action.Executor{})
	go func() {
		manager.ExecutePredefinedCommand(websocket.Command{
			ConnectionID: "test",
			Command:      "fake",
		})
	}()

	message := <-manager.Out()
	assert.Equal(t, message.(websocket.ShellIO).Message, notFoundMessage.Message)
}
