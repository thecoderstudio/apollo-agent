package client_test

import (
    "errors"
	"net/url"
    "net/http"
    "testing"
    "time"
	"os"

    "github.com/stretchr/testify/mock"
	"github.com/gorilla/websocket"

    "github.com/thecoderstudio/apollo-agent/websocket"
)

type ConnMock struct {
    mock.Mock
}

func (mocked ConnMock) Close() error {
    args := mocked.Called()
    return args.Error(0)
}

func (mocked ConnMock) ReadMessage() (messageType int, p []byte, err error) {
    return 0, []byte("test message"), nil
}

func (mocked ConnMock) WriteMessage(messageType int, data []byte) error {
    args := mocked.Called(messageType, data)
    return args.Error(0)
}

type DialerMock struct {
    mock.Mock
}

func (mocked DialerMock) Dial(urlString string, header http.Header)(client.WebsocketConn, *http.Response, error) {
    args := mocked.Called(urlString, header)

    var connection client.WebsocketConn
    if args.Get(0) != nil {
        connection = args.Get(0).(client.WebsocketConn)
    }

    var response http.Response
    if args.Get(1) != nil {
        response = args.Get(1).(http.Response)
    }

    return connection, &response, args.Error(2)
}


func TestListen(t *testing.T) {
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)

    mockConn := new(ConnMock)
    mockConn.On("Close").Return(nil)
    mockConn.On(
        "WriteMessage",
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")).Return(nil)

    mockDialer := new(DialerMock)
    mockDialer.On("Dial", u.String(), http.Header(nil)).Return(mockConn, nil, nil)

    wsClient := client.Create(mockDialer)
    go wsClient.Listen(u, &interrupt)
    close(interrupt)
}

func TestCloseConnectionWriteError(t *testing.T) {
    expectedError := errors.New("test")
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)

    mockConn := new(ConnMock)
    mockConn.On("Close").Return(nil)
    mockConn.On(
        "WriteMessage",
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")).Return(expectedError)

    mockDialer := new(DialerMock)
    mockDialer.On("Dial", u.String(), http.Header(nil)).Return(mockConn, nil, nil)

    wsClient := client.Create(mockDialer)
    go wsClient.Listen(u, &interrupt)
    close(interrupt)
    time.Sleep(5 * time.Second)

    mockConn.AssertExpectations(t)
}
