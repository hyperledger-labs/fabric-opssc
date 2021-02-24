#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

export PATH=${PWD}/../bin:${PWD}:$PATH

function one_line_pem {
  echo "`awk 'NF {sub(/\\n/, ""); printf "%s::::",$0;}' $1`"
}

function createYamlCCP {
ORG=$1
CAPEM=$(one_line_pem $2)
ORDERERPEM=$(one_line_pem $3)
PEERPEM=$(one_line_pem $4)
CAPORT=$5
O0PORT=$6
P0PORT=$7
P1PORT=$8

CCP_YAML_TMP=organizations/peerOrganizations/org${ORG}.example.com/connection-org${ORG}-tmp.yaml
CCP_YAML=organizations/peerOrganizations/org${ORG}.example.com/connection-org${ORG}.yaml

cat > $CCP_YAML_TMP <<EOF
---
name: test-network-org${ORG}
version: 1.0.0
client:
  organization: Org${ORG}
  connection:
    timeout:
      peer:
        endorser: '300'
organizations:
  Org${ORG}:
    mspid: Org${ORG}MSP
    peers:
    - peer0.org${ORG}.example.com
EOF

if [ "$P1PORT" != "" ] ; then
cat >> $CCP_YAML_TMP <<EOF
    - peer1.org${ORG}.example.com
EOF
fi

cat >> $CCP_YAML_TMP <<EOF
    certificateAuthorities:
    - ca.org${ORG}.example.com
peers:
  peer0.org${ORG}.example.com:
    url: grpcs://peer0.org${ORG}.example.com:${P0PORT}
    tlsCACerts:
      pem: |
        ${PEERPEM}
    grpcOptions:
      ssl-target-name-override: peer0.org${ORG}.example.com
      hostnameOverride: peer0.org${ORG}.example.com
EOF

if [ "$P1PORT" != "" ] ; then
cat >> $CCP_YAML_TMP <<EOF
  peer1.org${ORG}.example.com:
    url: grpcs://peer1.org${ORG}.example.com:${P1PORT}
    tlsCACerts:
      pem: |
        ${PEERPEM}
    grpcOptions:
      ssl-target-name-override: peer1.org${ORG}.example.com
      hostnameOverride: peer1.org${ORG}.example.com
EOF
fi

cat >> $CCP_YAML_TMP <<EOF
orderers:
  orderer0.org${ORG}.example.com:
    url: grpcs://orderer0.org${ORG}.example.com:${O0PORT}
    tlsCACerts:
      pem: |
        ${ORDERERPEM}
    grpcOptions:
      ssl-target-name-override: orderer0.org${ORG}.example.com
      hostnameOverride: orderer0.org${ORG}.example.com
certificateAuthorities:
  ca.org${ORG}.example.com:
    url: https://ca_org${ORG}:${CAPORT}
    caName: ca-org${ORG}
    tlsCACerts:
      pem: |
        ${CAPEM}
    httpOptions:
      verify: false
EOF

cat $CCP_YAML_TMP | sed 's/::::/\'$'\n        /g' > $CCP_YAML
rm $CCP_YAML_TMP

}

function createOrg {

  ORG=$1
  CA_NAME=$2
  CA_PORT=$3
  PEER_NUM=$4

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

  for (( i=0; i < ${PEER_NUM}; i++ ));
  do
    echo
    echo "Register peer${i}"
    echo
    set -x
    fabric-ca-client register --caname ${CA_NAME} --id.name peer${i} --id.secret peer${i}pw --id.type peer --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
    set +x

    mkdir -p organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com

    echo
    echo "## Generate the peer${i} msp"
    echo
    set -x
    fabric-ca-client enroll -u https://peer${i}:peer${i}pw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/msp --csr.hosts peer${i}.${ORG}.example.com --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
    set +x

    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/msp/config.yaml

    echo
    echo "## Generate the peer${i}-tls certificates"
    echo
    set -x
    fabric-ca-client enroll -u https://peer${i}:peer${i}pw@localhost:${CA_PORT} --caname ${CA_NAME} -M ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls --enrollment.profile tls --csr.hosts peer${i}.${ORG}.example.com --csr.hosts localhost --tls.certfiles ${PWD}/organizations/fabric-ca/${ORG}/tls-cert.pem
    set +x


    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/ca.crt
    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/signcerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/server.crt
    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/keystore/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/server.key

    mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/tlscacerts
    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/msp/tlscacerts/ca.crt

    mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/tlsca
    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/tlsca/tlsca.${ORG}.example.com-cert.pem

    mkdir ${PWD}/organizations/peerOrganizations/${ORG}.example.com/ca
    cp ${PWD}/organizations/peerOrganizations/${ORG}.example.com/peers/peer${i}.${ORG}.example.com/msp/cacerts/* ${PWD}/organizations/peerOrganizations/${ORG}.example.com/ca/ca.${ORG}.example.com-cert.pem
  done

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
  echo './registerEnroll.sh $ORG_INDEX $CA_NAME $CAPORT $O0PORT $P0PORT ($P1PORT)'
  exit 1
fi

ORG_INDEX=$1
CA_NAME=$2
CAPORT=$3
O0PORT=$4
P0PORT=$5
P1PORT=$6
PEERPEM=organizations/peerOrganizations/org${ORG_INDEX}.example.com/tlsca/tlsca.org${ORG_INDEX}.example.com-cert.pem
CAPEM=organizations/peerOrganizations/org${ORG_INDEX}.example.com/ca/ca.org${ORG_INDEX}.example.com-cert.pem
ORDERERPEM=organizations/peerOrganizations/org${ORG_INDEX}.example.com/orderers/orderer0.org${ORG_INDEX}.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

PEER_NUM=1
if [ "$P1PORT" != "" ] ; then
  PEER_NUM=2
fi

createOrg "org${ORG_INDEX}" "${CA_NAME}" "${CAPORT}" "${PEER_NUM}"

createYamlCCP $ORG_INDEX $CAPEM $ORDERERPEM $PEERPEM $CAPORT $O0PORT $P0PORT $P1PORT
