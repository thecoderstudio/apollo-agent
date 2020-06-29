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

// ShellCommunication is used for communicating shell input, output and error streams.
type ShellCommunication struct {
	ConnectionID string `json:"connection_id"`
	Message      string `json:"message"`
}

// Command is used to instruct the agent to execute a pre-defined command.
type Command struct {
    ConnectionID    string `json:"connection_id"`
    Command         string `json:"command"`
}

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

// Client is used to connect over the WebSocket protocol and receive as well as send messages.
type Client struct {
	dialer Dialer
}

// Listen connects to the given endpoint and handles incoming messages. It's interruptable
// by closing the interrupt channel. Outgoing communication send through in are sent to Apollo.
func (client *Client) Listen(
    endpointURL url.URL,
    accessToken oauth.AccessToken,
	in *chan ShellCommunication,
    interrupt *chan struct{},
) (<-chan ShellCommunication, <-chan Command, <-chan struct{}, <-chan error) {
	out := make(chan ShellCommunication)
	commands := make(chan Command)
	errs := make(chan error)
	done := make(chan struct{})

	go func() {
		defer close(out)
		defer close(errs)

		connection, err := client.createConnection(endpointURL, accessToken)
		if err != nil {
			log.Println("Connection error")
			errs <- err
			close(done)
			return
		}

		// Used by awaitMessages to communicate when done. Only when done can we close the channels
		// to prevent sending messages to closed channels.
		doneListening := make(chan struct{})

		go client.awaitMessages(&connection, &out, &commands, &errs, &done, &doneListening)
		err = client.handleEvents(&connection, in, &doneListening, interrupt)

		connection.Close()
		close(done)
		<-doneListening
	}()

	return out, commands, done, errs
}

func (client *Client) createConnection(endpointURL url.URL, accessToken oauth.AccessToken) (Connection, error) {
	urlString := endpointURL.String()
	log.Printf("connecting to %s", urlString)
	authorizationHeader := fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken)
	connection, _, err := client.dialer.Dial(urlString, http.Header{"Authorization": []string{authorizationHeader}})
	return connection, err
}

func (client *Client) awaitMessages(
    connection *Connection,
    out *chan ShellCommunication,
    commands *chan Command,
    errs *chan error,
    done, doneListening *chan struct{},
) {
	defer close(*doneListening)

	conn := *connection
	for {
		select {
		case <-*done:
			return
        default:
			_, rawMessage, err := conn.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				*errs <- err
				return
			}

			shellComm := ShellCommunication{}
            command := Command{}
            rawMessageBytes := []byte(rawMessage)


			json.Unmarshal(rawMessageBytes, &shellComm)
			json.Unmarshal(rawMessageBytes, &command)

            switch {
            case command.Command != "":
                *commands <- command
            case shellComm.Message != "":
                *out <- shellComm
            default:
                log.Println("Message skipped")
            }
		}
	}
}

func (client *Client) handleEvents(connection *Connection, in *chan ShellCommunication,
	doneListening *chan struct{},
	interrupt *chan struct{}) error {
	for {
		select {
		case <-*doneListening:
			return nil
		case message := <-*in:
			conn := *connection
			jsonMessage, _ := json.Marshal(message)
			conn.WriteMessage(websocket.TextMessage, jsonMessage)
		case <-*interrupt:
			log.Println("interrupt")
			err := client.closeConnection(connection)
			return err
		}
	}
}

func (client *Client) closeConnection(connection *Connection) error {
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
	return Client{dialer: dialer}
}
