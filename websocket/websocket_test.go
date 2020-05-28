package websocket

import (
    "errors"
	"net/url"
    "net/http"
    "testing"
	"os"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
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
    args := mocked.Called()
    return args.Error(0)
}

type DialerMock struct {
    mock.Mock
}

func (mocked DialerMock) Dial(urlString string, header http.Header)(WebsocketConn, *http.Response, error) {
    return ConnMock{}, nil, nil
}


func TestListen(t *testing.T) {
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)
    client := CreateWebsocketClient(DialerMock{})
    go client.Listen(u, &interrupt)
    close(interrupt)
}

func TestCloseConnectionWriteError(t *testing.T) {
    expectedError := errors.New("test")

    mockObj := new(ConnMock)
    mockObj.On("WriteMessage").Return(expectedError)
    done := make(chan struct{})

    client := CreateWebsocketClient(DialerMock{})
    err := client.closeConnection(mockObj, &done)

    assert.Equal(t, err, expectedError)
}
