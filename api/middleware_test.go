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
	commands := make(<-chan websocket.Command)

	shellInterfaceMock := new(mocks.ShellInterface)
	shellInterfaceMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(readOnlyDone)
	shellInterfaceMock.On("Out").Return(readOnlyOut)
	shellInterfaceMock.On("Commands").Return(commands)
	shellInterfaceMock.On("Errs").Return(shellErrs)

	shellManagerMock := new(mocks.ShellManager)
	shellManagerMock.On("Close")
	shellManagerMock.On("Execute", mock.Anything)

	interruptSignal := make(chan os.Signal, 1)
	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		ShellInterface:  shellInterfaceMock,
		PTYManager:      shellManagerMock,
		OAuthClient:     authProviderMock,
	}

	go func() {
		middleware.Start()
	}()
	accessTokenChan <- accessToken
	out <- websocket.ShellIO{ConnectionID: "1", Message: "echo 'test'"}
	close(done)
}

func convertReadOnlyDone(channel chan struct{}) <-chan struct{} {
	return channel
}

func convertReadOnlyOut(channel chan websocket.ShellIO) <-chan websocket.ShellIO {
	return channel
}
