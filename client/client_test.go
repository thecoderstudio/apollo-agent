package client_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/thecoderstudio/apollo-agent/client"
	"github.com/thecoderstudio/apollo-agent/oauth"
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
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    ).Return(expectedError)
}

type DialerMock struct {
	mock.Mock
}

func (mocked *DialerMock) Dial(urlString string, header http.Header) (client.WebsocketConn, *http.Response, error) {
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
	interrupt := make(chan struct{})
	in := make(chan client.Message)
	defer close(in)

	mockConn := new(ConnMock)
    mockConn.MockClosed(nil)
	mockConn.On("ReadMessage").Return(0, []byte("{\"message\": \"test message\"}"), nil)

	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(mockConn, nil, nil)

	wsClient := client.Create(mockDialer)
	out, done, _ := wsClient.Listen(u, oauth.AccessToken{}, &in, &interrupt)
	message := <-out

	assert.Equal(suite.T(), "test message", message.Message)

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
}

func (suite *ClientTestSuite) TestCloseConnectionWriteError() {
	expectedError := errors.New("test")
	interrupt := make(chan struct{})
	in := make(chan client.Message)
	defer close(in)

	mockConn := new(ConnMock)
    mockConn.MockClosed(expectedError)
	mockConn.On("ReadMessage").Return(0, nil, nil)

	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(mockConn, nil, nil)

	wsClient := client.Create(mockDialer)
	out, done, _ := wsClient.Listen(u, oauth.AccessToken{}, &in, &interrupt)
	<-out

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())
}

func (suite *ClientTestSuite) TestConnectionError() {
	expectedError := errors.New("connection error")
	interrupt := make(chan struct{})
	in := make(chan client.Message)
	defer close(interrupt)
	defer close(in)

	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(nil, nil, expectedError)

	wsClient := client.Create(mockDialer)
	_, _, errs := wsClient.Listen(u, oauth.AccessToken{}, &in, &interrupt)
	err := <-errs

	mockDialer.AssertExpectations(suite.T())
	assert.EqualError(suite.T(), err, "connection error")
}

func (suite *ClientTestSuite) TestReadMessageError() {
	expectedError := errors.New("read error")
	interrupt := make(chan struct{})
	in := make(chan client.Message)
	defer close(interrupt)
	defer close(in)

	mockConn := new(ConnMock)
	mockConn.On("Close").Return(nil)
	mockConn.On("ReadMessage").Return(0, nil, expectedError)

	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(mockConn, nil, nil)

	wsClient := client.Create(mockDialer)
	_, done, errs := wsClient.Listen(u, oauth.AccessToken{}, &in, &interrupt)
	err := <-errs

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())
	assert.EqualError(suite.T(), err, "read error")
}

func (suite *ClientTestSuite) TestWriteMessage() {
	interrupt := make(chan struct{})
	in := make(chan client.Message)
	defer close(in)

	testMessage := client.Message{
		ConnectionID: "test",
		Message:      "test",
	}
	jsonMessage, _ := json.Marshal(testMessage)

	mockConn := new(ConnMock)
    mockConn.MockClosed(nil)
	mockConn.On(
		"WriteMessage",
		websocket.TextMessage,
		jsonMessage).Return(nil)

	mockDialer := new(DialerMock)
	mockDialer.On("Dial", u.String(), http.Header{"Authorization": []string{" "}}).Return(mockConn, nil, nil)

	wsClient := client.Create(mockDialer)
	_, done, _ := wsClient.Listen(u, oauth.AccessToken{}, &in, &interrupt)

	in <- testMessage

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())

}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
