package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/ohmpatel1997/vwap/clients"
	"github.com/ohmpatel1997/vwap/consumers"
	"github.com/ohmpatel1997/vwap/entity"
	"github.com/ohmpatel1997/vwap/factory"
	"github.com/ohmpatel1997/vwap/producers"
	"github.com/ohmpatel1997/vwap/repository"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultProducts   = "BTC-USD,ETH-USD,ETH-BTC"
	DefaultVolumeSize = 200
)

func main() {
	logger := log.New()
	logger.SetOutput(os.Stderr)
	logger.Info("Live vwap")

	feedURL := flag.String("feed-url", clients.DefaultCoinbaseRateFeedWebsocketURL, "Coinbase feed URL")
	capacity := flag.Int("capacity", DefaultVolumeSize, "Capacity for storing data for VWAP calculation")
	logLevel := flag.String("log-level", "error", "Logging level")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		level = log.ErrorLevel
	}

	logger.SetLevel(level)

	cfg := &entity.Config{
		Channels:   strings.Split(clients.DefaultCoinbaseRateFeedChannel, ","),
		ProductIDs: strings.Split(DefaultProducts, ","),
		URL:        *feedURL,
		Capacity:   *capacity,
	}

	wg := sync.WaitGroup{}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	client, err := clients.NewCoinbaseRateFeed(logger, &wg, cfg)
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to create Coinbase websocket client")
	}

	repo := repository.NewRepository(cfg)
	producer := producers.NewProducer()
	useCase := factory.New(repo, producer, cfg)

	matchConsumer := consumers.NewVWAPConsumer(logger, useCase, cfg)
	client.RegisterMatchConsumer(matchConsumer)

	client.Run()

	go func() {
		for {
			if x := <-interrupt; x != nil {
				logger.Info("interrupt")
				client.Stop()
				return
			}
		}
	}()

	wg.Wait()
	logger.Debug("Finished.")

}
