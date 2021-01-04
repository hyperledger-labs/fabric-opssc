#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

export PATH=${PWD}/../bin:${PWD}:$PATH

function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function json_ccp {
    local PP=$(one_line_pem $5)
    local OP=$(one_line_pem $6)
    local CP=$(one_line_pem $7)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${O0PORT}/$3/" \
        -e "s/\${CAPORT}/$4/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        -e "s#\${ORDERERPEM}#$OP#" \
        organizations/ccp-template.json
}

function yaml_ccp {
    local PP=$(one_line_pem $5)
    local OP=$(one_line_pem $6)
    local CP=$(one_line_pem $7)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${O0PORT}/$3/" \
        -e "s/\${CAPORT}/$4/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        -e "s#\${ORDERERPEM}#$OP#" \
        organizations/ccp-template.yaml | sed -e $'s/\\\\n/\\\n        /g'
}

function createOrg {

  ORG=$1
  CA_NAME=$2
  CA_PORT=$3

  echo
	echo "Enroll the CA admin"
  echo
	mkdir -p organizations/peerOrganizations/${ORG}.example.com/

	export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/${ORG}.example.com/
#  rm -rf $FABRIC_CA_CLIENT_HOME/fabric-ca-client-config.yaml
#  rm -rf $FABRIC_CA_CLIENT_HOME/msp

  set -x
  fabric-ca-client enroll -u https://admin:adminpw@localhost:${CA_PORT} --caname ${CA_NAME} --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  echo "NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/localhost-${CA_PORT}-${CA_NAME}.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/localhost-${CA_PORT}-${CA_NAME}.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/localhost-${CA_PORT}-${CA_NAME}.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/localhost-${CA_PORT}-${CA_NAME}.pem
    OrganizationalUnitIdentifier: orderer" > ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/config.yaml

  echo
	echo "Register peer0"
  echo
  set -x
	fabric-ca-client register --caname ${CA_NAME} --id.name peer0 --id.secret peer0pw --id.type peer --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  echo
  echo "Register user"
  echo
  set -x
  fabric-ca-client register --caname ${CA_NAME} --id.name user1 --id.secret user1pw --id.type client --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  echo
  echo "Register the org admin"
  echo
  set -x
  fabric-ca-client register --caname ${CA_NAME} --id.name ${ORG}admin --id.secret ${ORG}adminpw --id.type admin --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

	mkdir -p organizations/peerOrganizations/${ORG}.example.com/peers
  mkdir -p organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com

  echo
  echo "## Generate the peer0 msp"
  echo
  set -x
	fabric-ca-client enroll -u https://peer0:peer0pw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/msp --csr.hosts peer0.${ORG}.example.com --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/msp/config.yaml

  echo
  echo "## Generate the peer0-tls certificates"
  echo
  set -x
  fabric-ca-client enroll -u https://peer0:peer0pw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls --enrollment.profile tls --csr.hosts peer0.${ORG}.example.com --csr.hosts localhost --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x


  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/ca.crt
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/signcerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/server.crt
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/keystore/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/server.key

  mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/tlscacerts
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/tlscacerts/ca.crt

  mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/tlsca
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/tlsca/tlsca.${ORG}.example.com-cert.pem

  mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/ca
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer0.${ORG}.example.com/msp/cacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/ca/ca.${ORG}.example.com-cert.pem

  mkdir -p organizations/peerOrganizations/${ORG}.example.com/users
  mkdir -p organizations/peerOrganizations/${ORG}.example.com/users/User1@${ORG}.example.com

  echo
	echo "Register orderer"
  echo
  set -x
	fabric-ca-client register --caname ${CA_NAME} --id.name orderer0 --id.secret ordererpw --id.type orderer --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
    set +x

  echo
  echo "## Generate the orderer0 msp"
  echo
  set -x
	fabric-ca-client enroll -u https://orderer0:ordererpw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/msp --csr.hosts orderer0.${ORG}.example.com --csr.hosts localhost --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/msp/config.yaml

  echo
  echo "## Generate the orderer0-tls certificates"
  echo
  set -x
  fabric-ca-client enroll -u https://orderer0:ordererpw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls --enrollment.profile tls --csr.hosts orderer0.${ORG}.example.com --csr.hosts localhost --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/ca.crt
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/signcerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/server.crt
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/keystore/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/server.key

  mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/msp/tlscacerts
  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/orderers/orderer0.${ORG}.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

  echo
  echo "## Generate the user msp"
  echo
  set -x
	fabric-ca-client enroll -u https://user1:user1pw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/users/User1@${ORG}.example.com/msp --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/${ORG}.example.com/users/User1@${ORG}.example.com/msp/config.yaml

  mkdir -p organizations/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com

  echo
  echo "## Generate the org admin msp"
  echo
  set -x
	fabric-ca-client enroll -u https://${ORG}admin:${ORG}adminpw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com/msp --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
  set +x

  cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com/msp/config.yaml
  mv ${PWD}/organizations/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com/msp/keystore/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/users/Admin@${ORG}.example.com/msp/keystore/priv_sk
}


if [ "$1" == "" ] || [ "$2" == "" ] || [ "$3" == "" ] || [ "$4" == "" ] || [ "$5" == "" ] ; then
  echo "Usage:"
  echo './registerEnroll.sh $ORG_INDEX $CA_NAME $CAPORT $P0PORT $O0PORT'
  exit 1
fi

ORG_INDEX=$1
CA_NAME=$2
CAPORT=$3
P0PORT=$4
O0PORT=$5
PEERPEM=organizations/peerOrganizations/org${ORG_INDEX}.example.com/tlsca/tlsca.org${ORG_INDEX}.example.com-cert.pem
CAPEM=organizations/peerOrganizations/org${ORG_INDEX}.example.com/ca/ca.org${ORG_INDEX}.example.com-cert.pem
ORDERERPEM=organizations/peerOrganizations/org${ORG_INDEX}.example.com/orderers/orderer0.org${ORG_INDEX}.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

createOrg "org${ORG_INDEX}" "${CA_NAME}" "${CAPORT}"

echo "$(json_ccp $ORG_INDEX $P0PORT $O0PORT $CAPORT $PEERPEM $ORDERERPEM $CAPEM)" > organizations/peerOrganizations/org${ORG_INDEX}.example.com/connection-org${ORG_INDEX}.json
echo "$(yaml_ccp $ORG_INDEX $P0PORT $O0PORT $CAPORT $PEERPEM $ORDERERPEM $CAPEM)" > organizations/peerOrganizations/org${ORG_INDEX}.example.com/connection-org${ORG_INDEX}.yaml
