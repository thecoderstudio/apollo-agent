package oauth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/testutil"
)

func TestGetAccessToken(t *testing.T) {
	serverMock := testutil.CreateServerMock(false)
	defer serverMock.Close()
	oauthClient := oauth.Create(strings.TrimPrefix(serverMock.URL, "http://"), "", "")
	token, err := oauthClient.GetAccessToken()

	assert.NoError(t, err)
	assert.Equal(t, token.AccessToken, "faketoken")
	assert.Equal(t, token.ExpiresIn, 121)
	assert.Equal(t, token.TokenType, "Bearer")
}

func TestGetAccessTokenMalformedURL(t *testing.T) {
	serverMock := testutil.CreateServerMock(false)
	defer serverMock.Close()
	oauthClient := oauth.Create("fakeurl", "", "")
	_, err := oauthClient.GetAccessToken()
	assert.Error(t, err, "lookup fakeurl: no such host")
}

func TestGetContinuousAccessToken(t *testing.T) {
	serverMock := testutil.CreateServerMock(false)
	defer serverMock.Close()
	oauthClient := oauth.Create(strings.TrimPrefix(serverMock.URL, "http://"), "", "")

	start := time.Now()
	tokenChannel, _ := oauthClient.GetContinuousAccessToken()
	firstAccessToken := <-*tokenChannel
	secondAccessToken := <-*tokenChannel
	elapsed := time.Since(start)

	assert.NotNil(t, firstAccessToken)
	assert.NotNil(t, secondAccessToken)
	assert.LessOrEqual(t, elapsed.Seconds(), float64(2))
}

func TestGetContinuousAccessTokenMalformedURL(t *testing.T) {
	serverMock := testutil.CreateServerMock(false)
	defer serverMock.Close()
	oauthClient := oauth.Create("fakeurl", "", "")
	_, errs := oauthClient.GetContinuousAccessToken()
	err := <-*errs
	assert.Error(t, err, "lookup fakeurl: no such host")
}
