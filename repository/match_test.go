package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/ohmpatel1997/vwap/entity"
)

type MatchRepositorySuite struct {
	suite.Suite
	repo Repository
}

func TestMatchRepositorySuite(t *testing.T) {
	suite.Run(t, new(MatchRepositorySuite))
}

func (suite *MatchRepositorySuite) SetupTest() {
	config := &entity.Config{
		URL:        "https://test.org",
		Capacity:   2,
		Channels:   []string{"best_channel"},
		ProductIDs: []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
	}
	suite.repo = NewRepository(config)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryAppendWrongProductID() {
	err := suite.repo.Match().Append("Testing", &entity.Deal{
		Volume: 10,
		Price:  5,
	})
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrTradingPairNotFound, err)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryAppendSuccess() {
	err := suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 1000,
		Price:  0.01,
	})
	assert.NoError(suite.T(), err)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryLenWrongProductID() {
	err := suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 1000,
		Price:  0.01,
	})
	assert.NoError(suite.T(), err)

	_, err = suite.repo.Match().Len("Test")
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrTradingPairNotFound, err)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryLenSuccess() {
	err := suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 1000,
		Price:  0.01,
	})
	assert.NoError(suite.T(), err)
	err = suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 500,
		Price:  0.02,
	})
	assert.NoError(suite.T(), err)

	length, err := suite.repo.Match().Len("BTC-USD")
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), 2, length)

	length, err = suite.repo.Match().Len("ETH-USD")
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), 0, length)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryGetVWAPWrongProductID() {
	_, err := suite.repo.Match().GetVWAP("Test")
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrTradingPairNotFound, err)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryGetVWAPDivisionByZero() {
	err := suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 0,
		Price:  0.01,
	})
	assert.NoError(suite.T(), err)

	_, err = suite.repo.Match().GetVWAP("BTC-USD")
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), ErrDivisionByZero, err)
}

func (suite *MatchRepositorySuite) TestMatchRepositoryGetVWAPSuccess() {
	err := suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 1000,
		Price:  0.01,
	})
	assert.NoError(suite.T(), err)
	err = suite.repo.Match().Append("BTC-USD", &entity.Deal{
		Volume: 500,
		Price:  0.02,
	})
	assert.NoError(suite.T(), err)

	vwap, err := suite.repo.Match().GetVWAP("BTC-USD")
	assert.NoError(suite.T(), err)
	expectedVWAP := &entity.VWAP{
		ProductID: "BTC-USD",
		VWAP:      (1000*0.01 + 500*0.02) / (1000 + 500),
	}
	assert.EqualValues(suite.T(), expectedVWAP, vwap)
}
