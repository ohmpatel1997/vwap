package factory

import (
	"errors"
	"strconv"

	"github.com/ohmpatel1997/vwap/entity"
	"github.com/ohmpatel1997/vwap/producers"
	"github.com/ohmpatel1997/vwap/repository"
)

//go:generate mockery --name MatchUseCase --output ../../pkg/mocks/usecase --outpkg usecase
type MatchUseCase interface {
	UpdateVWAP(match *entity.Match) error
}

type matchUseCase struct {
	repo     repository.Repository
	producer producers.Producer
	config   *entity.Config
}

var ErrNegativeOrZeroValue = errors.New("negative or zero value of volume or price")

func newMatchUseCase(
	repo repository.Repository,
	producer producers.Producer,
	config *entity.Config,
) MatchUseCase {
	return &matchUseCase{
		repo:     repo,
		producer: producer,
		config:   config,
	}
}

func (m *matchUseCase) UpdateVWAP(match *entity.Match) error {
	volume, err := strconv.ParseFloat(match.Size, 32)
	if err != nil {
		return err
	}
	price, err := strconv.ParseFloat(match.Price, 32)
	if err != nil {
		return err
	}
	deal := entity.Deal{
		Volume: volume,
		Price:  price,
	}

	if deal.Volume <= 0 || deal.Price <= 0 {
		return ErrNegativeOrZeroValue
	}

	err = m.repo.Match().Append(match.ProductID, &deal)
	if err != nil {
		return err
	}

	vwap, err := m.repo.Match().GetVWAP(match.ProductID)
	if err != nil {
		return err
	}

	return m.producer.VWAP().Send(vwap)
}
