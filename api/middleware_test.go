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
	readOnlyDone := convertReadOnly(done)
	out := make(<-chan websocket.ShellIO)
	shellErrs := make(<-chan error)
	commands := make(<-chan websocket.Command)

	shellInterfaceMock := new(mocks.ShellInterface)
	shellInterfaceMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(readOnlyDone)
	shellInterfaceMock.On("Out").Return(out)
	shellInterfaceMock.On("Commands").Return(commands)
	shellInterfaceMock.On("Errs").Return(shellErrs)

	interruptSignal := make(chan os.Signal, 1)
	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		ShellInterface:  shellInterfaceMock,
		OAuthClient:     authProviderMock,
	}

	go func() {
		middleware.Start("bash")
	}()
	accessTokenChan <- accessToken
	close(done)
}

func convertReadOnly(channel chan struct{}) <-chan struct{} {
	return channel
}
