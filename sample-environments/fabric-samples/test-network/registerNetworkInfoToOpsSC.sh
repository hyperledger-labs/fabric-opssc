#!/bin/bash
#
# Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

export FABRIC_CFG_PATH=$PWD/../config/
export PATH=${PWD}/../bin:${PWD}:$PATH
export CORE_PEER_TLS_ENABLED=true

function setOrg1() {
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export CORE_PEER_ADDRESS=localhost:7051
  # export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/Admin@org1.example.com-cert.pem
  export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/cert.pem
  export ADMIN_KEY_PATH=${CORE_PEER_MSPCONFIGPATH}/keystore/priv_sk
}

function setOrg1Orderer() {
  # export ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_CA=${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer0.org1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  # export ORDERER_TLS_HOSTNAME=orderer.example.com
  export ORDERER_LOCAL_ADDRESS=localhost:7050
  export ORDERER_TLS_HOSTNAME=orderer0.org1.example.com
}

function setOrg2() {
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
  export CORE_PEER_ADDRESS=localhost:9051
  # export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/Admin@org2.example.com-cert.pem
  export ADMIN_CERT_PATH=${CORE_PEER_MSPCONFIGPATH}/signcerts/cert.pem
  export ADMIN_KEY_PATH=${CORE_PEER_MSPCONFIGPATH}/keystore/priv_sk
}

function setOrg2Orderer() {
  export ORDERER_CA=${PWD}/organizations/peerOrganizations/org2.example.com/orderers/orderer0.org2.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export ORDERER_LOCAL_ADDRESS=localhost:9050
  export ORDERER_TLS_HOSTNAME=orderer0.org2.example.com
}

function registerChannelToOpsSC() {
  peer chaincode invoke -o $ORDERER_LOCAL_ADDRESS --ordererTLSHostnameOverride $ORDERER_TLS_HOSTNAME --tls --cafile $ORDERER_CA -C $OPS_CHANNEL_NAME -n channel-ops ${PEER_CONN_PARMS} -c '{"function":"CreateChannel","Args":["'$NEW_CHANNEL_NAME'", "'$NEW_CHANNEL_TYPE'", "[]"]}' --waitForEvent
}

function registerOrgToOpsSC() {
  peer chaincode invoke -o $ORDERER_LOCAL_ADDRESS --ordererTLSHostnameOverride $ORDERER_TLS_HOSTNAME --tls --cafile $ORDERER_CA -C $OPS_CHANNEL_NAME -n channel-ops ${PEER_CONN_PARMS} -c '{"function":"AddOrganization","Args":["'$NEW_CHANNEL_NAME'","'${CORE_PEER_LOCALMSPID}'"]}' --waitForEvent
}

function getAllChannels() {
  peer chaincode query -o $ORDERER_LOCAL_ADDRESS --ordererTLSHostnameOverride $ORDERER_TLS_HOSTNAME --tls --cafile $ORDERER_CA -C $OPS_CHANNEL_NAME -n channel-ops -c '{"function":"GetAllChannels","Args":[]}'
}

function parsePeerConnectionParameters() {
  PEER_CONN_PARMS=""
  PEERS=""
  while [ "$#" -gt 0 ]; do
    setOrg$1
    PEER="peer0.org$1"
    ## Set peer adresses
    PEERS="$PEERS $PEER"
    PEER_CONN_PARMS="$PEER_CONN_PARMS --peerAddresses $CORE_PEER_ADDRESS"
    ## Set path to TLS certificate
    TLSINFO=$(eval echo "--tlsRootCertFiles \$CORE_PEER_TLS_ROOTCERT_FILE")
    PEER_CONN_PARMS="$PEER_CONN_PARMS $TLSINFO"
    # shift by one to get to the next organization
    shift
  done
  # remove leading space for output
  PEERS="$(echo -e "$PEERS" | sed -e 's/^[[:space:]]*//')"
}


if [ "$1" == "" ] || [ "$2" == "" ]; then
  echo "Usage:"
  echo './registerNetworkInfoToOpsSC.sh $OPS_CHANNEL_NAME $NEW_CHANNEL_NAME $NEW_CHANNEL_TYPE'
  exit 1
fi

OPS_CHANNEL_NAME=$1
NEW_CHANNEL_NAME=$2
NEW_CHANNEL_TYPE=$3

if [ "$NEW_CHANNEL_TYPE" == "" ] ; then
  NEW_CHANNEL_TYPE="application"
fi

parsePeerConnectionParameters 1 2
setOrg1Orderer

registerChannelToOpsSC

setOrg1
registerOrgToOpsSC

setOrg2
registerOrgToOpsSC

getAllChannels