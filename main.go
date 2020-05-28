package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "", "host address")

func main() {
	flag.Parse()
	log.SetFlags(0)

    if *addr == "" {
        log.Fatal("--addr is required, please give a valid host address")
    }

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
    client := WebsocketClient{}
    client.connectAndListen(u, &interrupt)
}
