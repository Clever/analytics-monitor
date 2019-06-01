include sfncli.mk
include golang.mk
.DEFAULT_GOAL := all

SHELL := /bin/bash
PKG := github.com/Clever/analytics-monitor
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))
SFNCLI_VERSION := latest

.PHONY: test $(PKGS) run clean vendor

$(eval $(call golang-version-check,1.12))

export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432

all: test build

test: $(PKGS)

build: bin/sfncli
	go build -o bin/$(EXECUTABLE) $(PKG)
	mkdir -p bin/config
	cp config/example_config.json bin/config/example_config.json
	cp kvconfig.yml bin/kvconfig.yml

run: build
	./bin/$(EXECUTABLE)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)





install_deps: golang-dep-vendor-deps
	$(call golang-dep-vendor)