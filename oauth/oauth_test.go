package oauth_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"

    "github.com/thecoderstudio/apollo-agent/oauth"
)

func TestGetAccessToken(t *testing.T) {
    serverMock := createServerMock()
    defer serverMock.Close()
    oauthClient := oauth.Create(strings.TrimPrefix(serverMock.URL, "http://"), "", "")
    token := oauthClient.GetAccessToken()

    assert.Equal(t, token.AccessToken, "faketoken")
    assert.Equal(t, token.ExpiresIn, 121)
    assert.Equal(t, token.TokenType, "Bearer")
}

func TestGetContinuousAccessToken(t *testing.T) {
    serverMock := createServerMock()
    defer serverMock.Close()
    oauthClient := oauth.Create(strings.TrimPrefix(serverMock.URL, "http://"), "", "")

    start := time.Now()
    tokenChannel := oauthClient.GetContinuousAccessToken()
    firstAccessToken := <-*tokenChannel
    secondAccessToken := <-*tokenChannel
    elapsed := time.Since(start)

    assert.NotNil(t, firstAccessToken)
    assert.NotNil(t, secondAccessToken)
    assert.LessOrEqual(t, elapsed.Seconds(), float64(2))
}

func createServerMock() *httptest.Server {
    handler := http.NewServeMux()
    handler.HandleFunc("/oauth/token", authTokenMock)
    return httptest.NewServer(handler)
}

func authTokenMock(writer http.ResponseWriter, request *http.Request) {
    if request.Header.Get("Authorization") == "" {
        errorBody := map[string]string{"detail": "Invalid Authorization header"}
        errorBodyJSON, _ := json.Marshal(errorBody)
        writer.WriteHeader(http.StatusBadRequest)
        writer.Write(errorBodyJSON)
    }

    accessToken := &oauth.AccessToken{
        AccessToken:    "faketoken",
        ExpiresIn:      121,
        TokenType:      "Bearer",
    }

    accessTokenJSON, _ := json.Marshal(accessToken)
    _, _ = writer.Write(accessTokenJSON)
}
