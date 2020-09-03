package websocket_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

var u = url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}

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

type ClientTestSuite struct {
	suite.Suite
}

func (suite *ClientTestSuite) TestListenForShellIOSuccess() {
	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	mockConn := new(ConnMock)
	mockConn.MockClosed(nil)
	mockConn.On("ReadMessage").Return(0, []byte("{\"command\": \"new connection\"}"), nil).Once()
	mockConn.On("ReadMessage").Return(0, []byte("{\"message\": \"test message\"}"), nil)

	wsClient := createWsClient(mockConn)
	done := wsClient.Listen(u, oauth.AccessToken{}, in, &interrupt)
	command := <-wsClient.Commands()
	message := <-wsClient.Out()

	assert.Equal(suite.T(), "new connection", command.Command)
	assert.Equal(suite.T(), "test message", message.Message)

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())
}

func (suite *ClientTestSuite) TestCloseConnectionWriteError() {
	expectedError := errors.New("test")
	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	mockConn := new(ConnMock)
	mockConn.MockClosed(expectedError)
	mockConn.On("ReadMessage").Maybe().Return(0, nil, nil)

	wsClient := createWsClient(mockConn)
	done := wsClient.Listen(u, oauth.AccessToken{}, in, &interrupt)

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())
}

func (suite *ClientTestSuite) TestConnectionError() {
	expectedError := errors.New("connection error")
	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(interrupt)
	defer close(in)

	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(nil, nil, expectedError)

	wsClient := websocket.CreateClient(mockDialer)
	wsClient.Listen(u, oauth.AccessToken{}, in, &interrupt)
	err := <-wsClient.Errs()

	mockDialer.AssertExpectations(suite.T())
	assert.EqualError(suite.T(), err, "connection error")
}

func (suite *ClientTestSuite) TestReadMessageError() {
	expectedError := errors.New("read error")
	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(interrupt)
	defer close(in)

	mockConn := new(ConnMock)
	mockConn.On("Close").Return(nil)
	mockConn.On("ReadMessage").Return(0, nil, expectedError)

	wsClient := createWsClient(mockConn)
	done := wsClient.Listen(u, oauth.AccessToken{}, in, &interrupt)
	err := <-wsClient.Errs()

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())
	assert.EqualError(suite.T(), err, "read error")
}

func (suite *ClientTestSuite) TestWriteMessage() {
	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	testShellIO := websocket.ShellIO{
		ConnectionID: "test",
		Message:      "test",
	}
	jsonShellIO, _ := json.Marshal(testShellIO)

	mockConn := new(ConnMock)
	mockConn.MockClosed(nil)
	mockConn.On("ReadMessage").Maybe().Return(0, nil, nil)
	mockConn.On(
		"WriteMessage",
		gorilla.TextMessage,
		jsonShellIO).Return(nil)

	wsClient := createWsClient(mockConn)
	done := wsClient.Listen(u, oauth.AccessToken{}, in, &interrupt)

	in <- testShellIO

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())

}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func createWsClient(mockConn *ConnMock) websocket.Client {
	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(mockConn, nil, nil)

	wsClient := websocket.CreateClient(mockDialer)
	return wsClient
}
