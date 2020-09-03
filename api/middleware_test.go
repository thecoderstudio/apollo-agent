package api_test

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
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

	remoteTerminalMock := createRemoteTerminalMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	remoteTerminalMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(readOnlyDone)

	shellIO := websocket.ShellIO{ConnectionID: "1", Message: "echo 'test'"}
	command := websocket.Command{ConnectionID: "1", Command: "test"}
	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Out").Return(make(<-chan websocket.ShellIO))
	shellManagerMock.On("ExecutePredefinedCommand", command)
	shellManagerMock.On("Execute", shellIO).Return(nil).Run(func(arguments mock.Arguments) {
		interruptSignal <- syscall.SIGINT
		close(done)
	})

	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		RemoteTerminal:  remoteTerminalMock,
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

	remoteTerminalMock := createRemoteTerminalMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	remoteTerminalMock.On("Listen", mock.Anything, initialAccessToken, mock.Anything, mock.Anything).Return(readOnlyDone).Once()
	remoteTerminalMock.On("Listen", mock.Anything, secondAccessToken, mock.Anything, mock.Anything).Return(readOnlyDone).Run(func(args mock.Arguments) {
		close(done)
	})

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Out").Return(make(<-chan websocket.ShellIO))

	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		RemoteTerminal:  remoteTerminalMock,
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
	remoteTerminalMock.AssertExpectations(t)
}

func TestAuthenticationFailure(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	accessTokenChan := make(chan oauth.AccessToken)
	authErrs := make(chan error)
	authProviderMock := new(mocks.AuthProvider)
	authProviderMock.On("GetContinuousAccessToken").Return(&accessTokenChan, &authErrs)

	_, readOnlyDone := createDoneChannels()
	_, readOnlyOut := createShellIOChannels()
	_, readOnlyCommands := createCommandChannels()
	_, readOnlyConnErrs := createErrChannels()

	remoteTerminalMock := createRemoteTerminalMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)

	shellManagerMock := new(mocks.ShellManager)

	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		RemoteTerminal:  remoteTerminalMock,
		PTYManager:      shellManagerMock,
		OAuthClient:     authProviderMock,
	}

	notify := make(chan struct{})
	go func() {
		middleware.Start()
		close(notify)
	}()
	authErrs <- errors.New("test")
	<-notify
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
	done, readOnlyDone := createDoneChannels()
	_, readOnlyOut := createShellIOChannels()
	_, readOnlyCommands := createCommandChannels()
	connErrs, readOnlyConnErrs := createErrChannels()

	remoteTerminalMock := createRemoteTerminalMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	remoteTerminalMock.On("Listen", mock.Anything, accessToken, mock.Anything, mock.Anything).Return(readOnlyDone)

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Out").Return(make(<-chan websocket.ShellIO))

	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		RemoteTerminal:  remoteTerminalMock,
		PTYManager:      shellManagerMock,
		OAuthClient:     authProviderMock,
		Connected:       make(chan bool),
	}

	notify := make(chan struct{})
	go func() {
		middleware.Start()
		close(notify)
	}()
	accessTokenChan <- accessToken

	go func() {
		for {
			connected := <-middleware.Connected
			fmt.Println(connected)
			if connected {
				close(done)
			}
		}
	}()
	connErrs <- errors.New("test")
	<-notify
}

func TestCreateMiddleware(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	middleware, err := api.CreateMiddleware("", "", "", "/bin/bash", &interruptSignal)
	assert.NoError(t, err)
	assert.NotNil(t, middleware)
}

func TestCreateMiddlewareInvalidShell(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	_, err := api.CreateMiddleware("", "", "", "", &interruptSignal)
	assert.Error(t, err)
}

func createRemoteTerminalMock(
	done <-chan struct{},
	out <-chan websocket.ShellIO,
	commands <-chan websocket.Command,
	errs <-chan error,
) *mocks.RemoteTerminal {
	remoteTerminalMock := new(mocks.RemoteTerminal)
	remoteTerminalMock.On("Out").Return(out)
	remoteTerminalMock.On("Commands").Return(commands)
	remoteTerminalMock.On("Errs").Return(errs)
	return remoteTerminalMock
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
