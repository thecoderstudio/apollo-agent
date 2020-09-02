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
	done := make(chan struct{})
	readOnlyDone := convertReadOnlyDone(done)
	out := make(chan websocket.ShellIO)
	readOnlyOut := convertReadOnlyOut(out)
	shellErrs := make(<-chan error)
	commands := make(chan websocket.Command)
	readOnlyCommands := convertReadOnlyCommands(commands)

	shellInterfaceMock := new(mocks.ShellInterface)
	shellInterfaceMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(readOnlyDone)
	shellInterfaceMock.On("Out").Return(readOnlyOut)
	shellInterfaceMock.On("Commands").Return(readOnlyCommands)
	shellInterfaceMock.On("Errs").Return(shellErrs)

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

func convertReadOnlyDone(channel chan struct{}) <-chan struct{} {
	return channel
}

func convertReadOnlyOut(channel chan websocket.ShellIO) <-chan websocket.ShellIO {
	return channel
}

func convertReadOnlyCommands(channel chan websocket.Command) <-chan websocket.Command {
	return channel
}
