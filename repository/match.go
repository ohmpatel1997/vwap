package repository

import (
	"errors"

	"github.com/ohmpatel1997/vwap/entity"
)

//go:generate mockery --name MatchRepository --case underscore --output ../../pkg/mocks/repository --outpkg repository
type MatchRepository interface {
	Len(tradingPair string) (int, error)
	Append(tradingPair string, deal *entity.Deal) error
	GetVWAP(tradingPair string) (*entity.VWAP, error)
}

type MatchStorage struct {
	Counter        int64
	VolumeSum      float64
	VolumePriceSum float64
	Deals          []*entity.Deal
}

type matchRepository struct {
	config  *entity.Config
	storage map[string]*MatchStorage
}

var (
	ErrTradingPairNotFound = errors.New("trading pair not found")
	ErrDivisionByZero      = errors.New("division by zero")
)

func newMatchRepository(config *entity.Config) MatchRepository {
	storage := make(map[string]*MatchStorage)

	for _, tradingPair := range config.ProductIDs {
		storage[tradingPair] = &MatchStorage{
			Deals: make([]*entity.Deal, 0, config.Capacity),
		}
	}

	return &matchRepository{
		config:  config,
		storage: storage,
	}
}

func (m *matchRepository) getStorage(tradingPair string) (*MatchStorage, error) {
	storage, ok := m.storage[tradingPair]

	if !ok {
		return nil, ErrTradingPairNotFound
	}

	return storage, nil
}

func (m *matchRepository) Len(tradingPair string) (int, error) {
	storage, err := m.getStorage(tradingPair)
	if err != nil {
		return 0, err
	}

	return len(storage.Deals), nil
}

func (m *matchRepository) Append(tradingPair string, deal *entity.Deal) error {
	storage, err := m.getStorage(tradingPair)
	if err != nil {
		return err
	}

	if len(storage.Deals) >= m.config.Capacity { // if buffer full capacity, remove the oldest entry
		storage.VolumeSum -= storage.Deals[storage.Counter].Volume
		storage.VolumePriceSum -= storage.Deals[storage.Counter].Volume * storage.Deals[storage.Counter].Price
		storage.Deals[storage.Counter] = deal
	} else {
		storage.Deals = append(storage.Deals, deal) // else append in back of buffer
	}

	// add the current deal data
	storage.VolumeSum += deal.Volume
	storage.VolumePriceSum += deal.Volume * deal.Price
	storage.Counter = (storage.Counter + 1) % int64(len(storage.Deals))

	return nil
}

func (m *matchRepository) GetVWAP(tradingPair string) (*entity.VWAP, error) {
	storage, err := m.getStorage(tradingPair)
	if err != nil {
		return nil, err
	}

	if storage.VolumeSum == 0 {
		return nil, ErrDivisionByZero
	}

	vwap := entity.VWAP{
		ProductID: tradingPair,
		VWAP:      storage.VolumePriceSum / storage.VolumeSum,
	}
	return &vwap, nil
}
