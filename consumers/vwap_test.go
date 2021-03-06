package consumers

import (
	"testing"

	"github.com/ohmpatel1997/vwap/factory"
	"github.com/ohmpatel1997/vwap/repository"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ohmpatel1997/vwap/entity"
	"github.com/ohmpatel1997/vwap/pkg/mocks/usecase"
)

type MatchConsumerSuite struct {
	suite.Suite
	config       *entity.Config
	matchUseCase *usecase.MatchUseCase
	useCase      *usecase.UseCase
	consumer     Consumer
}

func TestMatchConsumerSuite(t *testing.T) {
	suite.Run(t, new(MatchConsumerSuite))
}

func (suite *MatchConsumerSuite) SetupTest() {
	logger := log.New()
	suite.config = &entity.Config{
		URL:        "https://test.org",
		Capacity:   200,
		Channels:   []string{"best_channel"},
		ProductIDs: []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
	}
	suite.matchUseCase = &usecase.MatchUseCase{}
	suite.useCase = &usecase.UseCase{}
	suite.useCase.On("MatchVWAP").Return(suite.matchUseCase)
	suite.consumer = NewVWAPConsumer(logger, suite.useCase, suite.config)
}

func (suite *MatchConsumerSuite) TestTradingPairNotFound() {
	suite.matchUseCase.On("UpdateVWAP", mock.Anything).Return(repository.ErrTradingPairNotFound)
	err := suite.consumer.Consume(&entity.Match{
		ProductID: "Test",
		Price:     "0.01",
		Size:      "100",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, repository.ErrTradingPairNotFound)
}

func (suite *MatchConsumerSuite) TestSuccessfullyConsume() {
	suite.matchUseCase.On("UpdateVWAP", mock.Anything).Return(nil)
	err := suite.consumer.Consume(&entity.Match{
		ProductID: "BTC-USD",
		Price:     "0.01",
		Size:      "100",
	})
	assert.NoError(suite.T(), err)
}

func (suite *MatchConsumerSuite) TestNegativeOrZeroValue() {
	suite.matchUseCase.On("UpdateVWAP", mock.Anything).Return(factory.ErrNegativeOrZeroValue)
	err := suite.consumer.Consume(&entity.Match{
		ProductID: "BTC-USD",
		Price:     "-0.01",
		Size:      "100",
	})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err, factory.ErrNegativeOrZeroValue, "TestNegativeOrZeroValue error mismatch")
}

func (suite *MatchConsumerSuite) TestInvalidFloatValues() {
	suite.matchUseCase.On("UpdateVWAP", mock.Anything).Return(factory.ErrNegativeOrZeroValue)
	err := suite.consumer.Consume(&entity.Match{
		ProductID: "BTC-USD",
		Price:     "-1.131.343",
		Size:      "100.12.334",
	})
	assert.Error(suite.T(), err)
}
