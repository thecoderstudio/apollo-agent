package main

import (
    "log"
    "net/url"
    "os"
    "os/signal"

    "github.com/jessevdk/go-flags"

    "github.com/thecoderstudio/apollo-agent/client"
    "github.com/thecoderstudio/apollo-agent/oauth"
)

var opts struct {
    Host string `short:"h" long:"host" description:"Host address" required:"true"`
    AgentID string `long:"agent-id" description:"Apollo agent id" required:"true"`
    Secret string `long:"secret" description:"Apollo OAuth client secret" required:"true"`
}

func main() {
    log.SetFlags(0)
    _, err := flags.Parse(&opts)
    if err != nil {
        panic(err)
    }
    host := opts.Host

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    var initialToken oauth.AccessToken
    oauthClient := oauth.Create(host, opts.AgentID, opts.Secret)
    accessTokenChan, oauthErrs := oauthClient.GetContinuousAccessToken()
    select {
    case newAccessToken := <-*accessTokenChan:
        initialToken = newAccessToken
    case err := <-*oauthErrs:
        log.Fatal(err)
    }

    u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
    wsClient := client.Create(new(client.DialWrapper))
    out, done, errs := wsClient.Listen(u, initialToken, &interrupt)
    for {
        select {
        case newAccessToken := <-*accessTokenChan:
            out, done, errs = wsClient.Listen(u, newAccessToken, &interrupt)
        case msg := <-out:
            log.Println(msg)
        case err := <-errs:
            log.Println(err)
        case <-done:
            return
        }
    }
}
