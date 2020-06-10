package oauth_test

import (
    "encoding/json"
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

    assert.Equal(t, token.AccessToken, "faketoken")
    assert.Equal(t, token.ExpiresIn, 3600)
    assert.Equal(t, token.TokenType, "Bearer")
}

func createServerMock() *httptest.Server {
    handler := http.NewServeMux()
    handler.HandleFunc("/oauth/token", authTokenMock)
    return httptest.NewServer(handler)
}

func authTokenMock(writer http.ResponseWriter, request *http.Request) {
    accessToken := &oauth.AccessToken{
        AccessToken:    "faketoken",
        ExpiresIn:      3600,
        TokenType:      "Bearer",
    }

    accessTokenJson, _ := json.Marshal(accessToken)
    _, _ = writer.Write(accessTokenJson)
}
