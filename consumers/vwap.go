package consumers

import (
	"errors"

	"github.com/ohmpatel1997/vwap/usecase"
	log "github.com/sirupsen/logrus"

	"github.com/ohmpatel1997/vwap/entity"
)

//go:generate mockery --name Consumer --case underscore --output ../../pkg/mocks/consumers --outpkg consumers
type Consumer interface {
	Consume(message interface{}) error
}

type vwapConsumer struct {
	config  *entity.Config
	useCase usecase.UseCase
	logger  *log.Logger
}

var (
	ErrBadMatchMessage = errors.New("bad message")
)

func NewVWAPConsumer(logger *log.Logger, useCase usecase.UseCase, config *entity.Config) Consumer {
	return &vwapConsumer{
		config:  config,
		useCase: useCase,
		logger:  logger,
	}
}

func (m *vwapConsumer) Consume(msg interface{}) error {
	message, ok := msg.(*entity.Match)

	if !ok {
		m.logger.WithFields(log.Fields{
			"msg": msg,
		}).Error("matchConsumer.Consume Bad message")

		return ErrBadMatchMessage
	}

	return m.useCase.MatchVWAP().UpdateVWAP(message)
}
