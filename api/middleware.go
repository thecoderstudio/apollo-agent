package api

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/thecoderstudio/apollo-agent/oauth"
	"github.com/thecoderstudio/apollo-agent/pty"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

// Middleware is responsible for dealing with the API by handling both authentication and communication.
type Middleware struct {
	Host            string
	InterruptSignal *chan os.Signal
	WebsocketClient websocket.Client
	OAuthClient     oauth.AuthProvider
}

// Start starts the communication with the API by authenticating and maintaining the connection. Incoming websocket
// commands will be forwarded to the manager.
func (middleware *Middleware) Start(shell string) {
	accessTokenChan, initialToken := middleware.authenticate()
	middleware.connect(accessTokenChan, initialToken, shell)
}

func (middleware *Middleware) authenticate() (*chan oauth.AccessToken, oauth.AccessToken) {
	var initialToken oauth.AccessToken
	accessTokenChan, oauthErrs := middleware.OAuthClient.GetContinuousAccessToken()

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

	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	ptyManager := pty.CreateManager(&in, shell)
	defer ptyManager.Close()

	done := middleware.WebsocketClient.Listen(u, accessToken, &in, &interrupt)

	for {
		select {
		case newAccessToken := <-*accessTokenChan:
			previousInterrupt := interrupt
			accessToken = newAccessToken
			interrupt = make(chan struct{})
			done = middleware.WebsocketClient.Listen(u, newAccessToken, &in, &interrupt)
			close(previousInterrupt)
		case shellIO := <-middleware.WebsocketClient.Out():
			go ptyManager.Execute(shellIO)
		case command := <-middleware.WebsocketClient.Commands():
			ptyManager.ExecutePredefinedCommand(command)
		case err := <-middleware.WebsocketClient.Errs():
			log.Println(err)
			done = nil
			go func() {
				done = middleware.reconnect(u, accessToken, &in, &interrupt)
			}()
		case <-*middleware.InterruptSignal:
			close(interrupt)
			if done == nil {
				return
			}
		case <-done:
			return
		}
	}
}

func (middleware *Middleware) reconnect(
	u url.URL,
	accessToken oauth.AccessToken,
	in *chan websocket.ShellIO,
	interrupt *chan struct{},
) <-chan struct{} {
	time.Sleep(5 * time.Second)
	return middleware.WebsocketClient.Listen(u, accessToken, in, interrupt)
}

// CreateMiddleware is the factory to create a properly instantiated middleware.
func CreateMiddleware(host, agentID, secret string, interruptSignal *chan os.Signal) Middleware {
	wsClient := websocket.CreateClient(new(websocket.DialWrapper))
	oauthClient := oauth.Create(host, agentID, secret)
	return Middleware{host, interruptSignal, wsClient, oauthClient}
}
