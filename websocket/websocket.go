package websocket

import (
	"log"
	"net/url"
    "net/http"
	"time"
	"os"

	"github.com/gorilla/websocket"
)

type WebsocketConn interface {
    Close() error
    ReadMessage() (int, []byte, error)
    WriteMessage(int, []byte) error
}

type Dialer interface {
    Dial(string, http.Header) (WebsocketConn, *http.Response, error)
}


type DialWrapper struct {}

func (wrapper DialWrapper) Dial(urlString string, header http.Header) (WebsocketConn, *http.Response, error) {
    return websocket.DefaultDialer.Dial(urlString, header)
}

type websocketClient struct {
    dialer Dialer
}

func (client *websocketClient) Listen(endpointUrl url.URL, interrupt *chan os.Signal) {
    connection, err := client.createConnection(endpointUrl)
    if err != nil {
        log.Fatal("Connection error")
    }

    // Ensure connection gets closed no matter what.
    defer connection.Close()

    done := make(chan struct{})

    go client.awaitMessages(connection, &done)
    client.handleEvents(connection, &done, interrupt)
}


func (client *websocketClient) awaitMessages(connection WebsocketConn, done *chan struct{}) {
    defer close(*done)
    for {
        _, message, err := connection.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            return
        }
        log.Printf("recv: %s", message)
    }
}

func (client *websocketClient) handleEvents(connection WebsocketConn, done *chan struct{}, interrupt *chan os.Signal) {
    for {
		select {
		case <-*done:
			return
		case <-*interrupt:
	        client.closeConnection(connection, done)
            return
        }
	}
}

func (client *websocketClient) closeConnection(connection WebsocketConn, done *chan struct{}) error {
    log.Println("interrupt")

    err := connection.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    if err != nil {
        log.Println("Close err:", err)
        return err
    }
    select {
    case <-*done:
    case <-time.After(time.Second):
    }

    return nil
}

func (client *websocketClient) createConnection(endpointUrl url.URL) (WebsocketConn, error) {
    urlString := endpointUrl.String()
    log.Printf("connecting to %s", urlString)
	connection, _, err := client.dialer.Dial(urlString, nil)
    return connection, err
}

func CreateWebsocketClient(dialer Dialer) websocketClient {
    return websocketClient{dialer: dialer}
}
