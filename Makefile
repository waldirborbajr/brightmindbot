include .env
include .boturl
export CGO_ENABLED=0
# export

BINARY_NAME = bombot
# GIT_TAG = $(shell git describe --tags --always)
# LDFLAGS = -X 'main.GitTag=$(GIT_TAG)' -w -s

GCFLAGS =
debug: GCFLAGS += -gcflags=all='-l -N'

VERSION ?= $(shell git rev-parse --short HEAD)
LDFLAGS = -ldflags '-s -w -X main.BuildVersion=$(VERSION)'

help: ## 💬 This help message :)
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## 🔨 Build development binaries for Linux
	@go mod tidy
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME) $(LDFLAGS) $(GCFLAGS) -debug-trace=tmp/trace.json main.go

run: ## 󰜎 Build development binaries for Linux
	 go run main.go

air:
	air go run main.go

clean: ## ♻️  Clean up
	@rm -rf bin
	@rm $(GOBIN)/$(BINARY_NAME)

cache: ## ♻️  Clean up
	go clean -modcache
	go clean --cache
	go mod tidy

lint: ## 🔍 Lint & format, will try to fix errors and modify code
	golangci-lint --version
	GOMEMLIMIT=1024MiB golangci-lint run -v --fix --config .golangci.yaml

install: ## Install into GOBIN directory
	@go install ./...

test: ## 📝 Run all tests
	@go test -coverprofile cover.out -v $(shell go list ./... | grep -v /test/)
	@go tool cover -html=cover.out

snap:
	@rm -rf dist/
	@goreleaser release --snapshot

layout: ## 💻 Run Zellij with a layout
	@zellij --layout go-layout.kdl

.PHONY: authors
authors:
	git log --format="%an" | sort | uniq > AUTHORS.txt

gencert:
	openssl genrsa -des3 -passout file:pwd.txt -out rootCA.key 4096
	openssl req -x509 -new -nodes -passin file:pwd.txt -pubout -key rootCA.key -sha256 -days 1024 -out rootCA.crt -subj "/C=BR/ST=Paraná/L=Curitiba/O=Personal/OU=TI/CN=DevOps"
	openssl genrsa -out localhost.key 2048
	openssl req -new -key localhost.key -out localhost.csr -subj "/C=BR/ST=Paraná/L=Curitiba/O=Personal/OU=TI/CN=DevOps"
	openssl x509 -req -in localhost.csr -passin file:pwd.txt -pubout -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out localhost.crt -days 500 -sha256
