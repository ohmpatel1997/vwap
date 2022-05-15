package consumers

import (
	"errors"

	"github.com/ohmpatel1997/vwap/usecase"
	log "github.com/sirupsen/logrus"

	"github.com/ohmpatel1997/vwap/entity"
)

type matchConsumer struct {
	config  *entity.Config
	useCase usecase.UseCase
	logger  *log.Logger
}

var (
	ErrBadMatchMessage = errors.New("bad message")
)

func NewMatchConsumer(logger *log.Logger, useCase usecase.UseCase, config *entity.Config) Consumer {
	return &matchConsumer{
		config:  config,
		useCase: useCase,
		logger:  logger,
	}
}

func (m *matchConsumer) Consume(msg interface{}) error {
	message, ok := msg.(*entity.Match)

	if !ok {
		m.logger.WithFields(log.Fields{
			"msg": msg,
		}).Error("matchConsumer.Consume Bad message")

		return ErrBadMatchMessage
	}

	return m.useCase.Match().UpdateVWAP(message)
}
