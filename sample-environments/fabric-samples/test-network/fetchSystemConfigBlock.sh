#!/bin/bash
#
# Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# set -ex

export FABRIC_CFG_PATH=$PWD/../config/
export PATH=${PWD}/../bin:${PWD}:$PATH
export CORE_PEER_TLS_ENABLED=true

function setOrg1() {
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export CORE_PEER_ADDRESS=localhost:7051
  export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/cert.pem
  export ADMIN_KEY_PATH=${CORE_PEER_MSPCONFIGPATH}/keystore/priv_sk
}

function setOrg1Orderer() {
  export ORDERER_CA=${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer0.org1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_LOCAL_ADDRESS=localhost:7050
  export ORDERER_TLS_HOSTNAME=orderer0.org1.example.com
}

function setOrg2() {
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
  export CORE_PEER_ADDRESS=localhost:9051
  export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/cert.pem
  export ADMIN_KEY_PATH=${CORE_PEER_MSPCONFIGPATH}/keystore/priv_sk
}

function setOrg2Orderer() {
  export ORDERER_CA=${PWD}/organizations/peerOrganizations/org2.example.com/orderers/orderer0.org2.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_LOCAL_ADDRESS=localhost:9050
  export ORDERER_TLS_HOSTNAME=orderer0.org2.example.com
}

function setOrg3() {
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="Org3MSP"
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
  export CORE_PEER_ADDRESS=localhost:11051
  export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/cert.pem
  export ADMIN_KEY_PATH=${CORE_PEER_MSPCONFIGPATH}/keystore/priv_sk
}

function setOrg3Orderer() {
  export ORDERER_CA=${PWD}/organizations/peerOrganizations/org3.example.com/orderers/orderer0.org3.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_LOCAL_ADDRESS=localhost:11050
  export ORDERER_TLS_HOSTNAME=orderer0.org3.example.com
}

function setOrg4() {
  export CORE_PEER_LOCALMSPID="Org4MSP"
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org4.example.com/users/Admin@org4.example.com/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org4.example.com/peers/peer0.org4.example.com/tls/ca.crt
  export CORE_PEER_ADDRESS=localhost:13051
  export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/cert.pem
  export ADMIN_KEY_PATH=${CORE_PEER_MSPCONFIGPATH}/keystore/priv_sk
}

function setOrg4Orderer() {
  export ORDERER_CA=${PWD}/organizations/peerOrganizations/org4.example.com/orderers/orderer0.org4.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_LOCAL_ADDRESS=localhost:13050
  export ORDERER_TLS_HOSTNAME=orderer0.org4.example.com
}


if [ "$1" == "" ]; then
  echo "Usage:"
  echo './fetchSystemConfigBlock.sh $ORG_INDEX'
  exit 1
fi

if [ ! -d "system-genesis-block" ]; then
	mkdir system-genesis-block
fi

setOrg$1
setOrg1Orderer # To fetch system channel config from existing orderers

peer channel fetch config system-genesis-block/updated_genesis.block -o $ORDERER_LOCAL_ADDRESS --ordererTLSHostnameOverride $ORDERER_TLS_HOSTNAME -c system-channel --tls --cafile $ORDERER_CA