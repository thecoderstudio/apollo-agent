package oauth

import (
    "bytes"
    b64 "encoding/base64"
    "encoding/json"
	"fmt"
    "time"
	"io/ioutil"
    "net/http"
    "net/url"
)

// AccessToken is meant to be sent in an Authorization header to the Apollo API.
type AccessToken struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int `json:"expires_in"`
    TokenType   string `json:"token_type"`
}

// Client is responsible for getting an AccessToken through the OAuth protocol.
type Client struct {
    Host string
    Client http.Client
}


// GetAccessToken requests and returns an AccessToken.
func (client *Client) GetAccessToken() AccessToken {
    url := url.URL{Scheme: "http", Host: client.Host, Path: "/oauth/token"}
    values := map[string]string{"grant_type": "client_credentials"}
    jsonValue, _ := json.Marshal(values)

    // Fake creds
    creds := "73d711e0-923d-42a7-9857-5f3d67d88370:8f5712b5efc5fd711abb3d16925e25a41561e92a041ab4956083d2cfdb5f442e"

    auth := fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(creds)))

    req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonValue))
    if err != nil {
        panic(err)
    }

    req.Header.Add("Authorization", auth)
    req.Header.Add("Content-Type", "application/json")

    resp, err := client.Client.Do(req)
    if err != nil {
        panic(err)
    }

    defer resp.Body.Close()

    fmt.Println("response Status:", resp.Status)
    body, _ := ioutil.ReadAll(resp.Body)
    accessToken := AccessToken{}
    json.Unmarshal(body, &accessToken)
    return accessToken
}

// GetContinuousAccessToken returns a channel over which an AccessToken will be sent.
// When the AccessToken is 2 minutes to expiration a new one will get requested and sent.
func (client *Client) GetContinuousAccessToken() *chan AccessToken {
    channel := make(chan AccessToken)
    go client.keepTokenAlive(&channel, 120)
    return &channel
}

func (client *Client) keepTokenAlive(accessTokenChannel *chan AccessToken, offsetInSeconds int) {
    for {
        newAccessToken := client.GetAccessToken()
        *accessTokenChannel <- newAccessToken
        time.Sleep(time.Duration(newAccessToken.ExpiresIn - offsetInSeconds) * time.Second)
    }
}

// Create is a factory to create a properly instantiated Client
func Create(host string) Client {
    return Client{Host: host, Client: http.Client{}}
}
