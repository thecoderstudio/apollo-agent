package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/thecoderstudio/apollo-agent/oauth"
)

func CreateServerMock(throwErr bool) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/oauth/token", func(writer http.ResponseWriter, request *http.Request) {
		AuthTokenMock(writer, request, throwErr)
	})
	return httptest.NewServer(handler)
}

func AuthTokenMock(writer http.ResponseWriter, request *http.Request, throwErr bool) {
	if throwErr {
		_, _ = writer.Write([]byte("something wrong"))
		return
	}

	if request.Header.Get("Authorization") == "" {
		errorBody := map[string]string{"detail": "Invalid Authorization header"}
		errorBodyJSON, _ := json.Marshal(errorBody)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(errorBodyJSON)
	}

	accessToken := &oauth.AccessToken{
		AccessToken: "faketoken",
		ExpiresIn:   121,
		TokenType:   "Bearer",
	}

	accessTokenJSON, _ := json.Marshal(accessToken)
	_, _ = writer.Write(accessTokenJSON)
}
