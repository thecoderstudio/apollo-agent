package testutil

import (
	"net/http"
	"net/url"

	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"

	"github.com/thecoderstudio/apollo-agent/websocket"
)

type ConnMock struct {
	mock.Mock
}

func (mocked *ConnMock) Close() error {
	args := mocked.Called()
	return args.Error(0)
}

func (mocked *ConnMock) ReadMessage() (messageType int, p []byte, err error) {
	args := mocked.Called()

	var message []byte
	if args.Get(1) != nil {
		message = args.Get(1).([]byte)
	}

	return args.Int(0), message, args.Error(2)
}

func (mocked *ConnMock) WriteMessage(messageType int, data []byte) error {
	args := mocked.Called(messageType, data)
	return args.Error(0)
}

func (mocked *ConnMock) MockClosed(expectedError error) {
	mocked.On("Close").Return(nil)
	mocked.On(
		"WriteMessage",
		gorilla.CloseMessage,
		gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""),
	).Return(expectedError)
}

type DialerMock struct {
	mock.Mock
}

func (mocked *DialerMock) Dial(urlString string, header http.Header) (websocket.Connection, *http.Response, error) {
	args := mocked.Called(urlString, header)

	var connection websocket.Connection
	if args.Get(0) != nil {
		connection = args.Get(0).(websocket.Connection)
	}

	var response http.Response
	if args.Get(1) != nil {
		response = args.Get(1).(http.Response)
	}

	return connection, &response, args.Error(2)
}

var Url = url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}

func CreateWsClient(mockConn *ConnMock, authString string) websocket.Client {
	mockDialer := new(DialerMock)
	mockDialer.On("Dial", Url.String(), http.Header{"Authorization": []string{authString}}).Return(mockConn, nil, nil)

	wsClient := websocket.CreateClient(mockDialer)
	return wsClient
}
