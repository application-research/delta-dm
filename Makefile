SHELL=/usr/bin/env bash
GO_BUILD_IMAGE?=golang:1.19

.PHONY: all
all: build

.PHONY: build
build:
	go build  -o delta-dm

.PHONE: clean
clean:
	rm -f delta-dm

install:
	install -C -m 0755 delta-dm /usr/local/bin