# Copyright 2020 Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# -------------------------------------------------------------
# This makefile defines the following targets
#   - setup - updates 'vendor' folder and dependency metadata
#   - build - builds a configtx-cli binary
#   - test - runs the go-test based unit tests
#   - profile - runs unit tests for all packages in coverprofile mode (slow) and then writes out an HTML file displaying annotated source code

GO111MODULES=on
BIN = $(CURDIR)/bin

.PHONY: profile
profile: setup
	@go test ./... -coverprofile coverprofile.out
	@go tool cover -html=coverprofile.out -o coverage.html

.PHONY: test
test: setup
	@echo "Testing..."
	@go test ./...

.PHONY: build
build: setup
	@echo "Building..."
	@mkdir -p $(BIN)
	@go build -o $(BIN)/fabric-configtx-cli


.PHONY: setup
setup:
	@go mod tidy \
		&& go mod vendor
