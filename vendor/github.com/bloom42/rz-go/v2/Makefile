.PHONY: all test release build benchmarks

VERSION := $(shell cat version.go| grep "\sVersion =" | cut -d '"' -f2)

all: test build

test:
	go tool vet -all -shadowstrict .
	go test -v -race ./...

bench:
	go test -v -race -cpu=1,2,4 -bench . -benchmem ./...

build:
	go build ./...

release:
	git tag v$(VERSION)
	git push origin v$(VERSION)

benchmarks:
	cd benchmarks && ./run.sh
