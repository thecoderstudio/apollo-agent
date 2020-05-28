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
    connection websocket.Conn
}

func (client *WebsocketClient) connectAndListen(endpointUrl url.URL, interrupt *chan os.Signal) {
    client.connection = *client.createConnection(endpointUrl)
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

func (client *WebsocketClient) closeConnection() {
    log.Println("interrupt")

    err := client.connection.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    if err != nil {
        log.Println("Close err:", err)
        return
    }
    select {
    case <-client.done:
    case <-time.After(time.Second):
    }
}

func createConnection(endpointUrl url.URL) *websocket.Conn {
    urlString := endpointUrl.String()
    log.Printf("connecting to %s", urlString)

	connection, _, err := websocket.DefaultDialer.Dial(urlString, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
    return connection
}
