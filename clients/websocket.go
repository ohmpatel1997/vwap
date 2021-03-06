package clients

// based on https://github.com/aglyzov/ws-machine
import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WSState status of process
type WSState int

// WSCommand command to use protocol
type WSCommand int

type (
	WebSocket struct {
		url            string
		headers        http.Header
		inputCh        chan []byte
		outputCh       chan []byte
		statusCh       chan WSStatus
		cmdCh          chan WSCommand
		readErrorCh    chan error
		writeErrorCh   chan error
		writeControlCh chan WSCommand
		wg             sync.WaitGroup // global wait group to wait for all process to complete
		logger         *log.Logger
	}
	WSStatus struct {
		State WSState
		Error error
	}
)

// states
const (
	WS_DISCONNECTED WSState = iota
	WS_CONNECTING
	WS_CONNECTED
	WS_WAITING
)

// commands
const (
	// WS_QUIT used to quite the process
	WS_QUIT WSCommand = 16 + iota
)

var (
	ErrWSCanceled             = errors.New("cancelled")
	ErrWSOutputChannelClosed  = errors.New("output channel closed")
	ErrWSControlChannelClosed = errors.New("control channel closed")
)

func (s WSState) String() string {
	switch s {
	case WS_DISCONNECTED:
		return "DISCONNECTED"
	case WS_CONNECTING:
		return "CONNECTING"
	case WS_CONNECTED:
		return "CONNECTED"
	case WS_WAITING:
		return "WAITING"
	}
	return fmt.Sprintf("UNKNOWN STATUS %d", s)
}

func (c WSCommand) String() string {
	switch c {
	case WS_QUIT:
		return "QUIT"
	}
	return fmt.Sprintf("UNKNOWN COMMAND %d", c)
}

// NewWebSocket returns the new websocket
func NewWebSocket(logger *log.Logger, url string, headers http.Header) *WebSocket {
	res := &WebSocket{
		url:            url,
		headers:        headers,
		inputCh:        make(chan []byte, 8),
		outputCh:       make(chan []byte, 8),
		statusCh:       make(chan WSStatus, 2),
		cmdCh:          make(chan WSCommand, 2),
		readErrorCh:    make(chan error, 1),
		writeErrorCh:   make(chan error, 1),
		writeControlCh: make(chan WSCommand, 1),
		wg:             sync.WaitGroup{},
		logger:         logger,
	}

	return res
}

// URL ??? get WebSocket URL
func (m *WebSocket) URL() string {
	return m.url
}

// Headers ??? get request headers
func (m *WebSocket) Headers() http.Header {
	return m.headers
}

// Input ??? get input channel
func (m *WebSocket) Input() <-chan []byte {
	return m.inputCh
}

// Output ??? get output channel
func (m *WebSocket) Output() chan<- []byte {
	return m.outputCh
}

// Status ??? get status channel
func (m *WebSocket) Status() <-chan WSStatus {
	return m.statusCh
}

// Command ??? get Command channel
func (m *WebSocket) Command() chan<- WSCommand {
	return m.cmdCh
}

func (m *WebSocket) connect() (*websocket.Conn, error) {
	m.wg.Add(1)
	defer func() {
		m.wg.Done()
	}()

	m.logger.Debug("connect has started")

	m.statusCh <- WSStatus{State: WS_CONNECTING}
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(m.url, m.headers)
	if err == nil {
		m.statusCh <- WSStatus{State: WS_CONNECTED}
		return conn, nil
	}

	m.logger.WithField("error", err).Error("connect error")
	m.statusCh <- WSStatus{State: WS_WAITING}
	m.statusCh <- WSStatus{WS_DISCONNECTED, err}
	return nil, fmt.Errorf("could not able to connect")
}

func (m *WebSocket) read(conn *websocket.Conn) {
	m.wg.Add(1)
	defer func() {
		m.wg.Done()
	}()

	m.logger.Debug("read has started")

	for {
		if _, msg, err := conn.ReadMessage(); err == nil {
			m.logger.WithField("msg", string(msg)).Debug("received message")
			m.inputCh <- msg
		} else {
			m.logger.WithField("error", err).Error("read error")
			m.readErrorCh <- err
			break
		}
	}
}

func (m *WebSocket) write(conn *websocket.Conn) {
	m.wg.Add(1)
	defer func() {
		m.wg.Done()
	}()

	m.logger.Debug("write has started")

	for {
		select {
		case msg, ok := <-m.outputCh:
			if ok {
				if err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
					m.logger.WithFields(log.Fields{
						"msg": string(msg),
					}).Error("set write deadline")
					m.writeErrorCh <- err
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					m.logger.WithFields(log.Fields{
						"msg": string(msg),
					}).Error("writing message")
					m.writeErrorCh <- err
					return
				}
				_ = conn.SetWriteDeadline(time.Time{}) // reset write deadline
			} else {
				m.logger.Error("write error: outCh closed")
				m.writeErrorCh <- ErrWSOutputChannelClosed
				return
			}
		case cmd, ok := <-m.writeControlCh:
			if !ok {
				m.writeErrorCh <- ErrWSControlChannelClosed
				return
			}
			switch cmd {
			case WS_QUIT:
				m.logger.Debug("write received WS_QUIT command")
				m.writeErrorCh <- ErrWSCanceled
				return
			}
		}
	}
}

func (m *WebSocket) cleanup() {
	// close local output channels
	close(m.writeControlCh) // this makes write      to exit

	// drain inputCh channels
	<-time.After(50 * time.Millisecond) // small pause to let things react

drainLoop:
	for {
		select {
		case _, ok := <-m.outputCh:
			if !ok {
				m.outputCh = nil
			}
		case _, ok := <-m.cmdCh:
			if !ok {
				m.inputCh = nil
			}
		case _, ok := <-m.readErrorCh:
			if !ok {
				m.readErrorCh = nil
			}
		case _, ok := <-m.writeErrorCh:
			if !ok {
				m.writeErrorCh = nil
			}
		default:
			break drainLoop
		}
	}

	// wait for all goroutines to stop
	m.wg.Wait()

	// close output channels
	close(m.inputCh)
	close(m.statusCh)
}

// DriverProgram start WebSocket reader, writer, connection establishment
func (m *WebSocket) DriverProgram() {
	var conn *websocket.Conn
	defer func() {
		m.logger.Debug("cleanup has started")
		if conn != nil {
			conn.Close()
		} // this also makes reader to exit

		m.cleanup()
	}()

	conn, err := m.connect()
	if err != nil { // clean it
		m.cmdCh <- WS_QUIT
	}

	m.logger.WithFields(log.Fields{
		"local":  conn.LocalAddr(),
		"remote": conn.RemoteAddr(),
	}).Info("connected")

	reading := true
	writing := true
	go m.read(conn)
	go m.write(conn)

	for {
		select {
		case err := <-m.readErrorCh:
			reading = false
			if writing {
				// write goroutine is still active
				m.logger.Error("read error -> stopping write")
				m.writeControlCh <- WS_QUIT // ask write to exit
				m.statusCh <- WSStatus{WS_DISCONNECTED, err}
			}
		case err := <-m.writeErrorCh:
			// write goroutine has exited
			writing = false
			if reading {
				// read goroutine is still active
				m.logger.Error("write error -> stopping read")
				if conn != nil {
					conn.Close() // this also makes read to exit
					conn = nil
				}
				m.statusCh <- WSStatus{WS_DISCONNECTED, err}
			}
		case cmd, ok := <-m.cmdCh:
			if ok {
				m.logger.WithField("cmd", cmd).Debug("received command")
			}
			switch {
			case !ok || cmd == WS_QUIT:
				if reading || writing || conn != nil {
					m.statusCh <- WSStatus{WS_DISCONNECTED, nil}
				}
				return // defer should clean everything up
			default:
				panic(fmt.Sprintf("unsupported command: %v", cmd))
			}
		}
	}
}
