package main

import (
    "flag"
    "log"
    "net/url"
    "os"
    "os/signal"

    "github.com/thecoderstudio/apollo-agent/client"
    "github.com/thecoderstudio/apollo-agent/oauth"
)

var addr = flag.String("addr", "", "host address")
var agentID = flag.String("agent-id", "", "Apollo agent id")
var secret = flag.String("secret", "", "Apollo OAuth client secret")

func main() {
    flag.Parse()
    log.SetFlags(0)

    if *addr == "" {
        log.Fatal("--addr is required, please give a valid host address")
    }

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    oauthClient := oauth.Create(*addr, *agentID, *secret)
    newAccessToken := oauthClient.GetContinuousAccessToken()
    accessToken := <-*newAccessToken

    u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
    wsClient := client.Create(new(client.DialWrapper))
    out, done, errs := wsClient.Listen(u, accessToken, &interrupt)
    for {
        select {
        case msg := <-out:
            log.Println(msg)
        case err := <-errs:
            log.Println(err)
        case <-done:
            return
        }
    }
}
