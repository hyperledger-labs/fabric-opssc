#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
## This script is based on https://github.com/hyperledger/fabric-samples/blob/main/test-network-k8s/scripts/rest_sample.sh

set -o errexit
cd "$(dirname "$0")"

# Set an environment variable based on an optional override (TEST_NETWORK_${name})
# from the calling shell.  If the override is not available, assign the parameter
# to a default value.
function context() {
  local name=$1
  local default_value=$2
  local override_name=TEST_NETWORK_${name}

  export ${name}="${!override_name:-${default_value}}"
}

# This magical awk script led to 30 hours of debugging a "TLS handshake error"
# moral: do not edit / alter the number of '\' in the following transform:
function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function json_ccp {
  local ORG=$1
  local PP=$(one_line_pem $2)
  local CP=$(one_line_pem $3)
  local NS=$4
  local OP=$(one_line_pem $5)
  sed -e "s/\${ORG}/$ORG/" \
      -e "s#\${PEERPEM}#$PP#" \
      -e "s#\${CAPEM}#$CP#" \
      -e "s#\${NS}#$NS#" \
      -e "s#\${ORDERERPEM}#$OP#" \
      ccp-template-${SAMPLE_ENV_NAME}.json
}

function construct_rest_sample_configmap() {
  local ns=$ORG1_NS
  echo "Constructing fabric-rest-sample connection profiles"

  ENROLLMENT_DIR=${TEMP_DIR}/enrollments
  CHANNEL_MSP_DIR=${TEMP_DIR}/channel-msp
  CONFIG_DIR=${TEMP_DIR}/fabric-rest-sample-config

  mkdir -p $CONFIG_DIR

  local peer_pem=$CHANNEL_MSP_DIR/peerOrganizations/org1/msp/tlscacerts/tlsca-signcert.pem
  local ca_pem=$CHANNEL_MSP_DIR/peerOrganizations/org1/msp/cacerts/ca-signcert.pem
  local orderer_pem=$CHANNEL_MSP_DIR/ordererOrganizations/org0/msp/tlscacerts/tlsca-signcert.pem
  echo "$(json_ccp 1 $peer_pem $ca_pem $ORG1_NS $orderer_pem)" > ${CONFIG_DIR}/HLF_CONNECTION_PROFILE_ORG1

  peer_pem=$CHANNEL_MSP_DIR/peerOrganizations/org2/msp/tlscacerts/tlsca-signcert.pem
  ca_pem=$CHANNEL_MSP_DIR/peerOrganizations/org2/msp/cacerts/ca-signcert.pem
  echo "$(json_ccp 2 $peer_pem $ca_pem $ORG2_NS $orderer_pem)" > ${CONFIG_DIR}/HLF_CONNECTION_PROFILE_ORG2

  cp $ENROLLMENT_DIR/org1/users/org1admin/msp/signcerts/cert.pem $CONFIG_DIR/HLF_CERTIFICATE_ORG1
  cp $ENROLLMENT_DIR/org2/users/org2admin/msp/signcerts/cert.pem $CONFIG_DIR/HLF_CERTIFICATE_ORG2

  cp $ENROLLMENT_DIR/org1/users/org1admin/msp/keystore/key.pem $CONFIG_DIR/HLF_PRIVATE_KEY_ORG1
  cp $ENROLLMENT_DIR/org2/users/org2admin/msp/keystore/key.pem $CONFIG_DIR/HLF_PRIVATE_KEY_ORG2

  kubectl -n $ns delete configmap fabric-rest-sample-config || true
  kubectl -n $ns create configmap fabric-rest-sample-config --from-file=$CONFIG_DIR

}

context NETWORK_NAME                  test-network
context KUBE_NAMESPACE                ${NETWORK_NAME}
context NS                            ${KUBE_NAMESPACE}
context ORG0_NS                       ${NS}
context ORG1_NS                       ${NS}
context ORG2_NS                       ${NS}
context TEMP_DIR                      ${PWD}/../fabric-samples/test-network-k8s/build
context SAMPLE_ENV_NAME               test-network-k8s

echo "Creating fabric-rest-sample connection profile configmap"
construct_rest_sample_configmap
echo "🏁 - fabric-rest-sample connection profile configmap is ready."