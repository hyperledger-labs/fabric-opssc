#!/bin/bash
#
# Copyright 2020-2022 Hitachi, Ltd. Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

docker-compose -f ../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-agents.yaml  -f ../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-agents-org3.yaml -f ../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-agents-org4.yaml down --volumes --remove-orphans
docker-compose -f ../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-api-servers.yaml -f ../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-api-servers-org3.yaml -f ../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-api-servers-org4.yaml down --remove-orphans

cd ../sample-environments/fabric-samples/test-network && ./network.sh down