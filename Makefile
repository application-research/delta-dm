SHELL=/usr/bin/env bash
GO_BUILD_IMAGE?=golang:1.19
VERSION=$(shell git describe --always --tag --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
DOCKER_COMPOSE_FILE=docker-compose.yml

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

.PHONY: docker-compose-build
docker-compose-build:
	docker-compose -f $(DOCKER_COMPOSE_FILE) build

.PHONY: docker-compose-up
docker-compose-up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-push
docker-push:
	docker build -t delta-dm:$(VERSION) .
	docker tag delta-dm:$(VERSION) 0utercore/delta-dm:$(VERSION)
	docker push 0utercore/delta-dm:$(VERSION)