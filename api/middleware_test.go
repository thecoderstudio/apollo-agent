package api_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/mocks"
	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestMiddleware(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
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
	_, readOnlyConnErrs := createErrChannels()

	shellInterfaceMock := createShellInterfaceMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	shellInterfaceMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(readOnlyDone)

	shellIO := websocket.ShellIO{ConnectionID: "1", Message: "echo 'test'"}
	command := websocket.Command{ConnectionID: "1", Command: "test"}
	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("ExecutePredefinedCommand", command)
	shellManagerMock.On("Execute", shellIO).Run(func(arguments mock.Arguments) {
		close(done)
	})

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

func TestReAuthentication(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	accessTokenChan := make(chan oauth.AccessToken)
	authErrs := make(chan error)
	authProviderMock := new(mocks.AuthProvider)
	authProviderMock.On("GetContinuousAccessToken").Return(&accessTokenChan, &authErrs)

	initialAccessToken := oauth.AccessToken{
		AccessToken: "initial",
		ExpiresIn:   3600,
		TokenType:   "",
	}
	secondAccessToken := oauth.AccessToken{
		AccessToken: "second",
		ExpiresIn:   3600,
		TokenType:   "",
	}
	done, readOnlyDone := createDoneChannels()
	_, readOnlyOut := createShellIOChannels()
	_, readOnlyCommands := createCommandChannels()
	_, readOnlyConnErrs := createErrChannels()

	shellInterfaceMock := createShellInterfaceMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	shellInterfaceMock.On("Listen", mock.Anything, initialAccessToken, mock.Anything, mock.Anything).Return(readOnlyDone).Once()
	shellInterfaceMock.On("Listen", mock.Anything, secondAccessToken, mock.Anything, mock.Anything).Return(readOnlyDone).Run(func(args mock.Arguments) {
		close(done)
	})

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")

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
	accessTokenChan <- initialAccessToken
	accessTokenChan <- secondAccessToken
	<-notify
	shellInterfaceMock.AssertExpectations(t)
}

func TestReconnect(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	accessTokenChan := make(chan oauth.AccessToken)
	authErrs := make(chan error)
	authProviderMock := new(mocks.AuthProvider)
	authProviderMock.On("GetContinuousAccessToken").Return(&accessTokenChan, &authErrs)

	accessToken := oauth.AccessToken{
		AccessToken: "initial",
		ExpiresIn:   3600,
		TokenType:   "",
	}
	_, readOnlyDone := createDoneChannels()
	_, readOnlyOut := createShellIOChannels()
	_, readOnlyCommands := createCommandChannels()
	connErrs, readOnlyConnErrs := createErrChannels()

	shellInterfaceMock := createShellInterfaceMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	shellInterfaceMock.On("Listen", mock.Anything, accessToken, mock.Anything, mock.Anything).Return(readOnlyDone)

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")

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
	connErrs <- errors.New("test")
	<-notify
}

func createShellInterfaceMock(
	done <-chan struct{},
	out <-chan websocket.ShellIO,
	commands <-chan websocket.Command,
	errs <-chan error,
) *mocks.ShellInterface {
	shellInterfaceMock := new(mocks.ShellInterface)
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

func createErrChannels() (chan error, <-chan error) {
	errChannel := make(chan error)
	return errChannel, errChannel
}
