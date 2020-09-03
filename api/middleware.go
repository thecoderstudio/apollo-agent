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
	RemoteTerminal  websocket.RemoteTerminal
	PTYManager      pty.ShellManager
	OAuthClient     oauth.AuthProvider
	Connected       chan bool
}

// Start starts the communication with the API by authenticating and maintaining the connection. Incoming websocket
// commands will be forwarded to the manager.
func (middleware *Middleware) Start() error {
	accessTokenChan, initialToken, err := middleware.authenticate()
	if err != nil {
		return err
	}

	middleware.connect(accessTokenChan, initialToken)
	return nil
}

func (middleware *Middleware) authenticate() (*chan oauth.AccessToken, oauth.AccessToken, error) {
	var err error
	var initialToken oauth.AccessToken
	accessTokenChan, oauthErrs := middleware.OAuthClient.GetContinuousAccessToken()

	select {
	case newAccessToken := <-*accessTokenChan:
		initialToken = newAccessToken
	case authErr := <-*oauthErrs:
		err = authErr
	}

	return accessTokenChan, initialToken, err
}

func (middleware *Middleware) connect(
	accessTokenChan *chan oauth.AccessToken,
	accessToken oauth.AccessToken,
) {
	u := url.URL{Scheme: "ws", Host: middleware.Host, Path: "/ws"}
	interrupt := make(chan struct{})
	defer middleware.PTYManager.Close()

	done := middleware.RemoteTerminal.Listen(u, accessToken, middleware.PTYManager.Out(), &interrupt)

	for {
		select {
		case newAccessToken := <-*accessTokenChan:
			previousInterrupt := interrupt
			accessToken = newAccessToken
			interrupt = make(chan struct{})
			done = middleware.RemoteTerminal.Listen(u, newAccessToken, middleware.PTYManager.Out(), &interrupt)
			close(previousInterrupt)
		case shellIO := <-middleware.RemoteTerminal.Out():
			go middleware.PTYManager.Execute(shellIO)
		case command := <-middleware.RemoteTerminal.Commands():
			middleware.PTYManager.ExecutePredefinedCommand(command)
		case err := <-middleware.RemoteTerminal.Errs():
			log.Println(err)
			go middleware.setConnection(false)
			done = middleware.reconnect(u, accessToken, middleware.PTYManager.Out(), &interrupt)
			go middleware.setConnection(true)
		case <-*middleware.InterruptSignal:
			close(interrupt)
		case <-done:
			go middleware.setConnection(false)
			return
		}
	}
}

func (middleware *Middleware) setConnection(connected bool) {
	middleware.Connected <- connected
}

func (middleware *Middleware) reconnect(
	u url.URL,
	accessToken oauth.AccessToken,
	in <-chan websocket.ShellIO,
	interrupt *chan struct{},
) <-chan struct{} {
	time.Sleep(5 * time.Second)
	return middleware.RemoteTerminal.Listen(u, accessToken, in, interrupt)
}

// CreateMiddleware is the factory to create a properly instantiated middleware.
func CreateMiddleware(host, agentID, secret, shell string, interruptSignal *chan os.Signal) (Middleware, error) {
	var middleware Middleware
	ptyManager, err := pty.CreateManager(shell)
	if err != nil {
		return middleware, err
	}

	wsClient := websocket.CreateClient(new(websocket.DialWrapper))
	oauthClient := oauth.Create(host, agentID, secret)
	connected := make(chan bool)
	middleware = Middleware{host, interruptSignal, wsClient, ptyManager, oauthClient, connected}
	return middleware, err
}
