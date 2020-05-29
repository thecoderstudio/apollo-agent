package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

    "github.com/thecoderstudio/apollo-agent/client"
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
    wsClient := client.Create(client.DialWrapper{})
    wsClient.Listen(u, &interrupt)
}
