package producers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ohmpatel1997/vwap/entity"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VWAPProducerSuite struct {
	suite.Suite
	producer Producer
}

func TestVWAPProducerSuite(t *testing.T) {
	suite.Run(t, new(VWAPProducerSuite))
}

func (suite *VWAPProducerSuite) SetupTest() {
	suite.producer = NewProducer()
}

func (suite *VWAPProducerSuite) Test_VWAPProducer_Ok() {
	reader, writer, err := os.Pipe()
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create STDOUT mock")
	}
	rescueStdout := os.Stdout
	defer func() {
		os.Stdout = rescueStdout
	}()
	os.Stdout = writer
	err = suite.producer.VWAP().Send(&entity.VWAP{
		ProductID: "BTC-USD",
		VWAP:      1.23,
	})
	assert.NoError(suite.T(), err)
	writer.Close()
	output, _ := ioutil.ReadAll(reader)
	assert.Contains(suite.T(), string(output), `"product_id":"BTC-USD"`)
	assert.Contains(suite.T(), string(output), `"vwap":1.23`)
}
