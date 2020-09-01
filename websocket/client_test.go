package websocket_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/testutil"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

type ClientTestSuite struct {
	suite.Suite
}

func (suite *ClientTestSuite) TestListenForShellIOSuccess() {
	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	mockConn := new(testutil.ConnMock)
	mockConn.MockClosed(nil)
	mockConn.On("ReadMessage").Return(0, []byte("{\"command\": \"new connection\"}"), nil).Once()
	mockConn.On("ReadMessage").Return(0, []byte("{\"message\": \"test message\"}"), nil)

	wsClient := testutil.CreateWsClient(mockConn, " ")
	done := wsClient.Listen(testutil.Url, oauth.AccessToken{}, &in, &interrupt)
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

	mockConn := new(testutil.ConnMock)
	mockConn.MockClosed(expectedError)
	mockConn.On("ReadMessage").Maybe().Return(0, nil, nil)

	wsClient := testutil.CreateWsClient(mockConn, " ")
	done := wsClient.Listen(testutil.Url, oauth.AccessToken{}, &in, &interrupt)

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

	mockDialer := new(testutil.DialerMock)
	mockDialer.On("Dial", testutil.Url.String(), http.Header{"Authorization": []string{" "}}).Return(nil, nil, expectedError)

	wsClient := websocket.CreateClient(mockDialer)
	wsClient.Listen(testutil.Url, oauth.AccessToken{}, &in, &interrupt)
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

	mockConn := new(testutil.ConnMock)
	mockConn.On("Close").Return(nil)
	mockConn.On("ReadMessage").Return(0, nil, expectedError)

	wsClient := testutil.CreateWsClient(mockConn, " ")
	done := wsClient.Listen(testutil.Url, oauth.AccessToken{}, &in, &interrupt)
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

	mockConn := new(testutil.ConnMock)
	mockConn.MockClosed(nil)
	mockConn.On("ReadMessage").Return(0, nil, nil)
	mockConn.On(
		"WriteMessage",
		gorilla.TextMessage,
		jsonShellIO).Return(nil)

	wsClient := testutil.CreateWsClient(mockConn, " ")
	done := wsClient.Listen(testutil.Url, oauth.AccessToken{}, &in, &interrupt)

	in <- testShellIO

	close(interrupt)

	assert.NotNil(suite.T(), <-done)
	mockConn.AssertExpectations(suite.T())

}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
