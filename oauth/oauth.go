package oauth

import (
    "bytes"
    b64 "encoding/base64"
    "encoding/json"
	"fmt"
	"io/ioutil"
    "net/http"
    "net/url"
)

type OAuthClient struct {
    Host string
    Client http.Client
}

func (client *OAuthClient) GetAccessToken() string {
    url := url.URL{Scheme: "http", Host: client.Host, Path: "/oauth/token"}
    values := map[string]string{"grant_type": "client_credentials"}
    jsonValue, _ := json.Marshal(values)

    creds := "73d711e0-923d-42a7-9857-5f3d67d88370:8f5712b5efc5fd711abb3d16925e25a41561e92a041ab4956083d2cfdb5f442e"

    auth := fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(creds)))

    req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonValue))
    if err != nil {
        panic(err)
    }

    req.Header.Add("authorization", auth)
    req.Header.Add("Content-Type", "application/json")

    resp, err := client.Client.Do(req)
    if err != nil {
        panic(err)
    }

	defer resp.Body.Close()

    fmt.Println("response Status:", resp.Status)
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println("response Body:", string(body))
    return string(body)
}

func Create(host string) OAuthClient {
    return OAuthClient{Host: host, Client: http.Client{}}
}
