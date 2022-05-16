package consumers

import (
	"errors"

	"github.com/ohmpatel1997/vwap/factory"
	log "github.com/sirupsen/logrus"

	"github.com/ohmpatel1997/vwap/entity"
)

//go:generate mockery --name Consumer --output ../../pkg/mocks/consumers --outpkg consumers
type Consumer interface {
	Consume(match *entity.Match) error
}

type vwapConsumer struct {
	config  *entity.Config
	useCase factory.UseCase
	logger  *log.Logger
}

var (
	ErrBadMatchMessage = errors.New("bad message")
)

func NewVWAPConsumer(logger *log.Logger, useCase factory.UseCase, config *entity.Config) Consumer {
	return &vwapConsumer{
		config:  config,
		useCase: useCase,
		logger:  logger,
	}
}

func (m *vwapConsumer) Consume(msg *entity.Match) error {
	return m.useCase.MatchVWAP().UpdateVWAP(msg)
}
