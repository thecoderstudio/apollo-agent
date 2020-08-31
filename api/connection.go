package api

import (
	"log"
	"net/url"
	"os"

	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// Middleware is responsible for dealing with the API by handling both authentication and communication.
type Middleware struct {
	Host            string
	InterruptSignal *chan os.Signal
}

// Start starts the communication with the API by authenticating and maintaining the connection. Incoming websocket
// commands will be forwarded to the manager.
func (middleware *Middleware) Start(agentID, secret, shell string) {
	accessTokenChan, initialToken := middleware.authenticate(agentID, secret)
	middleware.connect(accessTokenChan, initialToken, shell)
}

func (middleware *Middleware) authenticate(agentID, secret string) (*chan oauth.AccessToken, oauth.AccessToken) {
	var initialToken oauth.AccessToken
	client := oauth.Create(middleware.Host, agentID, secret)
	accessTokenChan, oauthErrs := client.GetContinuousAccessToken()

	select {
	case newAccessToken := <-*accessTokenChan:
		initialToken = newAccessToken
	case err := <-*oauthErrs:
		log.Fatal(err)
	}

	return accessTokenChan, initialToken
}

func (middleware *Middleware) connect(
	accessTokenChan *chan oauth.AccessToken,
	accessToken oauth.AccessToken,
	shell string,
) {
	u := url.URL{Scheme: "ws", Host: middleware.Host, Path: "/ws"}
	wsClient := websocket.CreateClient(new(websocket.DialWrapper))

	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	ptyManager := pty.CreateManager(&in, shell)
	defer ptyManager.Close()

	done := wsClient.Listen(u, accessToken, &in, &interrupt)

	for {
		select {
		case newAccessToken := <-*accessTokenChan:
			previousInterrupt := interrupt
			accessToken = newAccessToken
			interrupt = make(chan struct{})
			done = wsClient.Listen(u, newAccessToken, &in, &interrupt)
			close(previousInterrupt)
		case shellIO := <-wsClient.Out():
			go ptyManager.Execute(shellIO)
		case command := <-wsClient.Commands():
			ptyManager.ExecutePredefinedCommand(command)
		case err := <-wsClient.Errs():
			log.Println(err)
			done = wsClient.Listen(u, accessToken, &in, &interrupt)
		case <-*middleware.InterruptSignal:
			close(interrupt)
		case <-done:
			return
		}
	}
}

func (middleware *Middleware) reconnect() {

}
