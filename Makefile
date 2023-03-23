SHELL=/usr/bin/env bash
GO_BUILD_IMAGE?=golang:1.19
VERSION=$(shell git describe --always --tag --dirty)
COMMIT=$(shell git rev-parse --short HEAD)

.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags="-X 'main.Commit=$(COMMIT)' -X main.Version=$(VERSION)"  -o delta-dm

.PHONE: clean
clean:
	rm -f delta-dm

install:
	install -C -m 0755 delta-dm /usr/local/bin