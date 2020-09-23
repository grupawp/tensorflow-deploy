.PHONY: all install deps clean test fmt vet
PACKAGES=$(shell go list -mod vendor ./... | grep -v '/vendor/')

export GOFLAGS=-mod=vendor

all: fmt test build install

deps:
	@go get -d -t ${PACKAGES}

build:
	@go build -v -ldflags "-X main.VERSION=`cat ./VERSION`" ${PACKAGES}

test:
	@go test -v ${PACKAGES}

install:
	@go install -v -ldflags "-X main.VERSION=`cat VERSION`" ${PACKAGES}

vet:
	@go tool vet ${PACKAGES}

fmt:
	@go fmt ${PACKAGES}

clean:
	@go clean -i -x
