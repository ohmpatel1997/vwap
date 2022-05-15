BINARY_NAME ?= coinbasevwap
FEED_URL    ?= wss://ws-feed.exchange.coinbase.com

.PHONY: all build help lint test run debug

all: build

build:	# Build application
build:
	go build -o ${BINARY_NAME}

test:	# Run tests
test:
	go test -timeout 30s -cover ./...

run:	# Start app
run:
	go run main.go -feed-url ${FEED_URL}

debug:	# Start app with debug logging
debug:
	go run main.go -log-level debug -feed-url ${FEED_URL}
