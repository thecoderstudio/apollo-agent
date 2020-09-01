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
	errs := make(chan error)
	authProviderMock := new(mocks.AuthProvider)
	authProviderMock.On("GetContinuousAccessToken").Return(&accessTokenChan, &errs)

	accessToken := oauth.AccessToken{
		AccessToken: "",
		ExpiresIn:   3600,
		TokenType:   "",
	}
	done := make(<-chan struct{})
	out := make(<-chan websocket.ShellIO)
	commands := make(<-chan websocket.Command)

	shellInterfaceMock := new(mocks.ShellInterface)
	shellInterfaceMock.On("Listen", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(done)
	shellInterfaceMock.On("Out").Return(out)
	shellInterfaceMock.On("Commands").Return(commands)
	shellInterfaceMock.On("Errs").Return(errs)

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
}
