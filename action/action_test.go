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

const command = "curl https://raw.githubusercontent.com/carlospolop/" +
	"privilege-escalation-awesome-scripts-suite/master/linPEAS/" +
	"linpeas.sh | sh && echo 'linPEAS done\n'\n"
const connectionID = "test"
const initialisationIndication = "Green"
const completionIndication = "linPEAS done"

func TestRun(t *testing.T) {
	expectedFinishedCommand := websocket.Command{
		ConnectionID: connectionID,
		Command:      "finished",
	}

	broadcaster := broadcast.NewBroadcaster(512)

	sessionMock := new(mocks.BaseSession)
	sessionMock.On("SessionID").Return(connectionID)
	sessionMock.On("Execute", command).Return(nil)
	sessionMock.On("Out").Return(&broadcaster)

	linPeas := action.CreateLinPeas(sessionMock)
	out := *linPeas.Run()

	time.Sleep(2 * time.Second)
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: connectionID,
		Message:      initialisationIndication,
	})
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: connectionID,
		Message:      "testing",
	})
	broadcaster.Submit(websocket.ShellIO{
		ConnectionID: connectionID,
		Message:      completionIndication,
	})

	finishedCommand := <-out
	assert.Equal(t, finishedCommand, expectedFinishedCommand)
}
