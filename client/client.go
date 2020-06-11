package client

import (
    "log"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "time"

    "github.com/gorilla/websocket"
    "github.com/thecoderstudio/apollo-agent/oauth"
)

// WebsocketConn specifies the interface for client connection
type WebsocketConn interface {
    Close() error
    ReadMessage() (int, []byte, error)
    WriteMessage(int, []byte) error
}

// Dialer should be used to dial TCP connections returning a connection
// that fits the expected WebsocketConn interface.
type Dialer interface {
    Dial(string, http.Header) (WebsocketConn, *http.Response, error)
}

// DialWrapper wraps the websocket.DefaultDialer to transform returned types
// to interfaces the client expects.
type DialWrapper struct{}

// Dial wraps de default Dial to return an interface instead of a struct.
func (wrapper DialWrapper) Dial(urlString string, header http.Header) (WebsocketConn, *http.Response, error) {
    return websocket.DefaultDialer.Dial(urlString, header)
}

type client struct {
    dialer Dialer
}

// Listen connects to the given endpoint and handles incoming messages. It's interruptable
// by closing the interrupt channel.
func (client *client) Listen(endpointURL url.URL, accessToken oauth.AccessToken,
                             interrupt *chan os.Signal) (<-chan string, <-chan struct{}, <-chan error) {
    out := make(chan string)
    errs := make(chan error)
    done := make(chan struct{})

    go func() {
        defer close(out)
        defer close(errs)

        connection, err := client.createConnection(endpointURL, accessToken)
        if err != nil {
            log.Println("Connection error")
            errs <- err
            close(done)
            return
        }

        // Used by awaitMessages to communicate when done. Only when done can we close the channels
        // to prevent sending messages to closed channels.
        doneListening := make(chan struct{})

        go client.awaitMessages(&connection, &out, &errs, &done, &doneListening)
        err = client.handleEvents(&connection, &doneListening, interrupt)

        connection.Close()
        close(done)
        <-doneListening
    }()

    return out, done, errs
}

func (client *client) createConnection(endpointURL url.URL, accessToken oauth.AccessToken) (WebsocketConn, error) {
    urlString := endpointURL.String()
    log.Printf("connecting to %s", urlString)
    authorizationHeader := fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken)
    connection, _, err := client.dialer.Dial(urlString, http.Header{"Authorization": []string{authorizationHeader}})
    return connection, err
}

func (client *client) awaitMessages(connection *WebsocketConn, out *chan string, errs *chan error, done, doneListening *chan struct{}) {
    defer close(*doneListening)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    conn := *connection
    for {
        select {
        case <-*done:
            return
        case <-ticker.C:
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("read error:", err)
                *errs <- err
                return
            }
            *out <- string(message)
        }
    }
}

func (client *client) handleEvents(connection *WebsocketConn, doneListening *chan struct{},
                                   interrupt *chan os.Signal) error {
    for {
        select {
        case <-*doneListening:
            return nil
        case <-*interrupt:
            log.Println("interrupt")
            err := client.closeConnection(connection)
            return err
        }
    }
}

func (client *client) closeConnection(connection *WebsocketConn) error {
    conn := *connection
    err := conn.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    if err != nil {
        log.Println("Close err:", err)
        return err
    }

    return nil
}

// Create is the factory to create a properly instantiated client.
func Create(dialer Dialer) client {
    return client{dialer: dialer}
}
