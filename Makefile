# Copyright 2021-2022 Hitachi, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# -------------------------------------------------------------
# This makefile defines the following targets
#   - build-and-tests-all - builds docker images for opssc-agent and opssc-api-server and then runs integration tests for all supported fabric versions
#   - build-and-tests - builds docker images for opssc-agent and opssc-api-server and then runs integration tests for a specific fabric version
#   - docker-all - builds docker images for opssc-agent and opssc-api-server for all supported fabric versions
#   - docker - builds docker images for opssc-agent and opssc-api-server for a specific fabric version
#   - docker-opssc-agent - builds docker images for opssc-agent
#   - docker-opssc-api-server - builds docker images for opssc-api-server
#   - integration-test - runs integration tests for a specific fabric version
#   - check-support-version - checks whether the specified FABRIC_TWO_DIGIT_VERSION is supported or not

BASE_VERSION = 0.2.0
FABRIC_TWO_DIGIT_VERSION ?= 2.4

SUPPORT_FABRIC_TWO_DIGIT_VERSIONS = 2.4 2.2

.PHONY: build-and-tests-all
build-and-tests-all: lint $(SUPPORT_FABRIC_TWO_DIGIT_VERSIONS:%=docker-opssc-agent/%) $(SUPPORT_FABRIC_TWO_DIGIT_VERSIONS:%=docker-opssc-api-server/%) $(SUPPORT_FABRIC_TWO_DIGIT_VERSIONS:%=integration-test/%)

.PHONY: build-and-tests
build-and-tests: lint $(FABRIC_TWO_DIGIT_VERSION:%=docker-opssc-agent/%) $(FABRIC_TWO_DIGIT_VERSION:%=docker-opssc-api-server/%) $(FABRIC_TWO_DIGIT_VERSION:%=integration-test/%)

.PHONY: docker
docker: $(FABRIC_TWO_DIGIT_VERSION:%=docker-opssc-agent/%) $(FABRIC_TWO_DIGIT_VERSION:%=docker-opssc-api-server/%)

.PHONY: docker-all
docker-all: $(SUPPORT_FABRIC_TWO_DIGIT_VERSIONS:%=docker-opssc-agent/%) $(SUPPORT_FABRIC_TWO_DIGIT_VERSIONS:%=docker-opssc-api-server/%)

.PHONY: docker-opssc-agent
docker-opssc-agent: $(FABRIC_TWO_DIGIT_VERSION:%=docker-opssc-agent/%)
docker-opssc-agent/%: $(FABRIC_TWO_DIGIT_VERSION:%=check-support-version)
	@echo "Building docker image for opssc-agent (base version: ${BASE_VERSION}, fabric version: $*)"
	@opssc-agent/scripts/build.sh ${BASE_VERSION} $*

.PHONY: docker-opssc-api-server
docker-opssc-api-server: $(FABRIC_TWO_DIGIT_VERSION:%=docker-opssc-agent/%)
docker-opssc-api-server/%: $(FABRIC_TWO_DIGIT_VERSION:%=check-support-version)
	@echo "Building docker image for opssc-api-server (base version: ${BASE_VERSION}, fabric version: $*)"
	@opssc-api-server/scripts/build.sh ${BASE_VERSION} $*

.PHONY: integration-test
integration-test: $(FABRIC_TWO_DIGIT_VERSION:%=integration-test/%)
integration-test/%: $(FABRIC_TWO_DIGIT_VERSION:%=check-support-version)
	@echo "Executing integration tests (fabric version: $*)"
	@cd integration && FABRIC_TWO_DIGIT_VERSION=$* npm test

.PHONY: lint
lint:
	@cd common/src && npm run lint
	@cd opssc-api-server/src && npm run lint
	@cd opssc-agent/src && npm run lint
	@cd integration && npm run lint

.PHONY: check-support-version
check-support-version:
ifeq ($(findstring $(FABRIC_TWO_DIGIT_VERSION),$(SUPPORT_FABRIC_TWO_DIGIT_VERSIONS)),)
	@echo "Version $(FABRIC_TWO_DIGIT_VERSION) is not supported."
	@exit 1
endif
