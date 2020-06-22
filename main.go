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
    in := make(chan client.Message)
    out, done, errs := wsClient.Listen(u, initialToken, &in, &interrupt)
    sessions := map[string] *shell.PTYSession{}

    for {
        select {
        case newAccessToken := <-*accessTokenChan:
            previousInterrupt := interrupt
            interrupt = make(chan struct{})
            out, done, errs = wsClient.Listen(u, newAccessToken, &in, &interrupt)
            close(previousInterrupt)
        case msg := <-out:
            message := client.Message{}
            json.Unmarshal([]byte(msg), &message)
            if pty, ok := sessions[message.SessionID]; ok {
                pty.Execute(message.Message)
            } else {
                pty := shell.CreateNewPTY(message.SessionID)
                pty.Execute(message.Message)

                go writeOutput(pty.Out, &in)
                sessions[message.SessionID] = pty
            }
        case err := <-errs:
            log.Println(err)
        case <-*interruptSignal:
            close(interrupt)
        case <-done:
            return
        }
    }
}

func writeOutput(in *chan client.Message, out *chan client.Message) {
    for {
        message := <-*in
        *out <- message
    }
}
