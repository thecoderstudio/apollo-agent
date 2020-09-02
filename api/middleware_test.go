package api_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/mocks"
	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestMiddleware(t *testing.T) {

	accessTokenChan := make(chan oauth.AccessToken)
	authErrs := make(chan error)
	authProviderMock := new(mocks.AuthProvider)
	authProviderMock.On("GetContinuousAccessToken").Return(&accessTokenChan, &authErrs)

	accessToken := oauth.AccessToken{
		AccessToken: "",
		ExpiresIn:   3600,
		TokenType:   "",
	}
	done, readOnlyDone := createDoneChannels()
	out, readOnlyOut := createShellIOChannels()
	commands, readOnlyCommands := createCommandChannels()
	shellErrs := make(<-chan error)

	shellInterfaceMock := createShellInterfaceMock(readOnlyDone, readOnlyOut, readOnlyCommands, shellErrs)

	shellIO := websocket.ShellIO{ConnectionID: "1", Message: "echo 'test'"}
	command := websocket.Command{ConnectionID: "1", Command: "test"}
	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("ExecutePredefinedCommand", command)
	shellManagerMock.On("Execute", shellIO).Run(func(arguments mock.Arguments) {
		close(done)
	})

	interruptSignal := make(chan os.Signal, 1)
	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		ShellInterface:  shellInterfaceMock,
		PTYManager:      shellManagerMock,
		OAuthClient:     authProviderMock,
	}

	notify := make(chan struct{})
	go func() {
		middleware.Start()
		close(notify)
	}()
	accessTokenChan <- accessToken
	commands <- command
	out <- shellIO
	<-notify
	shellManagerMock.AssertExpectations(t)
}

func createShellInterfaceMock(
	done <-chan struct{},
	out <-chan websocket.ShellIO,
	commands <-chan websocket.Command,
	errs <-chan error,
) *mocks.ShellInterface {
	shellInterfaceMock := new(mocks.ShellInterface)
	shellInterfaceMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(done)
	shellInterfaceMock.On("Out").Return(out)
	shellInterfaceMock.On("Commands").Return(commands)
	shellInterfaceMock.On("Errs").Return(errs)
	return shellInterfaceMock
}

func createDoneChannels() (chan struct{}, <-chan struct{}) {
	done := make(chan struct{})
	return done, done
}

func createShellIOChannels() (chan websocket.ShellIO, <-chan websocket.ShellIO) {
	shellIOChannel := make(chan websocket.ShellIO)
	return shellIOChannel, shellIOChannel
}

func createCommandChannels() (chan websocket.Command, <-chan websocket.Command) {
	commandChannel := make(chan websocket.Command)
	return commandChannel, commandChannel
}
