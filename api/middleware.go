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
	ShellInterface  websocket.ShellInterface
	PTYManager      pty.ShellManager
	OAuthClient     oauth.AuthProvider
}

// Start starts the communication with the API by authenticating and maintaining the connection. Incoming websocket
// commands will be forwarded to the manager.
func (middleware *Middleware) Start() {
	accessTokenChan, initialToken := middleware.authenticate()
	middleware.connect(accessTokenChan, initialToken)
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
) {
	u := url.URL{Scheme: "ws", Host: middleware.Host, Path: "/ws"}

	interrupt := make(chan struct{})
	in := make(chan websocket.ShellIO)
	defer close(in)

	defer middleware.PTYManager.Close()

	done := middleware.ShellInterface.Listen(u, accessToken, &in, &interrupt)

	for {
		select {
		case newAccessToken := <-*accessTokenChan:
			previousInterrupt := interrupt
			accessToken = newAccessToken
			interrupt = make(chan struct{})
			done = middleware.ShellInterface.Listen(u, newAccessToken, &in, &interrupt)
			close(previousInterrupt)
		case shellIO := <-middleware.ShellInterface.Out():
			go middleware.PTYManager.Execute(shellIO)
		case command := <-middleware.ShellInterface.Commands():
			middleware.PTYManager.ExecutePredefinedCommand(command)
		case err := <-middleware.ShellInterface.Errs():
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
	return middleware.ShellInterface.Listen(u, accessToken, in, interrupt)
}

// CreateMiddleware is the factory to create a properly instantiated middleware.
func CreateMiddleware(host, agentID, secret, shell string, interruptSignal *chan os.Signal) Middleware {
	wsClient := websocket.CreateClient(new(websocket.DialWrapper))
	ptyManager := pty.CreateManager(shell)
	oauthClient := oauth.Create(host, agentID, secret)
	return Middleware{host, interruptSignal, wsClient, ptyManager, oauthClient}
}
