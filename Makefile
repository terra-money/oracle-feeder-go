#!/usr/bin/make -f

#################################################
###                   START                   ###
#################################################
start-alliance-oracle-feeder:
	go run ./cmd/feeder/feeder.go alliance-oracle-feeder

start-alliance-rebalance-feeder:
	go run ./cmd/feeder/feeder.go alliance-rebalance-feeder

start-alliance-update-rewards:
	go run ./cmd/feeder/feeder.go alliance-update-rewards

start-alliance-rebalance-emissions:
	go run ./cmd/feeder/feeder.go alliance-rebalance-emissions

start-price-server:
	go run ./cmd/price-server/price_server.go

.PHONY: start-alliance-oracle-feeder start-alliance-rebalance-feeder start-price-server


#################################################
###                  INSTALL                  ###
#################################################
install-feeder:
	go install ./cmd/feeder/feeder.go

install-price-server:
	go install ./cmd/price-server/price_server.go

.PHONY: install-feeder install-price-server


#################################################
###                   BUILD                   ###
#################################################
build-feeder:
	go build -o ./build/ ./cmd/feeder/feeder.go

build-price-server:
	go build -o ./build/ ./cmd/price-server/price_server.go

.PHONY: build-feeder build-price-server


#################################################
###                   LINT                    ###
#################################################
lint:
	golangci-lint run --out-format=tab


format-tools:
	go install mvdan.cc/gofumpt@v0.4.0
	go install github.com/client9/misspell/cmd/misspell@v0.3.4
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.2

format: format-tools
	find . -name '*.go' -type f | xargs gofmt -w -s
	find . -name '*.go' -type f | xargs misspell -w
	find . -name '*.go' -type f | xargs goimports -w -local github.com/cosmos/cosmos-sdk
	

.PHONY: lint format-tools format