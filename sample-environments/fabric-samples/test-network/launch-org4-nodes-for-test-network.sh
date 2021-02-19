#!/bin/bash
#
# Copyright 2020-2021 Hitachi America, Ltd.
#
set -e

if [ "$FABRIC_VERSION" == "" ]; then
  echo 'FABRIC_VERSION should be set. (e.g., 2.3.0)'
  exit 1
fi

ORG_INDEX=4
ORG=org${ORG_INDEX}

if [ "$1" == "" ]; then
  # For test, fetch system config block via org1's nodes.
  ./fetchSystemConfigBlock.sh 1
else
  echo ""
  echo "[Save the specified system config block as a local file]"
  CONFIG_BLOCK_BASE64=$1
  echo "${CONFIG_BLOCK_BASE64}" | base64 -d > system-genesis-block/updated_genesis.block
fi

echo ""
echo "[Launch orderers and peers for ${ORG}]"
IMAGE_TAG=${FABRIC_VERSION} docker-compose -f docker/docker-compose-orderer-${ORG}.yaml up -d
IMAGE_TAG=${FABRIC_VERSION} docker-compose -f docker/docker-compose-peer-${ORG}.yaml up -d

echo ""

echo ""
echo "[Launch agents and clients for ${ORG}]"
docker-compose -f docker/docker-compose-opssc-api-servers-${ORG}.yaml up -d
docker-compose -f docker/docker-compose-opssc-agents-${ORG}.yaml up -d
