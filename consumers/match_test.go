package consumers

import (
	"testing"

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
	suite.useCase.On("Match").Return(suite.matchUseCase)
	suite.consumer = NewMatchConsumer(logger, suite.useCase, suite.config)
}

func (suite *MatchConsumerSuite) TearDownTest() {
}

func (suite *MatchConsumerSuite) Test_Consume_CastError() {
	err := suite.consumer.Consume(nil)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, ErrBadMatchMessage)
}

func (suite *MatchConsumerSuite) Test_Consume_UseCaseError() {
	suite.matchUseCase.On("UpdateVWAP", mock.Anything).Return(repository.ErrTradingPairNotFound)
	err := suite.consumer.Consume(&entity.Match{
		ProductID: "Test",
		Price:     "0.01",
		Size:      "100",
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, repository.ErrTradingPairNotFound)
}

func (suite *MatchConsumerSuite) Test_Consume_Ok() {
	suite.matchUseCase.On("UpdateVWAP", mock.Anything).Return(nil)
	err := suite.consumer.Consume(&entity.Match{
		ProductID: "BTC-USD",
		Price:     "0.01",
		Size:      "100",
	})
	assert.NoError(suite.T(), err)
}
