package oauth_test

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/thecoderstudio/apollo-agent/oauth"
)

func TestGetAccessToken(t *testing.T) {
    serverMock := createServerMock()
    defer serverMock.Close()
    oauthClient := oauth.Create(strings.TrimPrefix(serverMock.URL, "http://"))
    token := oauthClient.GetAccessToken()

    assert.Equal(t, token, "Test")
}

func createServerMock() *httptest.Server {
    handler := http.NewServeMux()
    handler.HandleFunc("/oauth/token", authTokenMock)
    return httptest.NewServer(handler)
}

func authTokenMock(writer http.ResponseWriter, request *http.Request) {
    _, _ = writer.Write([]byte("Test"))
}
