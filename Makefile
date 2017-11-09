include golang.mk
.DEFAULT_GOAL := all

SHELL := /bin/bash
PKG := github.com/Clever/analytics-pipeline-monitor
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))

.PHONY: test $(PKGS) run clean vendor

$(eval $(call golang-version-check,1.9))

all: test build

test: $(PKGS)

build:
	go build -o bin/$(EXECUTABLE) $(PKG)
	mkdir -p bin/config
	cp config/latency_config.json bin/config/latency_config.json
	cp kvconfig.yml bin/kvconfig.yml

run: build
	./bin/$(EXECUTABLE)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)





install_deps: golang-dep-vendor-deps
	$(call golang-dep-vendor)