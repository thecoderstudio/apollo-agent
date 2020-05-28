package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:1970", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

    c := connectToWebsocket(u)
    defer c.Close()

	done := make(chan struct{})

    go listen(c, &done)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
	        closeConnection(c, &done)
            return
        }
	}
}

func connectToWebsocket(endpointUrl url.URL) *websocket.Conn {
    urlString := endpointUrl.String()
    log.Printf("connecting to %s", urlString)

	c, _, err := websocket.DefaultDialer.Dial(urlString, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
    return c
}

func listen(c *websocket.Conn, done *chan struct{}) {
    defer close(*done)
    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            return
        }
        log.Printf("recv: %s", message)
    }
}

func closeConnection(c *websocket.Conn, done *chan struct{}) {
    log.Println("interrupt")

    // Cleanly close the connection by sending a close message and then
    // waiting (with timeout) for the server to close the connection.
    err := c.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    if err != nil {
        log.Println("write close:", err)
        return
    }
    select {
    case <-*done:
    case <-time.After(time.Second):
    }
}
