package clients

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/ohmpatel1997/vwap/consumers"
	log "github.com/sirupsen/logrus"

	"github.com/ohmpatel1997/vwap/entity"
)

const (
	DefaultCoinbaseRateFeedWebsocketURL = "wss://ws-feed.exchange.coinbase.com"
	DefaultCoinbaseRateFeedChannel      = "matches"
	TypeSubscribe                       = "subscribe"
	TypeUnsubscribe                     = "unsubscribe"
)

var (
	CommandTimeout            = 500 * time.Millisecond
	ErrBadConfiguration       = errors.New("bad configuration")
	ErrFailedToDeserialize    = errors.New("failed to deserialize")
	ErrUnsupportedMessageType = errors.New("skipping unsupported message with unknown type")
)

type CoinbaseRateFeed interface {
	RegisterConsumer(consumer consumers.Consumer)
	Run()
	Stop()
}

type coinbaseRateFeed struct {
	wg           *sync.WaitGroup
	wsm          *WebSocket
	config       *entity.Config
	state        WSState
	subscribers  []consumers.Consumer
	cmdTimeoutCh chan bool
	stopped      bool
	logger       *log.Logger
	mu           sync.RWMutex
}

func NewCoinbaseRateFeed(logger *log.Logger, wg *sync.WaitGroup, config *entity.Config) (CoinbaseRateFeed, error) {
	if config.URL == "" || len(config.Channels) == 0 || len(config.ProductIDs) == 0 {
		return nil, ErrBadConfiguration
	}

	res := &coinbaseRateFeed{
		wg:           wg,
		wsm:          NewWebSocket(logger, config.URL, http.Header{}),
		config:       config,
		state:        WS_CONNECTING,
		subscribers:  make([]consumers.Consumer, 0),
		cmdTimeoutCh: make(chan bool, 2),
		stopped:      false,
		logger:       logger,
	}

	return res, nil
}

func (m *coinbaseRateFeed) stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Stopping client")
	m.stopped = true
}

func (m *coinbaseRateFeed) isStopped() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopped
}

func (m *coinbaseRateFeed) RegisterConsumer(consumer consumers.Consumer) {
	m.subscribers = append(m.subscribers, consumer)
}

func (m *coinbaseRateFeed) publishMatchMessage(packet *entity.Match) {
	m.logger.WithFields(log.Fields{
		"type":   "Match",
		"packet": packet,
	}).Debug("Got Match packet")

	for _, subscriber := range m.subscribers {
		err := subscriber.Consume(packet)
		if err != nil {
			m.logger.WithFields(log.Fields{
				"packet": packet,
				"error":  err,
			}).Error("Match Consumer returned an error")
		}
	}
}

func (m *coinbaseRateFeed) subscriptionTimeout() {
	m.logger.Debug("subscriptionTimeout has started")

	timer := time.NewTimer(CommandTimeout)

	for {
		select {
		case <-m.cmdTimeoutCh: // closed or success
			m.logger.Debug("successfully changed subscription")
			timer.Stop()
			return
		case <-timer.C:
			m.logger.Error("Failed to change subscription")
			cmdCh := m.wsm.Command()
			m.stop()
			cmdCh <- WS_QUIT
			return
		}
	}
}

func (m *coinbaseRateFeed) changeSubscription(output chan<- []byte, command string) {
	msg, _ := json.Marshal(entity.SubscriptionRequest{
		Type:       command,
		Channels:   m.config.Channels,
		ProductIds: m.config.ProductIDs,
	})
	go m.subscriptionTimeout()
	m.logger.WithField("msg", string(msg)).Debug("sending request")
	output <- msg
}

func (m *coinbaseRateFeed) unsubscribe(output chan<- []byte) {
	m.changeSubscription(output, TypeUnsubscribe)
}

func (m *coinbaseRateFeed) subscribe(output chan<- []byte) {
	m.changeSubscription(output, TypeSubscribe)
}

func (m *coinbaseRateFeed) processStatus(output chan<- []byte, status WSStatus) error {
	if status.Error != nil {
		return status.Error
	}
	m.state = status.State
	m.logger.WithField("status", m.state).Debug("changed status")
	if !m.isStopped() { // if not stopped, very first subscribe message
		if status.State == WS_CONNECTED {
			m.subscribe(output)
		}
	}
	return nil
}

func (m *coinbaseRateFeed) processMessage(command chan<- WSCommand, msg []byte) error {
	basePacket := new(entity.Base)
	err := json.Unmarshal(msg, basePacket)
	if err != nil {
		m.logger.WithFields(log.Fields{
			"msg":   string(msg),
			"error": err,
		}).Error("Failed to get type")

		return ErrFailedToDeserialize
	}
	switch basePacket.Type {
	case "subscriptions":
		packet := new(entity.Subscription)
		err = json.Unmarshal(msg, packet)
		if err != nil {
			m.logger.WithFields(log.Fields{
				"msg":   string(msg),
				"type":  basePacket.Type,
				"error": err,
			}).Error("Unable to unmarshal message to subscription structure")
		}

		// this will stop the subscription write
		m.cmdTimeoutCh <- true
		m.logger.WithFields(log.Fields{
			"type":   "Subscription",
			"packet": packet,
		}).Debug("Got Subscription packet")

		if len(packet.Channels) == 0 {
			m.logger.Debug("Successfully unsubscribed. Disconnecting")
			m.stop()
			command <- WS_QUIT
		}
	case "last_match", "match":
		packet := new(entity.Match)
		err = json.Unmarshal(msg, packet)
		if err != nil {
			m.logger.WithFields(log.Fields{
				"msg":   string(msg),
				"type":  basePacket.Type,
				"error": err,
			}).Error("Unable to unmarshal message to match structure")
		}
		m.publishMatchMessage(packet)
	default:
		m.logger.WithFields(log.Fields{
			"msg":  string(msg),
			"type": basePacket.Type,
		}).Error("Skipping message with unknown type")

		return ErrUnsupportedMessageType
	}
	return nil
}

// Run start websocket connection
func (m *coinbaseRateFeed) Run() {
	m.wg.Add(1)

	go m.wsm.DriverProgram()

	command := m.wsm.Command()
	status := m.wsm.Status()
	input := m.wsm.Input()
	output := m.wsm.Output()

	go func() {
	status:
		for {
			select {
			case st, ok := <-status:
				if !ok {
					m.wg.Done()
					break status
				}
				_ = m.processStatus(output, st)
			case msg, ok := <-input:
				if ok {
					_ = m.processMessage(command, msg) // ignore protocol errors
				}
			}
		}
	}()
}

// Stop sends the unsubscribe message to coinbase
func (m *coinbaseRateFeed) Stop() {
	m.unsubscribe(m.wsm.Output())
}
