package api

import (
	"net/url"
	"os"
	"time"

	"github.com/thecoderstudio/apollo-agent/logging"
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
}

// Start starts the communication with the API by authenticating and maintaining the connection. Incoming websocket
// commands will be forwarded to the manager.
func (middleware *Middleware) Start(reconnectInterval time.Duration) error {
	accessTokenChan, initialToken, err := middleware.authenticate()
	if err != nil {
		return err
	}

	middleware.connect(accessTokenChan, initialToken, reconnectInterval)
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
	reconnectInterval time.Duration,
) {
	u := url.URL{Scheme: "ws", Host: middleware.Host, Path: "/ws"}
	defer middleware.PTYManager.Close()

	done := middleware.RemoteTerminal.Listen(u, accessToken, middleware.PTYManager.Out())

	for {
		select {
		case newAccessToken := <-*accessTokenChan:
			logging.Info("Reauthenticating..")
			accessToken = newAccessToken
			middleware.RemoteTerminal.Interrupt() <- struct{}{}
			done = middleware.RemoteTerminal.Listen(u, newAccessToken, middleware.PTYManager.Out())
		case shellIO := <-middleware.RemoteTerminal.Out():
			go middleware.PTYManager.Execute(shellIO)
		case command := <-middleware.RemoteTerminal.Commands():
			middleware.PTYManager.ExecutePredefinedCommand(command)
		case err := <-middleware.RemoteTerminal.Errs():
			logging.Critical(err)
			var interrupted bool
			done, interrupted = middleware.reconnect(u, accessToken, middleware.PTYManager.Out(), reconnectInterval)
			if interrupted {
				return
			}
		case <-*middleware.InterruptSignal:
			close(middleware.RemoteTerminal.Interrupt())
		case <-done:
			return
		}
	}
}

func (middleware *Middleware) reconnect(
	u url.URL,
	accessToken oauth.AccessToken,
	in <-chan websocket.ShellIO,
	reconnectInterval time.Duration,
) (<-chan struct{}, bool) {
	logging.Info("Reconnecting..")
	timerEnded, interrupted := middleware.startInterruptableTimer(reconnectInterval)
	select {
	case <-interrupted:
		return nil, true
	case <-timerEnded:
		return middleware.RemoteTerminal.Listen(u, accessToken, in), false
	}
}

func (middleware *Middleware) startInterruptableTimer(duration time.Duration) (<-chan struct{}, <-chan struct{}) {
	ended, interrupted := make(chan struct{}), make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		go func() {
			for {
				select {
				case <-*middleware.InterruptSignal:
					close(interrupted)
				case <-ticker.C:
					continue
				case <-ended:
					return
				}
			}
		}()
		time.Sleep(duration)
		ticker.Stop()
		close(ended)
	}()
	return ended, interrupted
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
	middleware = Middleware{host, interruptSignal, wsClient, ptyManager, oauthClient}
	return middleware, err
}
