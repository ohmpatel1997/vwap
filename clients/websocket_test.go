package clients

// based on https://github.com/aglyzov/ws-machine
import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WebSocketSuite struct {
	suite.Suite
	logger *log.Logger
}

type Echoer struct {
}

var upgrader = websocket.Upgrader{} // use default options

func NewEchoer() *Echoer {
	return new(Echoer)
}

func (e *Echoer) ServeHTTP(ans http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(ans, req, nil)
	if err != nil {
		log.WithField("error", err).Error("HTTP upgrade error")
		return
	}
	defer ws.Close()
	for {
		mt, p, err := ws.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.WithField("error", err).Error("Echoer: ReadMessage error")
			}
			return
		}
		if bytes.Equal(p, []byte("/CLOSE")) {
			log.Debug("Echoer: closing connection by client request")
			return
		}
		if err := ws.WriteMessage(mt, p); err != nil {
			log.WithField("error", err).Error("Echoer: WriteMessage error")
			return
		}
	}
}

func httpToWs(u string) string {
	return "ws" + u[len("http"):]
}

func TestWebSocketSuite(t *testing.T) {
	suite.Run(t, new(WebSocketSuite))
}

func (suite *WebSocketSuite) SetupTest() {
	suite.logger = log.New()
	suite.logger.Level = log.DebugLevel
}

func (suite *WebSocketSuite) wrongState(expected, state WSState, err error) {
	suite.logger.WithFields(log.Fields{
		"expectedState": expected,
		"state":         state,
		"error":         err,
	}).Error("wrong state")
	suite.T().Fatal("wrong state", "expectedState", expected, "state", state, "error", err)
}

func (suite *WebSocketSuite) TestBadURL() {
	url := "ws://websocket.bad.url/"
	ws := NewWebSocket(suite.logger, url, http.Header{})
	assert.Equal(suite.T(), url, ws.URL())
	assert.Equal(suite.T(), http.Header{}, ws.Headers())

	ws.wg.Add(1)

	go ws.DriverProgram()

	statusCh := ws.Status()

	fmt.Println("1")
	st := <-statusCh
	if st.State != WS_CONNECTING {
		suite.wrongState(WS_CONNECTING, st.State, st.Error)
	}

	fmt.Println("2")
	st = <-statusCh
	if st.State != WS_WAITING {
		suite.wrongState(WS_WAITING, st.State, st.Error)
	}

	fmt.Println("3")
	st = <-statusCh
	if st.State != WS_DISCONNECTED {
		suite.wrongState(WS_DISCONNECTED, st.State, st.Error)
	}

}

func (suite *WebSocketSuite) TestConnect() {
	srv := httptest.NewServer(NewEchoer())
	defer srv.Close()

	URL := httpToWs(srv.URL)
	suite.logger.WithField("URL", URL).Debug("set server URL")

	ws := NewWebSocket(suite.logger, URL, http.Header{})

	ws.wg.Add(1)

	go ws.DriverProgram()

	statusCh := ws.Status()
	cmdCh := ws.Command()

	st := <-statusCh
	if st.State != WS_CONNECTING {
		suite.T().Fatal(WS_CONNECTING, st.State, st.Error)
	}
	st = <-statusCh
	if st.State != WS_CONNECTED {
		suite.T().Fatal(WS_CONNECTED, st.State, st.Error)
	}

	// cleanup
	cmdCh <- WS_QUIT

	st = <-statusCh
	if st.State != WS_DISCONNECTED {
		suite.T().Fatal(WS_DISCONNECTED, st.State, st.Error)
	}

	srv.Close()
}

func (suite *WebSocketSuite) TestEcho() {
	srv := httptest.NewServer(NewEchoer())
	defer srv.Close()

	URL := httpToWs(srv.URL)
	suite.logger.WithField("URL", URL).Debug("set server URL")

	ws := NewWebSocket(suite.logger, URL, http.Header{})

	ws.wg.Add(1)

	go ws.DriverProgram()

	statusCh := ws.Status()
	inputCh := ws.Input()
	outCh := ws.Output()
	cmdCh := ws.Command()

	// send a message right away
	orig := []byte("Test Message")
	outCh <- orig

	if st := <-statusCh; st.State != WS_CONNECTING {
		suite.wrongState(WS_CONNECTING, st.State, st.Error)
	}
	if st := <-statusCh; st.State != WS_CONNECTED {
		suite.wrongState(WS_CONNECTED, st.State, st.Error)
	}

	// wait for an answer
	select {
	case msg, ok := <-inputCh:
		if !ok {
			suite.logger.Error("ws.Input unexpectedly closed")
		} else if !bytes.Equal(msg, orig) {
			suite.logger.WithFields(log.Fields{
				"expectedMsg": string(orig),
				"receivedMsg": string(msg),
			}).Error("echo message mismatch")
		}
	case <-time.After(100 * time.Millisecond):
		suite.logger.Error("Timeout when waiting for an echo message")
	}

	// cleanup
	cmdCh <- WS_QUIT

	<-statusCh // DISCONNECTED
	srv.Close()
}

func (suite *WebSocketSuite) TestStrings() {
	var testState WSState = 12
	assert.Equal(suite.T(), "UNKNOWN STATUS 12", testState.String())

	for _, state := range []WSState{WS_CONNECTING, WS_CONNECTED, WS_DISCONNECTED, WS_WAITING} {
		assert.NotContains(suite.T(), state.String(), "UNKNOWN")
	}

	var testCommand WSCommand = 12
	assert.Equal(suite.T(), "UNKNOWN COMMAND 12", testCommand.String())

	for _, command := range []WSCommand{WS_QUIT} {
		assert.NotContains(suite.T(), command.String(), "UNKNOWN")
	}
}
