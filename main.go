package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/jessevdk/go-flags"

	"github.com/thecoderstudio/apollo-agent/net"
	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

var opts struct {
	Host    string `short:"h" long:"host" description:"Host address" required:"true"`
	AgentID string `long:"agent-id" description:"Apollo agent id" required:"true"`
	Secret  string `long:"secret" description:"Apollo OAuth client secret" required:"true"`
	Shell   string `long:"shell" description:"Path to shell" default:"/bin/bash"`
}

func main() {
	log.SetFlags(0)
	parseArguments()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	host := net.GetHostFromURLString(opts.Host)
	if host == "" {
		log.Fatal("No valid host given")
	}
	accessTokenChan, initialToken := setupOAuth(host, opts.AgentID, opts.Secret)
	connect(host, accessTokenChan, initialToken, &interrupt)
}

func parseArguments() {
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}
}

func setupOAuth(host, agentID, secret string) (*chan oauth.AccessToken, oauth.AccessToken) {
	var initialToken oauth.AccessToken
	client := oauth.Create(host, agentID, secret)
	accessTokenChan, oauthErrs := client.GetContinuousAccessToken()

	select {
	case newAccessToken := <-*accessTokenChan:
		initialToken = newAccessToken
	case err := <-*oauthErrs:
		log.Fatal(err)
	}

	return accessTokenChan, initialToken
}

func connect(
	host string,
	accessTokenChan *chan oauth.AccessToken,
	initialToken oauth.AccessToken,
	interruptSignal *chan os.Signal,
) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	wsClient := websocket.CreateClient(new(websocket.DialWrapper))

	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	ptyManager := pty.CreateManager(&in, opts.Shell)
	defer ptyManager.Close()

	done := wsClient.Listen(u, initialToken, &in, &interrupt)

	for {
		select {
		case newAccessToken := <-*accessTokenChan:
			previousInterrupt := interrupt
			interrupt = make(chan struct{})
			done = wsClient.Listen(u, newAccessToken, &in, &interrupt)
			close(previousInterrupt)
		case shellIO := <-wsClient.Out():
			go ptyManager.Execute(shellIO)
		case command := <-wsClient.Commands():
			ptyManager.ExecutePredefinedCommand(command)
		case err := <-wsClient.Errs():
			log.Println(err)
		case <-*interruptSignal:
			close(interrupt)
		case <-done:
			return
		}
	}
}
