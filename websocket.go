package main

import (
	"log"
	"net/url"
	"time"
	"os"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
    done chan struct{}
    connection *websocket.Conn
}

func (client *WebsocketClient) connectAndListen(endpointUrl url.URL, interrupt *chan os.Signal) {
    client.connection = client.connect(endpointUrl)
    defer client.connection.Close()

    done := make(chan struct{})

    go client.listen()

	for {
		select {
		case <-done:
			return
		case <-*interrupt:
	        client.closeConnection()
            return
        }
	}
}

func (client WebsocketClient) connect(endpointUrl url.URL) *websocket.Conn {
    urlString := endpointUrl.String()
    log.Printf("connecting to %s", urlString)

	c, _, err := websocket.DefaultDialer.Dial(urlString, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
    return c
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

func (client *WebsocketClient) closeConnection() {
    log.Println("interrupt")

    // Cleanly close the connection by sending a close message and then
    // waiting (with timeout) for the server to close the connection.
    err := client.connection.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    if err != nil {
        log.Println("write close:", err)
        return
    }
    select {
    case <-client.done:
    case <-time.After(time.Second):
    }
}
