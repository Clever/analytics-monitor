include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

SHELL := /bin/bash
PKG := github.com/Clever/analytics-pipeline-monitor
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))

.PHONY: test $(PKGS) run clean vendor

$(eval $(call golang-version-check,1.7))

test: $(PKGS)

build:
	go build -o bin/$(EXECUTABLE) $(PKG)

run: build
	gearcmd --name=analytics-pipeline-monitor --cmd=bin/$(EXECUTABLE) --parseargs=false

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))
