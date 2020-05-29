package client

import (
	"log"
	"net/http"
	"net/url"
    "time"
	"os"

	"github.com/gorilla/websocket"
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
func (client *client) Listen(endpointURL url.URL, interrupt *chan os.Signal) (<-chan string, <-chan struct{}, <-chan error) {
    out := make(chan string)
    errs := make(chan error)
    done := make(chan struct{})

    go func(){
        defer close(out)
        defer close(errs)

        connection, err := client.createConnection(endpointURL)
        if err != nil {
            log.Println("Connection error")
            errs <- err
            return
        }

        awaitDone := make(chan struct{})

        go client.awaitMessages(&connection, &out, &errs, &done, &awaitDone)
        err = client.handleEvents(&connection, &awaitDone, interrupt)

        connection.Close()
        close(done)
        <-awaitDone
    }()

    return out, done, errs
}

func (client *client) createConnection(endpointURL url.URL) (WebsocketConn, error) {
	urlString := endpointURL.String()
	log.Printf("connecting to %s", urlString)
	connection, _, err := client.dialer.Dial(urlString, nil)
	return connection, err
}

func (client *client) awaitMessages(connection *WebsocketConn, out *chan string, errs *chan error, done, awaitDone *chan struct{}) {
    defer close(*awaitDone)
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

func (client *client) handleEvents(connection *WebsocketConn, awaitDone *chan struct{}, interrupt *chan os.Signal) error {
	for {
		select {
		case <-*awaitDone:
			return nil
		case <-*interrupt:
            log.Println("interrupt")
			err := client.closeConnection(connection, awaitDone)
			return err
		}
	}
}

func (client *client) closeConnection(connection *WebsocketConn, awaitDone *chan struct{}) error {
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
