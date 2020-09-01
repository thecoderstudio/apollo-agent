package api_test

import (
	"os"
	"testing"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/mocks"
)

func TestMiddleware(t *testing.T) {
	shellInterfaceMock := new(mocks.ShellInterface)
	authProviderMock := new(mocks.AuthProvider)
	interruptSignal := make(chan os.Signal, 1)
	middleware := api.Middleware{
		Host:            "localhost:8080",
		InterruptSignal: &interruptSignal,
		ShellInterface:  shellInterfaceMock,
		OAuthClient:     authProviderMock,
	}

	middleware.Start("bash")
}
