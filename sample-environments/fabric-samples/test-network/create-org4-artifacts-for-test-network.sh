#!/bin/bash
#
# Copyright 2020-2021 Hitachi America, Ltd.
#
set -e

if [ "$FABRIC_CA_VERSION" == "" ]; then
  echo 'FABRIC_CA_VERSION should be set. (e.g., 1.4.9)'
  exit 1
fi

ORG_INDEX=4
ORG=org${ORG_INDEX}

CA_PORT=13054
PEER1_PORT=13051
ORDERER_PORT=13050

IMAGE_TAG=${FABRIC_CA_VERSION} docker-compose -f docker/docker-compose-ca-${ORG}.yaml up -d
sleep 3

./registerEnroll.sh ${ORG_INDEX} ca-${ORG} ${CA_PORT} ${PEER1_PORT} ${ORDERER_PORT}

echo ""
echo "[CA Certificate for ${ORG}]"

cat organizations/peerOrganizations/${ORG}.example.com/msp/cacerts/localhost-${CA_PORT}-ca-${ORG}.pem

echo ""
echo "[TLS CA Certificate for ${ORG}]"

cat organizations/peerOrganizations/${ORG}.example.com/msp/tlscacerts/ca.crt

echo ""
echo "[Client/Server Certificate for ${ORG} Consenter]"

cat organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/server.crt
