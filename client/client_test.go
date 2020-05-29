package client_test

import (
    "errors"
	"net/url"
    "net/http"
    "testing"
    "time"
	"os"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
	"github.com/gorilla/websocket"

    "github.com/thecoderstudio/apollo-agent/client"
)

type ConnMock struct {
    mock.Mock
}

func (mocked ConnMock) Close() error {
    args := mocked.Called()
    return args.Error(0)
}

func (mocked ConnMock) ReadMessage() (messageType int, p []byte, err error) {
    args := mocked.Called()

    var message []byte
    if args.Get(1) != nil {
        message = args.Get(1).([]byte)
    }

    return args.Int(0), message, args.Error(2)
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

type ClientTestSuite struct {
    suite.Suite
}

func (suite *ClientTestSuite) TestListenSuccess() {
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)

    mockConn := new(ConnMock)
    mockConn.On("Close").Return(nil)
    mockConn.On(
        "WriteMessage",
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")).Return(nil)
    mockConn.On("ReadMessage").Return(0, nil, nil)

    mockDialer := new(DialerMock)
    mockDialer.On("Dial", u.String(), http.Header(nil)).Return(mockConn, nil, nil)

    wsClient := client.Create(mockDialer)
    go wsClient.Listen(u, &interrupt)
    close(interrupt)
}

func (suite *ClientTestSuite) TestCloseConnectionWriteError() {
    expectedError := errors.New("test")
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)

    mockConn := new(ConnMock)
    mockConn.On("Close").Return(nil)
    mockConn.On(
        "WriteMessage",
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")).Return(expectedError)
    mockConn.On("ReadMessage").Return(0, nil, nil)

    mockDialer := new(DialerMock)
    mockDialer.On("Dial", u.String(), http.Header(nil)).Return(mockConn, nil, nil)

    wsClient := client.Create(mockDialer)
    go wsClient.Listen(u, &interrupt)
    close(interrupt)
    time.Sleep(1 * time.Second)

    mockConn.AssertExpectations(suite.T())
}

func (suite *ClientTestSuite) TestConnectionError() {
    expectedError := errors.New("connection error")
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)
    defer close(interrupt)

    mockDialer := new(DialerMock)
    mockDialer.On("Dial", u.String(), http.Header(nil)).Return(nil, nil, expectedError)

    wsClient := client.Create(mockDialer)
    err := wsClient.Listen(u, &interrupt)

    mockDialer.AssertExpectations(suite.T())
    assert.EqualError(suite.T(), err, "connection error")
}

func (suite *ClientTestSuite) TestReadMessageError() {
    expectedError := errors.New("read error")
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	interrupt := make(chan os.Signal, 1)
    defer close(interrupt)

    mockConn := new(ConnMock)
    mockConn.On("Close").Return(nil)
    mockConn.On("ReadMessage").Return(0, nil, expectedError)

    mockDialer := new(DialerMock)
    mockDialer.On("Dial", u.String(), http.Header(nil)).Return(mockConn, nil, nil)

    wsClient := client.Create(mockDialer)
    wsClient.Listen(u, &interrupt)

    mockConn.AssertExpectations(suite.T())
}

func TestClientSuite(t *testing.T) {
    suite.Run(t, new(ClientTestSuite))
}
