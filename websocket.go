package main

import (
    "errors"
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

type WebsocketClient struct {
    dialer Dialer
    done chan struct{}
    connection WebsocketConn
}

func (client *WebsocketClient) connectAndListen(endpointUrl url.URL, interrupt *chan os.Signal) {
    connection, err := client.createConnection(endpointUrl)
    if err != nil {
        log.Fatal("Connection error")
    }
    client.connection = connection

    defer client.connection.Close()

    client.done = make(chan struct{})

    go client.listen()

	for {
		select {
		case <-client.done:
			return
		case <-*interrupt:
	        client.closeConnection()
            return
        }
	}
}


func (client *WebsocketClient) listen() {
    defer close(client.done)
    for {
        _, message, err := client.connection.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            return
        }
        log.Printf("recv: %s", message)
    }
}

func (client *WebsocketClient) closeConnection() error {
    log.Println("interrupt")
    if client.connection == nil {
        return errors.New("No open connection")
    }

    err := client.connection.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    if err != nil {
        log.Println("Close err:", err)
        return err
    }
    select {
    case <-client.done:
    case <-time.After(time.Second):
    }

    return nil
}

func (client *WebsocketClient) createConnection(endpointUrl url.URL) (WebsocketConn, error) {
    urlString := endpointUrl.String()
    log.Printf("connecting to %s", urlString)
	connection, _, err := client.dialer.Dial(urlString, nil)
    return connection, err
}
