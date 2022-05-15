package factory

import (
	"container/list"
	"encoding/json"
	"strconv"
	"testing"

	repository2 "github.com/ohmpatel1997/vwap/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ohmpatel1997/vwap/entity"
	"github.com/ohmpatel1997/vwap/pkg/mocks/producers"
	mockRepository "github.com/ohmpatel1997/vwap/pkg/mocks/repository"
)

type MatchUseCaseSuite struct {
	suite.Suite
	producer     *producers.Producer
	vwapProducer *producers.VWAP
	repo         repository2.Repository
	useCase      UseCase
	config       *entity.Config
}

func TestMatchUseCaseSuite(t *testing.T) {
	suite.Run(t, new(MatchUseCaseSuite))
}

func (suite *MatchUseCaseSuite) SetupTest() {
	suite.config = &entity.Config{
		URL:        "https://test.org",
		Capacity:   200,
		Channels:   []string{"best_channel"},
		ProductIDs: []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
	}
	suite.repo = repository2.NewRepository(suite.config)
	suite.vwapProducer = new(producers.VWAP)
	suite.producer = new(producers.Producer)
	suite.producer.On("VWAP").Return(suite.vwapProducer)
	suite.useCase = New(suite.repo, suite.producer, suite.config)
}

func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPBadVolume() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size: "Testing",
	})
	assert.Error(suite.T(), err)
	expectedError := (&strconv.NumError{
		Func: "ParseFloat",
		Num:  "Testing",
		Err:  strconv.ErrSyntax,
	}).Error()
	assert.EqualValues(suite.T(), expectedError, err.Error())
}

func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPBadPrice() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:  "1000",
		Price: "Test",
	})
	assert.Error(suite.T(), err)
	expectedError := (&strconv.NumError{
		Func: "ParseFloat",
		Num:  "Test",
		Err:  strconv.ErrSyntax,
	}).Error()
	assert.EqualValues(suite.T(), expectedError, err.Error())
}

func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPZeroVolume() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:  "0",
		Price: "100.0",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrNegativeOrZeroValue, err)
}

func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPZeroPrice() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:  "100",
		Price: "0.0",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrNegativeOrZeroValue, err)
}

// volume is negative
func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPNegativeVolume() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:  "-1",
		Price: "100.0",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrNegativeOrZeroValue, err)
}

// price is negative
func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPNegativePrice() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:  "100",
		Price: "-1.0",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrNegativeOrZeroValue, err)
}

// error while appending new deal
func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPAppendError() {
	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:      "1000",
		Price:     "0.01",
		ProductID: "Test",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), repository2.ErrTradingPairNotFound, err)
}

// error while producing calculated VWAP
func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPProduceError() {
	_, expectedError := json.Marshal(make(chan int))
	suite.vwapProducer.On("Send", mock.Anything).Return(expectedError)
	suite.producer.On("VWAP").Return(suite.vwapProducer)

	err := suite.useCase.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:      "1000",
		Price:     "0.01",
		ProductID: "BTC-USD",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), expectedError, err)
}

// simulate GetVWAP error
func (suite *MatchUseCaseSuite) TestMatchUseCaseUpdateVWAPGetVWAPError() {
	matchRepo := new(mockRepository.MatchRepository)
	matchRepo.On("Len", mock.Anything).Return(0, nil)
	matchRepo.On("Append", mock.Anything, mock.Anything).Return(nil)
	matchRepo.On("GetVWAP", mock.Anything).Return(nil, repository2.ErrTradingPairNotFound)
	repo := new(mockRepository.Repository)
	repo.On("Match").Return(matchRepo)
	useCaseFactory := New(repo, nil, suite.config)
	err := useCaseFactory.MatchVWAP().UpdateVWAP(&entity.Match{
		Size:      "1000",
		Price:     "0.01",
		ProductID: "BTC-USD",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), repository2.ErrTradingPairNotFound, err)
}

func (suite *MatchUseCaseSuite) calcVWAP(history *list.List) float64 {
	var numerator float64
	var denominator float64

	for e := history.Front(); e != nil; e = e.Next() {
		deal := e.Value.(*entity.Deal)
		numerator += deal.Volume * deal.Price
		denominator += deal.Volume
	}

	return numerator / denominator
}
