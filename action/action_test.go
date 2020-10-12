package action_test

import (
	"testing"
	"time"

	"github.com/dustin/go-broadcast"
	"github.com/stretchr/testify/assert"

	"github.com/thecoderstudio/apollo-agent/action"
	"github.com/thecoderstudio/apollo-agent/mocks"
	"github.com/thecoderstudio/apollo-agent/websocket"
)

func TestExecuteLinPeas(t *testing.T) {
	expectedFinishedCommand := websocket.Command{
		ConnectionID: linPeasConnectionID,
		Command:      "finished",
	}

	broadcaster := broadcast.NewBroadcaster(512)

	sessionMock := new(mocks.BaseSession)
	sessionMock.On("SessionID").Return(linPeasConnectionID)
	sessionMock.On("Execute", command).Return(nil)
	sessionMock.On("Out").Return(&broadcaster)

	out, err := action.Execute(
		sessionMock,
		websocket.Command{
			ConnectionID: linPeasConnectionID,
			Command:      "linpeas",
		},
	)

	time.Sleep(2 * time.Second)
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: linPeasConnectionID,
		Message:      linPeasInitIndication,
	})
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: linPeasConnectionID,
		Message:      "testing",
	})
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: linPeasConnectionID,
		Message:      linPeasCompletionIndication,
	})

	finishedCommand := <-*out

	assert.Equal(t, finishedCommand, expectedFinishedCommand)
	assert.Nil(t, err)
}

func TestExecuteActionNotFound(t *testing.T) {
	sessionMock := new(mocks.BaseSession)
	out, err := action.Execute(
		sessionMock,
		websocket.Command{
			ConnectionID: linPeasConnectionID,
			Command:      "fake",
		},
	)

	assert.Nil(t, out)
	assert.EqualError(t, err, "action not found for given command: fake")
}
