package main

import (
    "encoding/json"
    "log"
    "net/url"
    "os"
    "os/signal"

    "github.com/jessevdk/go-flags"

    "github.com/thecoderstudio/apollo-agent/client"
    "github.com/thecoderstudio/apollo-agent/oauth"
    "github.com/thecoderstudio/apollo-agent/shell"
)

var opts struct {
    Host string `short:"h" long:"host" description:"Host address" required:"true"`
    AgentID string `long:"agent-id" description:"Apollo agent id" required:"true"`
    Secret string `long:"secret" description:"Apollo OAuth client secret" required:"true"`
}

func main() {
    log.SetFlags(0)
    parseArguments()

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    accessTokenChan, initialToken := setupOAuth()
    connect(accessTokenChan, initialToken, &interrupt)
    }

func parseArguments() {
    _, err := flags.Parse(&opts)
    if err != nil {
        panic(err)
    }
}

func setupOAuth() (*chan oauth.AccessToken, oauth.AccessToken) {
    var initialToken oauth.AccessToken
    client := oauth.Create(opts.Host, opts.AgentID, opts.Secret)
    accessTokenChan, oauthErrs := client.GetContinuousAccessToken()

    select {
    case newAccessToken := <-*accessTokenChan:
        initialToken = newAccessToken
    case err := <-*oauthErrs:
        log.Fatal(err)
    }

    return accessTokenChan, initialToken
}

func connect(accessTokenChan *chan oauth.AccessToken, initialToken oauth.AccessToken,
             interruptSignal *chan os.Signal) {
    u := url.URL{Scheme: "ws", Host: opts.Host, Path: "/ws"}
    wsClient := client.Create(new(client.DialWrapper))
    interrupt := make(chan struct{})
    out, done, errs := wsClient.Listen(u, initialToken, &interrupt)

    for {
        select {
        case newAccessToken := <-*accessTokenChan:
            previousInterrupt := interrupt
            interrupt = make(chan struct{})
            out, done, errs = wsClient.Listen(u, newAccessToken, &interrupt)
            close(previousInterrupt)
        case msg := <-out:
            message := client.Message{}
            json.Unmarshal([]byte(msg), &message)
            stdout := shell.Execute(message.Message)
            log.Println(stdout)
        case err := <-errs:
            log.Println(err)
        case <-*interruptSignal:
            close(interrupt)
        case <-done:
            return
        }
    }
}
