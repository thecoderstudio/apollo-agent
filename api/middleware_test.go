package api_test

import (
	"os"
	"testing"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/testutil"
)

func TestMiddleware(t *testing.T) {
	interrupt := make(chan os.Signal)
	mockConn := new(testutil.ConnMock)
	wsClient := testutil.CreateWsClient(mockConn)
	middleware := api.Middleware{
		Host:            "localhost:8000",
		InterruptSignal: &interrupt,
		WebsocketClient: wsClient,
	}
	middleware.Start("", "", "bash")
}
