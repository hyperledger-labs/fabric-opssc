# Copyright 2021 Hitachi, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# -------------------------------------------------------------
# This makefile defines the following targets
#   - docker - builds docker images for opssc-agent and opssc-api-server
#   - docker-opssc-agent - builds docker images for opssc-agent
#   - docker-opssc-api-server - builds docker images for opssc-api-server

BASE_VERSION = 0.2.0
FABRIC_TWO_DIGIT_VERSION ?= 2.3

.PHONY: docker
docker: docker-opssc-agent docker-opssc-api-server

.PHONY: docker-opssc-agent
docker-opssc-agent:
	@echo "Building docker image for opssc-agent (base version: ${BASE_VERSION}, fabric version: ${FABRIC_TWO_DIGIT_VERSION})"
	@opssc-agent/scripts/build.sh ${BASE_VERSION} ${FABRIC_TWO_DIGIT_VERSION}

.PHONY: docker-opssc-api-server
docker-opssc-api-server:
	@echo "Building docker image for opssc-api-server (base version: ${BASE_VERSION}, fabric version: ${FABRIC_TWO_DIGIT_VERSION})"
	@opssc-api-server/scripts/build.sh ${BASE_VERSION} ${FABRIC_TWO_DIGIT_VERSION}
