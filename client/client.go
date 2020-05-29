package client

import (
	"log"
	"net/http"
	"net/url"
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
func (client *client) Listen(endpointURL url.URL, interrupt *chan os.Signal) (<-chan string, <-chan error, <-chan struct{}) {
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
        }

        // Ensure connection gets closed no matter what.
        defer connection.Close()

        go client.awaitMessages(connection, &out, &errs, &done)
        errs <- client.handleEvents(connection, &done, interrupt)
    }()

    return out, errs, done
}

func (client *client) createConnection(endpointURL url.URL) (WebsocketConn, error) {
	urlString := endpointURL.String()
	log.Printf("connecting to %s", urlString)
	connection, _, err := client.dialer.Dial(urlString, nil)
	return connection, err
}

func (client *client) awaitMessages(connection WebsocketConn, out *chan string, errs *chan error, done *chan struct{}) {
    defer close(*done)
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}
        *out <- string(message)
	}
}

func (client *client) handleEvents(connection WebsocketConn, done *chan struct{}, interrupt *chan os.Signal) error {
	for {
		select {
		case <-*done:
			return nil
		case <-*interrupt:
            log.Println("interrupt")
			err := client.closeConnection(connection)
			return err
		}
	}
}

func (client *client) closeConnection(connection WebsocketConn) error {
	err := connection.WriteMessage(
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
