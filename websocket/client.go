package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/thecoderstudio/apollo-agent/oauth"
)

// Connection specifies the interface for client connection
type Connection interface {
	Close() error
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}

// Dialer should be used to dial TCP connections returning a connection
// that fits the expected Connection interface.
type Dialer interface {
	Dial(string, http.Header) (Connection, *http.Response, error)
}

// DialWrapper wraps the websocket.DefaultDialer to transform returned types
// to interfaces the client expects.
type DialWrapper struct{}

// Dial wraps de default Dial to return an interface instead of a struct.
func (wrapper DialWrapper) Dial(urlString string, header http.Header) (Connection, *http.Response, error) {
	return websocket.DefaultDialer.Dial(urlString, header)
}

// RemoteTerminal is an interface for structs that listen for remote commands and take the
// output to these commands to send back.
type RemoteTerminal interface {
	Out() <-chan ShellIO
	Commands() <-chan Command
	Errs() <-chan error
	Listen(url.URL, oauth.AccessToken, <-chan ShellIO) <-chan struct{}
	Interrupt() chan struct{}
}

// Client is used to connect over the WebSocket protocol and receive as well as send messages.
type Client struct {
	dialer      Dialer
	out         chan ShellIO
	commands    chan Command
	errs        chan error
	interrupt   chan struct{}
	interrupted bool
}

// Out contains received shell messages.
func (client Client) Out() <-chan ShellIO {
	return client.out
}

// Commands contains received pre-defined commands.
func (client Client) Commands() <-chan Command {
	return client.commands
}

// Errs contains any errors that occur.
func (client Client) Errs() <-chan error {
	return client.errs
}

// Interrupt allows for the interruption of the listening of the client. You can do this by
// sending a message to the channel returned by Interrupt or by simply closing it.
func (client Client) Interrupt() chan struct{} {
	return client.interrupt
}

// Listen connects to the given endpoint and handles incoming messages. It's interruptable
// by closing the interrupt channel. Outgoing communication send through `in` are sent to Apollo.
func (client Client) Listen(
	endpointURL url.URL,
	accessToken oauth.AccessToken,
	in <-chan ShellIO,
) <-chan struct{} {
	done := make(chan struct{})
	client.interrupted = false

	go func() {
		connection, err := client.createConnection(endpointURL, accessToken)
		if err != nil {
			log.Println("Connection error")
			client.errs <- err
			close(done)
			return
		}

		// Used by awaitMessages to communicate when done. Only when done can we close the channels
		// to prevent sending messages to closed channels.
		doneListening := make(chan struct{})

		log.Println("Connected")
		go client.awaitMessages(&connection, &done, &doneListening)
		err = client.handleEvents(&connection, in, &doneListening)

		connection.Close()
		log.Println("Disconnected")
		close(done)
		<-doneListening
	}()

	return done
}

func (client *Client) createConnection(endpointURL url.URL, accessToken oauth.AccessToken) (Connection, error) {
	urlString := endpointURL.String()
	log.Printf("Connecting to %s", urlString)
	authorizationHeader := fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken)
	connection, _, err := client.dialer.Dial(urlString, http.Header{"Authorization": []string{authorizationHeader}})
	return connection, err
}

func (client *Client) awaitMessages(connection *Connection, done, doneListening *chan struct{}) {
	defer close(*doneListening)

	conn := *connection
	for {
		select {
		case <-*done:
			return
		default:
			_, rawMessage, err := conn.ReadMessage()
			if err != nil {
				if client.interrupted {
					return
				}
				log.Println("Read error:", err)
				client.errs <- err
				return
			}

			client.sendOverChannels([]byte(rawMessage))
		}
	}
}

func (client *Client) sendOverChannels(rawMessage []byte) {
	shellIO := ShellIO{}
	command := Command{}

	json.Unmarshal(rawMessage, &shellIO)
	json.Unmarshal(rawMessage, &command)

	switch {
	case command.Command != "":
		client.commands <- command
	case shellIO.Message != "":
		client.out <- shellIO
	default:
		log.Println("Message skipped")
	}
}

func (client *Client) handleEvents(
	connection *Connection,
	in <-chan ShellIO,
	doneListening *chan struct{},
) error {
	log.Println("Listening..")
	for {
		select {
		case <-*doneListening:
			return nil
		case message := <-in:
			conn := *connection
			jsonMessage, _ := json.Marshal(message)
			conn.WriteMessage(websocket.TextMessage, jsonMessage)
		case <-client.interrupt:
			log.Println("Interrupted")
			client.interrupted = true
			err := client.closeConnection(connection)
			return err
		}
	}
}

func (client *Client) closeConnection(connection *Connection) error {
	log.Println("Shutting down gracefully")
	conn := *connection
	err := conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		log.Println("Close err:", err)
		return err
	}

	return nil
}

// CreateClient is the factory to create a properly instantiated client.
func CreateClient(dialer Dialer) Client {
	out := make(chan ShellIO)
	commands := make(chan Command)
	errs := make(chan error)
	interrupt := make(chan struct{})
	return Client{dialer, out, commands, errs, interrupt, false}
}
