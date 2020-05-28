package main

import (
	"net/url"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type ConnMock struct {
    mock.Mock
}

func (mocked ConnMock) Close() error {
    return nil
}

func (mocked ConnMock) ReadMessage() (messageType int, p []byte, err error) {
    return 0, nil, nil
}

func (mocked ConnMock) WriteMessage(messageType int, data []byte) error {
    return nil
}

type DialerMock struct {
    mock.Mock
}

func (mocked DialerMock) Dial(urlString string, header http.Header)(WebsocketConn, *http.Response, error) {
    return ConnMock{}, nil, nil
}

func TestCloseConnectError(t *testing.T) {
    client := WebsocketClient{dialer: DialerMock{}}
    assert.EqualError(t, client.closeConnection(), "No open connection")
}

func TestCloseConnectSuccess(t *testing.T) {
    u := url.URL{Scheme: "ws", Host: "localhost:1970", Path: "/ws"}
    client := WebsocketClient{dialer: DialerMock{}}
    client.connection, _ = client.createConnection(u)

    assert.Nil(t, client.closeConnection())
}

func TestCreateConnection(t *testing.T) {
    u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
    client := WebsocketClient{dialer: DialerMock{}}

    connection, err := client.createConnection(u)

    if err != nil {
        t.Errorf("createConnection erred: %s", err)
    } else {
        connection.Close()
    }
}
