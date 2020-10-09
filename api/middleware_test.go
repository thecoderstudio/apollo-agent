package api_test

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/mocks"
	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/shell"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestMiddleware(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	authProviderMock, accessTokenChan, _ := createAuthProviderMock()
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
	remoteTerminalMock.On("Listen", mock.Anything, mock.Anything, mock.Anything).Return(readOnlyDone)

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

	stopped := startMiddleware(&interruptSignal, remoteTerminalMock, shellManagerMock, authProviderMock)

	accessTokenChan <- accessToken
	commands <- command
	out <- shellIO
	<-stopped

	shellManagerMock.AssertExpectations(t)
}

func TestReAuthentication(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	authProviderMock, accessTokenChan, _ := createAuthProviderMock()
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
	remoteTerminalMock.On("Listen", mock.Anything, initialAccessToken, mock.Anything).Return(readOnlyDone).Once()
	remoteTerminalMock.On("Listen", mock.Anything, secondAccessToken, mock.Anything).Return(readOnlyDone).Run(func(args mock.Arguments) {
		close(done)
	})

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Out").Return(make(<-chan websocket.ShellIO))

	stopped := startMiddleware(&interruptSignal, remoteTerminalMock, shellManagerMock, authProviderMock)

	go func() {
		// Prevent blocking of the main thread because `Interrupt` goes unread due to
		// remoteTerminalMock being a mock
		for {
			<-remoteTerminalMock.Interrupt()
		}
	}()

	accessTokenChan <- initialAccessToken
	accessTokenChan <- secondAccessToken

	<-stopped
	remoteTerminalMock.AssertExpectations(t)
}

func TestAuthenticationFailure(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	authProviderMock, _, authErrs := createAuthProviderMock()

	_, readOnlyDone := createDoneChannels()
	_, readOnlyOut := createShellIOChannels()
	_, readOnlyCommands := createCommandChannels()
	_, readOnlyConnErrs := createErrChannels()

	remoteTerminalMock := createRemoteTerminalMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)

	shellManagerMock := new(mocks.ShellManager)

	stopped := startMiddleware(&interruptSignal, remoteTerminalMock, shellManagerMock, authProviderMock)
	authErrs <- errors.New("test")
	<-stopped
}

func TestReconnect(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	authProviderMock, accessTokenChan, _ := createAuthProviderMock()
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
	remoteTerminalMock.On("Listen", mock.Anything, accessToken, mock.Anything).Return(readOnlyDone).Once()
	remoteTerminalMock.On("Listen", mock.Anything, accessToken, mock.Anything).Return(readOnlyDone).Once().Run(
		func(args mock.Arguments) {
			close(done)
		},
	)

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Out").Return(make(<-chan websocket.ShellIO))

	stopped := startMiddleware(&interruptSignal, remoteTerminalMock, shellManagerMock, authProviderMock)

	accessTokenChan <- accessToken

	connErrs <- errors.New("test")
	<-stopped
}

func TestReconnectInterrupt(t *testing.T) {
	interruptSignal := make(chan os.Signal, 1)
	authProviderMock, accessTokenChan, _ := createAuthProviderMock()
	accessToken := oauth.AccessToken{
		AccessToken: "initial",
		ExpiresIn:   3600,
		TokenType:   "",
	}

	_, readOnlyDone := createDoneChannels()
	_, readOnlyOut := createShellIOChannels()
	_, readOnlyCommands := createCommandChannels()
	connErrs, readOnlyConnErrs := createErrChannels()

	remoteTerminalMock := createRemoteTerminalMock(readOnlyDone, readOnlyOut, readOnlyCommands, readOnlyConnErrs)
	remoteTerminalMock.On("Listen", mock.Anything, accessToken, mock.Anything).Return(readOnlyDone).Once()

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Out").Return(make(<-chan websocket.ShellIO))

	stopped := startMiddleware(&interruptSignal, remoteTerminalMock, shellManagerMock, authProviderMock)

	accessTokenChan <- accessToken

	connErrs <- errors.New("test")
	interruptSignal <- syscall.SIGINT
	<-stopped
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

func createAuthProviderMock() (*mocks.AuthProvider, chan oauth.AccessToken, chan error) {
	errs := make(chan error)
	accessTokenChan := make(chan oauth.AccessToken)
	authProviderMock := new(mocks.AuthProvider)
	authProviderMock.On("GetContinuousAccessToken").Return(&accessTokenChan, &errs)
	return authProviderMock, accessTokenChan, errs
}

func createRemoteTerminalMock(
	done <-chan struct{},
	out <-chan websocket.ShellIO,
	commands <-chan websocket.Command,
	errs <-chan error,
) *mocks.RemoteTerminal {
	interrupt := make(chan struct{})
	remoteTerminalMock := new(mocks.RemoteTerminal)
	remoteTerminalMock.On("Out").Return(out)
	remoteTerminalMock.On("Commands").Return(commands)
	remoteTerminalMock.On("Errs").Return(errs)
	remoteTerminalMock.On("Interrupt").Maybe().Return(interrupt)
	return remoteTerminalMock
}

func startMiddleware(
	interruptSignal *chan os.Signal,
	remoteTerminal websocket.RemoteTerminal,
	ptyManager shell.Manager,
	oauthClient oauth.AuthProvider,
) chan struct{} {
	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: interruptSignal,
		RemoteTerminal:  remoteTerminal,
		PTYManager:      ptyManager,
		OAuthClient:     oauthClient,
	}

	stopped := make(chan struct{})
	go func() {
		middleware.Start(1 * time.Second)
		close(stopped)
	}()
	return stopped
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
