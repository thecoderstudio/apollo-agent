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

const initIndication = "start"
const finishIndication = "done"
const fakeConnectionID = "test"

func TestWaitForCompletion(t *testing.T) {
	expectedFinishedCommand := websocket.Command{
		ConnectionID: fakeConnectionID,
		Command:      "finished",
	}

	broadcaster := broadcast.NewBroadcaster(512)

	sessionMock := new(mocks.BaseSession)
	sessionMock.On("SessionID").Return(fakeConnectionID)
	sessionMock.On("Out").Return(&broadcaster)

	commandObserver := action.CreateCommandObserver(
		initIndication,
		finishIndication,
	)
	out := *commandObserver.CommandOutput()
	go commandObserver.WaitForCompletion(sessionMock)

	time.Sleep(time.Second)
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: fakeConnectionID,
		Message:      initIndication,
	})
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: fakeConnectionID,
		Message:      "testing",
	})
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: fakeConnectionID,
		Message:      finishIndication,
	})

	finishedCommand := <-out
	assert.Equal(t, finishedCommand, expectedFinishedCommand)
}

func TestCreateCommandObserver(t *testing.T) {
	commandObserver := action.CreateCommandObserver(initIndication, finishIndication)
	assert.Equal(t, commandObserver.InitialisationIndication, initIndication)
	assert.Equal(t, commandObserver.CompletionIndication, finishIndication)
	assert.NotNil(t, *commandObserver.CommandOutput())
	assert.NotNil(t, commandObserver)
}
