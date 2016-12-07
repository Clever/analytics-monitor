include golang.mk
.DEFAULT_GOAL := all

SHELL := /bin/bash
PKG := github.com/Clever/analytics-pipeline-monitor
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))

.PHONY: test $(PKGS) run clean vendor

$(eval $(call golang-version-check,1.7))

all: test build

test: $(PKGS)

build:
	go build -o bin/$(EXECUTABLE) $(PKG)

run: build
	./bin/$(EXECUTABLE)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)


$(GOPATH)/bin/glide:
	@go get github.com/Masterminds/glide

install_deps: $(GOPATH)/bin/glide
	@$(GOPATH)/bin/glide install
