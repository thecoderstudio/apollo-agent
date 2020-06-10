package oauth

import (
    "bytes"
    "encoding/json"
	"fmt"
	"io/ioutil"
    "net/http"
    "net/url"
)

type OAuthClient struct {
    Host string
}

func (client *OAuthClient) GetAccessToken() {
    url := url.URL{Scheme: "http", Host: client.Host, Path: "/oauth/token"}
    values := map[string]string{"grant_type": "client_credentials"}
    jsonValue, _ := json.Marshal(values)

    resp, err := http.Post(url.String(), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
        panic(err)
    }

	defer resp.Body.Close()

    fmt.Println("response Status:", resp.Status)
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println("response Body:", string(body))
}
