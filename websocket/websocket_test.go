package websocket

import (
    "errors"
	"net/url"
    "net/http"
    "testing"
	"os"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
	"github.com/gorilla/websocket"
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

func (mocked DialerMock) Dial(urlString string, header http.Header)(WebsocketConn, *http.Response, error) {
    args := mocked.Called(urlString, header)

    var connection WebsocketConn
    if args.Get(0) != nil {
        connection = args.Get(0).(WebsocketConn)
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

    client := CreateWebsocketClient(mockDialer)
    go client.Listen(u, &interrupt)
    close(interrupt)
}

func TestCloseConnectionWriteError(t *testing.T) {
    expectedError := errors.New("test")

    mockObj := new(ConnMock)
    mockObj.On(
        "WriteMessage",
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")).Return(expectedError)

    done := make(chan struct{})

    client := CreateWebsocketClient(DialerMock{})
    err := client.closeConnection(mockObj, &done)

    assert.Equal(t, err, expectedError)
}
