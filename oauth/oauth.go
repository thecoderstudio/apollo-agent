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
    AgentID string
    ClientSecret string
}


// GetAccessToken requests and returns an AccessToken.
func (client *Client) GetAccessToken() (AccessToken, error) {
    var accessToken AccessToken
    url := url.URL{Scheme: "http", Host: client.Host, Path: "/oauth/token"}
    authHeader := client.getAuthHeader()

    req, _ := http.NewRequest("POST", url.String(), client.buildGrantTypeData())

    req.Header.Add("Authorization", authHeader)
    req.Header.Add("Content-Type", "application/json")

    resp, err := client.Client.Do(req)
    if err != nil {
        return accessToken, err
    }

    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    accessToken = AccessToken{}
    json.Unmarshal(body, &accessToken)

    return accessToken, nil
}

func (client *Client) getAuthHeader() string {
    creds := fmt.Sprintf("%s:%s", client.AgentID, client.ClientSecret)
    auth := fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(creds)))
    return auth
}

func (client *Client) buildGrantTypeData() *bytes.Buffer {
    values := map[string]string{"grant_type": "client_credentials"}
    jsonValue, _ := json.Marshal(values)
    return bytes.NewBuffer(jsonValue)
}

// GetContinuousAccessToken returns a channel over which an AccessToken will be sent.
// When the AccessToken is 2 minutes to expiration a new one will get requested and sent.
func (client *Client) GetContinuousAccessToken() (*chan AccessToken, *chan error) {
    channel := make(chan AccessToken)
    errs := make(chan error)
    go client.keepTokenAlive(&channel, &errs, 120)
    return &channel, &errs
}

func (client *Client) keepTokenAlive(accessTokenChannel *chan AccessToken, errs *chan error, offsetInSeconds int) {
    for {
        newAccessToken, err := client.GetAccessToken()
        if err != nil {
            *errs <- err
            return
        }
        *accessTokenChannel <- newAccessToken
        time.Sleep(time.Duration(newAccessToken.ExpiresIn - offsetInSeconds) * time.Second)
    }
}

// Create is a factory to create a properly instantiated Client
func Create(host string, agentID string, clientSecret string) Client {
    return Client{
        Host: host,
        Client: http.Client{},
        AgentID: agentID,
        ClientSecret: clientSecret,
    }
}
