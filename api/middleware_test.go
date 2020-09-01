package api_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/testutil"
)

func TestMiddleware(t *testing.T) {
	expectedError := errors.New("read error")
	interrupt := make(chan os.Signal)
	mockConn := new(testutil.ConnMock)
	mockConn.MockClosed(nil)
	mockConn.On("ReadMessage").Maybe().Return(0, nil, expectedError)

	serverMock := testutil.CreateServerMock(false)
	defer serverMock.Close()
	oauthClient := oauth.Create(strings.TrimPrefix(serverMock.URL, "http://"), "", "")

	wsClient := testutil.CreateWsClient(mockConn, "Bearer faketoken")
	middleware := api.Middleware{
		Host:            "localhost:8000",
		InterruptSignal: &interrupt,
		WebsocketClient: wsClient,
		OAuthClient:     oauthClient,
	}
	middleware.Start("bash")
}
